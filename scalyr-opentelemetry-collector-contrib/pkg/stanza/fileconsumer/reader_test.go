// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package fileconsumer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/decoder"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/fileconsumer/internal/fingerprint"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/fileconsumer/internal/header"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/fileconsumer/internal/splitter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/parser/regex"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/testutil"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/tokenize"
)

func TestPersistFlusher(t *testing.T) {
	flushPeriod := 100 * time.Millisecond
	sCfg := tokenize.NewSplitterConfig()
	sCfg.Flusher.Period = flushPeriod
	f, emitChan := testReaderFactoryWithSplitter(t, sCfg)

	temp := openTemp(t, t.TempDir())
	fp, err := f.newFingerprint(temp)
	require.NoError(t, err)

	r, err := f.newReader(temp, fp)
	require.NoError(t, err)

	_, err = temp.WriteString("log with newline\nlog without newline")
	require.NoError(t, err)

	// ReadToEnd will return when we hit eof, but we shouldn't emit the unfinished log yet
	r.ReadToEnd(context.Background())
	waitForToken(t, emitChan, []byte("log with newline"))

	// Even trying again shouldn't produce the log yet because the flush period still hasn't expired.
	r.ReadToEnd(context.Background())
	expectNoTokensUntil(t, emitChan, 2*flushPeriod)

	// A copy of the reader should remember that we last emitted about 200ms ago.
	copyReader, err := f.copy(r, temp)
	assert.NoError(t, err)

	// This time, the flusher will kick in and we should emit the unfinished log.
	// If the copy did not remember when we last emitted a log, then the flushPeriod
	// will not be expired at this point so we won't see the unfinished log.
	copyReader.ReadToEnd(context.Background())
	waitForToken(t, emitChan, []byte("log without newline"))
}

