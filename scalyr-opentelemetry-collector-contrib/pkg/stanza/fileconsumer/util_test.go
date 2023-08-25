// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package fileconsumer

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/observiq/nanojack"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/fileconsumer/emit"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/parser/regex"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/testutil"
)

func nopEmitFunc(_ context.Context, _ []byte, _ map[string]any) error {
	return nil
}

func testEmitFunc(emitChan chan *emitParams) emit.Callback {
	return func(_ context.Context, token []byte, attrs map[string]any) error {
		copied := make([]byte, len(token))
		copy(copied, token)
		emitChan <- &emitParams{attrs, copied}
		return nil
	}
}

// includeDir is a builder-like helper for quickly setting up a test config
func (c *Config) includeDir(dir string) *Config {
	c.Include = append(c.Include, fmt.Sprintf("%s/*", dir))
	return c
}

// withHeader is a builder-like helper for quickly setting up a test config header
func (c *Config) withHeader(headerMatchPattern, extractRegex string) *Config {
	regexOpConfig := regex.NewConfig()
	regexOpConfig.Regex = extractRegex

	c.Header = &HeaderConfig{
		Pattern: headerMatchPattern,
		MetadataOperators: []operator.Config{
			{
				Builder: regexOpConfig,
			},
		},
	}

	return c
}

func emitOnChan(received chan []byte) emit.Callback {
	return func(_ context.Context, token []byte, _ map[string]any) error {
		received <- token
		return nil
	}
}

type emitParams struct {
	attrs map[string]any
	token []byte
}

type testManagerConfig struct {
	emitChan chan *emitParams
}

type testManagerOption func(*testManagerConfig)

func withEmitChan(emitChan chan *emitParams) testManagerOption {
	return func(c *testManagerConfig) {
		c.emitChan = emitChan
	}
}

func buildTestManager(t *testing.T, cfg *Config, opts ...testManagerOption) (*Manager, chan *emitParams) {
	tmc := &testManagerConfig{emitChan: make(chan *emitParams, 100)}
	for _, opt := range opts {
		opt(tmc)
	}
	input, err := cfg.Build(testutil.Logger(t), testEmitFunc(tmc.emitChan))
	require.NoError(t, err)
	return input, tmc.emitChan
}

func openFile(tb testing.TB, path string) *os.File {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	require.NoError(tb, err)
	tb.Cleanup(func() { _ = file.Close() })
	return file
}

func openTemp(t testing.TB, tempDir string) *os.File {
	return openTempWithPattern(t, tempDir, "")
}

func reopenTemp(t testing.TB, name string) *os.File {
	return openTempWithPattern(t, filepath.Dir(name), filepath.Base(name))
}

func openTempWithPattern(t testing.TB, tempDir, pattern string) *os.File {
	file, err := os.CreateTemp(tempDir, pattern)
	require.NoError(t, err)
	t.Cleanup(func() { _ = file.Close() })
	return file
}

func getRotatingLogger(t testing.TB, tempDir string, maxLines, maxBackups int, copyTruncate, sequential bool) *log.Logger {
	file, err := os.CreateTemp(tempDir, "")
	require.NoError(t, err)
	require.NoError(t, file.Close()) // will be managed by rotator

	rotator := nanojack.Logger{
		Filename:     file.Name(),
		MaxLines:     maxLines,
		MaxBackups:   maxBackups,
		CopyTruncate: copyTruncate,
		Sequential:   sequential,
	}

	t.Cleanup(func() { _ = rotator.Close() })

	return log.New(&rotator, "", 0)
}

func writeString(t testing.TB, file *os.File, s string) {
	_, err := file.WriteString(s)
	require.NoError(t, err)
}

func tokenWithLength(length int) []byte {
	charset := "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return b
}

func waitForEmit(t *testing.T, c chan *emitParams) *emitParams {
	select {
	case call := <-c:
		return call
	case <-time.After(3 * time.Second):
		require.FailNow(t, "Timed out waiting for message")
		return nil
	}
}

func waitForNTokens(t *testing.T, c chan *emitParams, n int) [][]byte {
	emitChan := make([][]byte, 0, n)
	for i := 0; i < n; i++ {
		select {
		case call := <-c:
			emitChan = append(emitChan, call.token)
		case <-time.After(3 * time.Second):
			require.FailNow(t, "Timed out waiting for message")
			return nil
		}
	}
	return emitChan
}

func waitForToken(t *testing.T, c chan *emitParams, expected []byte) {
	select {
	case call := <-c:
		require.Equal(t, expected, call.token)
	case <-time.After(3 * time.Second):
		require.FailNow(t, fmt.Sprintf("Timed out waiting for token: %s", expected))
	}
}

func waitForTokenWithAttributes(t *testing.T, c chan *emitParams, expected []byte, attrs map[string]any) {
	select {
	case call := <-c:
		require.Equal(t, expected, call.token)
		require.Equal(t, attrs, call.attrs)
	case <-time.After(3 * time.Second):
		require.FailNow(t, fmt.Sprintf("Timed out waiting for token: %s", expected))
	}
}

func waitForTokens(t *testing.T, c chan *emitParams, expected [][]byte) {
	actual := make([][]byte, 0, len(expected))
LOOP:
	for {
		select {
		case call := <-c:
			actual = append(actual, call.token)
		case <-time.After(3 * time.Second):
			break LOOP
		}
	}

	require.ElementsMatch(t, expected, actual)
}

func expectNoTokens(t *testing.T, c chan *emitParams) {
	expectNoTokensUntil(t, c, 200*time.Millisecond)
}

func expectNoTokensUntil(t *testing.T, c chan *emitParams, d time.Duration) {
	select {
	case token := <-c:
		require.FailNow(t, "Received unexpected message", "Message: %s", token)
	case <-time.After(d):
	}
}

const mockOperatorType = "mock"

func init() {
	operator.Register(mockOperatorType, func() operator.Builder { return newMockOperatorConfig(NewConfig()) })
}

type mockOperatorConfig struct {
	helper.BasicConfig `mapstructure:",squash"`
	*Config            `mapstructure:",squash"`
}

func newMockOperatorConfig(cfg *Config) *mockOperatorConfig {
	return &mockOperatorConfig{
		BasicConfig: helper.NewBasicConfig(mockOperatorType, mockOperatorType),
		Config:      cfg,
	}
}

// This function is impelmented for compatibility with operatortest
// but is not meant to be used directly
func (h *mockOperatorConfig) Build(*zap.SugaredLogger) (operator.Operator, error) {
	panic("not impelemented")
}
