// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package compress_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awskinesisexporter/internal/compress"
)

func TestCompressorFormats(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		format string
	}{
		{format: "none"},
		{format: "noop"},
		{format: "gzip"},
		{format: "zlib"},
		{format: "flate"},
	}

	const data = "You know nothing Jon Snow"
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("format_%s", tc.format), func(t *testing.T) {
			c, err := compress.NewCompressor(tc.format)
			require.NoError(t, err, "Must have a valid compression format")
			require.NotNil(t, c, "Must have a valid compressor")

			out, err := c.Do([]byte(data))
			assert.NoError(t, err, "Must not error when processing data")
			assert.NotNil(t, out, "Must have a valid record")
		})
	}
	_, err := compress.NewCompressor("invalid-format")
	assert.Error(t, err, "Must error when an invalid compression format is given")
}

func BenchmarkNoopCompressor_1000Bytes(b *testing.B) {
	benchmarkCompressor(b, "none", 1000)
}

func BenchmarkNoopCompressor_1Mb(b *testing.B) {
	benchmarkCompressor(b, "noop", 131072)
}

func BenchmarkZlibCompressor_1000Bytes(b *testing.B) {
	benchmarkCompressor(b, "zlib", 1000)
}

func BenchmarkZlibCompressor_1Mb(b *testing.B) {
	benchmarkCompressor(b, "zlib", 131072)
}

func BenchmarkFlateCompressor_1000Bytes(b *testing.B) {
	benchmarkCompressor(b, "flate", 1000)
}

func BenchmarkFlateCompressor_1Mb(b *testing.B) {
	benchmarkCompressor(b, "flate", 131072)
}

func BenchmarkGzipCompressor_1000Bytes(b *testing.B) {
	benchmarkCompressor(b, "gzip", 1000)
}

func BenchmarkGzipCompressor_1Mb(b *testing.B) {
	benchmarkCompressor(b, "gzip", 131072)
}

func benchmarkCompressor(b *testing.B, format string, length int) {
	b.Helper()

	source := rand.NewSource(time.Now().UnixMilli())
	genRand := rand.New(source)

	compressor, err := compress.NewCompressor(format)
	require.NoError(b, err, "Must not error when given a valid format")
	require.NotNil(b, compressor, "Must have a valid compressor")

	data := make([]byte, length)
	for i := 0; i < length; i++ {
		data[i] = byte(genRand.Int31())
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		out, err := compressor.Do(data)
		assert.NoError(b, err, "Must not error when processing data")
		assert.NotNil(b, out, "Must have a valid byte array after")
	}
}