func TestTokenization(t *testing.T) {
	testCases := []struct {
		testName    string
		fileContent []byte
		expected    [][]byte
	}{
		{
			"simple",
			[]byte("testlog1\ntestlog2\n"),
			[][]byte{
				[]byte("testlog1"),
				[]byte("testlog2"),
			},
		},
		{
			"empty_only",
			[]byte("\n"),
			[][]byte{
				[]byte(""),
			},
		},
		{
			"empty_first",
			[]byte("\ntestlog1\ntestlog2\n"),
			[][]byte{
				[]byte(""),
				[]byte("testlog1"),
				[]byte("testlog2"),
			},
		},
		{
			"empty_between_lines",
			[]byte("testlog1\n\ntestlog2\n"),
			[][]byte{
				[]byte("testlog1"),
				[]byte(""),
				[]byte("testlog2"),
			},
		},
		{
			"multiple_empty",
			[]byte("\n\ntestlog1\n\n\ntestlog2\n"),
			[][]byte{
				[]byte(""),
				[]byte(""),
				[]byte("testlog1"),
				[]byte(""),
				[]byte(""),
				[]byte("testlog2"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			f, emitChan := testReaderFactory(t)

			temp := openTemp(t, t.TempDir())
			_, err := temp.Write(tc.fileContent)
			require.NoError(t, err)

			fp, err := f.newFingerprint(temp)
			require.NoError(t, err)

			r, err := f.newReader(temp, fp)
			require.NoError(t, err)

			r.ReadToEnd(context.Background())

			for _, expected := range tc.expected {
				require.Equal(t, expected, readToken(t, emitChan))
			}
		})
	}
}

func TestTokenizationTooLong(t *testing.T) {
	fileContent := []byte("aaaaaaaaaaaaaaaaaaaaaa\naaa\n")
	expected := [][]byte{
		[]byte("aaaaaaaaaa"),
		[]byte("aaaaaaaaaa"),
		[]byte("aa"),
		[]byte("aaa"),
	}

	f, emitChan := testReaderFactory(t)
	f.readerConfig.maxLogSize = 10

	temp := openTemp(t, t.TempDir())
	_, err := temp.Write(fileContent)
	require.NoError(t, err)

	fp, err := f.newFingerprint(temp)
	require.NoError(t, err)

	r, err := f.newReader(temp, fp)
	require.NoError(t, err)

	r.ReadToEnd(context.Background())

	for _, expected := range expected {
		require.Equal(t, expected, readToken(t, emitChan))
	}
}

func TestTokenizationTooLongWithLineStartPattern(t *testing.T) {
	fileContent := []byte("aaa2023-01-01aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa 2023-01-01 2 2023-01-01")
	expected := [][]byte{
		[]byte("aaa"),
		[]byte("2023-01-01aaaaa"),
		[]byte("aaaaaaaaaaaaaaa"),
		[]byte("aaaaaaaaaaaaaaa"),
		[]byte("aaaaa"),
		[]byte("2023-01-01 2"),
	}

	f, emitChan := testReaderFactory(t)

	mlc := tokenize.NewMultilineConfig()
	mlc.LineStartPattern = `\d+-\d+-\d+`
	f.splitterFactory = splitter.NewMultilineFactory(tokenize.SplitterConfig{
		Encoding:  "utf-8",
		Flusher:   tokenize.NewFlusherConfig(),
		Multiline: mlc,
	})
	f.readerConfig.maxLogSize = 15

	temp := openTemp(t, t.TempDir())
	_, err := temp.Write(fileContent)
	require.NoError(t, err)

	fp, err := f.newFingerprint(temp)
	require.NoError(t, err)

	r, err := f.newReader(temp, fp)
	require.NoError(t, err)

	r.ReadToEnd(context.Background())
	require.True(t, r.eof)

	for _, expected := range expected {
		require.Equal(t, expected, readToken(t, emitChan))
	}
}

func TestHeaderFingerprintIncluded(t *testing.T) {
	fileContent := []byte("#header-line\naaa\n")

	f, _ := testReaderFactory(t)
	f.readerConfig.maxLogSize = 10

	regexConf := regex.NewConfig()
	regexConf.Regex = "^#(?P<header>.*)"

	enc, err := decoder.LookupEncoding("utf-8")
	require.NoError(t, err)

	h, err := header.NewConfig("^#", []operator.Config{{Builder: regexConf}}, enc)
	require.NoError(t, err)
	f.headerConfig = h

	temp := openTemp(t, t.TempDir())

	fp, err := f.newFingerprint(temp)
	require.NoError(t, err)

	r, err := f.newReader(temp, fp)
	require.NoError(t, err)

	_, err = temp.Write(fileContent)
	require.NoError(t, err)

	r.ReadToEnd(context.Background())

	require.Equal(t, []byte("#header-line\naaa\n"), r.Fingerprint.FirstBytes)
}

func testReaderFactory(t *testing.T) (*readerFactory, chan *emitParams) {
	return testReaderFactoryWithSplitter(t, tokenize.NewSplitterConfig())
}

func testReaderFactoryWithSplitter(t *testing.T, splitterConfig tokenize.SplitterConfig) (*readerFactory, chan *emitParams) {
	emitChan := make(chan *emitParams, 100)
	enc, err := decoder.LookupEncoding(splitterConfig.Encoding)
	require.NoError(t, err)
	return &readerFactory{
		SugaredLogger: testutil.Logger(t),
		readerConfig: &readerConfig{
			fingerprintSize: fingerprint.DefaultSize,
			maxLogSize:      defaultMaxLogSize,
			emit:            testEmitFunc(emitChan),
		},
		fromBeginning:   true,
		splitterFactory: splitter.NewMultilineFactory(splitterConfig),
		encoding:        enc,
	}, emitChan
}

func readToken(t *testing.T, c chan *emitParams) []byte {
	select {
	case call := <-c:
		return call.token
	case <-time.After(3 * time.Second):
		require.FailNow(t, "Timed out waiting for token")
	}
	return nil
}
