package axiom

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/internal/test/testdata"
)

const testEncoderInput = "Some fox jumps over a fence."

func TestGzipEncoder(t *testing.T) {
	exp := testEncoderInput

	r, err := GzipEncoder()(strings.NewReader(exp))
	require.NoError(t, err)
	require.NotNil(t, r)

	gzr, err := gzip.NewReader(r)
	require.NoError(t, err)
	defer func() {
		closeErr := gzr.Close()
		require.NoError(t, closeErr)
	}()

	act, err := io.ReadAll(gzr)
	require.NoError(t, err)

	assert.Equal(t, exp, string(act))
}

func TestZstdEncoder(t *testing.T) {
	exp := testEncoderInput

	r, err := ZstdEncoder()(strings.NewReader(exp))
	require.NoError(t, err)

	zsr, err := zstd.NewReader(r)
	require.NoError(t, err)
	defer zsr.Close()

	act, err := io.ReadAll(zsr)
	require.NoError(t, err)

	assert.Equal(t, exp, string(act))
}

func TestZstdEncoderWithLevel(t *testing.T) {
	exp := testEncoderInput

	levels := []zstd.EncoderLevel{
		zstd.SpeedFastest,
		zstd.SpeedDefault,
		zstd.SpeedBetterCompression,
		zstd.SpeedBestCompression,
	}
	for _, level := range levels {
		t.Run(level.String(), func(t *testing.T) {
			r, err := ZstdEncoderWithLevel(level)(strings.NewReader(exp))
			require.NoError(t, err)

			zsr, err := zstd.NewReader(r)
			require.NoError(t, err)
			defer zsr.Close()

			act, err := io.ReadAll(zsr)
			require.NoError(t, err)

			assert.Equal(t, exp, string(act))
		})
	}

	t.Run("invalid", func(t *testing.T) {
		_, err := ZstdEncoderWithLevel(zstd.EncoderLevel(42))(strings.NewReader(exp))
		assert.ErrorContains(t, err, "unsupported zstd compression level")
	})
}

func BenchmarkZstdEncoder_Allocs(b *testing.B) {
	b.ReportAllocs()
	data := testdata.Load(b)
	for b.Loop() {
		r, err := ZstdEncoder()(bytes.NewReader(data))
		require.NoError(b, err)
		_, err = io.Copy(io.Discard, r)
		require.NoError(b, err)
	}
}

func BenchmarkGzipEncoder_Allocs(b *testing.B) {
	b.ReportAllocs()
	data := testdata.Load(b)
	for b.Loop() {
		r, err := GzipEncoder()(bytes.NewReader(data))
		require.NoError(b, err)
		_, err = io.Copy(io.Discard, r)
		require.NoError(b, err)
	}
}

func BenchmarkZstdEncoder_Parallel(b *testing.B) {
	b.ReportAllocs()
	data := testdata.Load(b)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r, err := ZstdEncoder()(bytes.NewReader(data))
			if err != nil {
				b.Fatal(err)
			}
			if _, err = io.Copy(io.Discard, r); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkGzipEncoder_Parallel(b *testing.B) {
	b.ReportAllocs()
	data := testdata.Load(b)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r, err := GzipEncoder()(bytes.NewReader(data))
			if err != nil {
				b.Fatal(err)
			}
			if _, err = io.Copy(io.Discard, r); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkEncoder(b *testing.B) {
	data := testdata.Load(b)

	benchmarks := []struct {
		name    string
		encoder ContentEncoder
	}{
		{
			name:    "gzip",
			encoder: GzipEncoder(),
		},
		{
			name:    "zstd",
			encoder: ZstdEncoder(),
		},
	}
	for _, bb := range benchmarks {
		b.Run(fmt.Sprintf("encoder=%s", bb.name), func(b *testing.B) {
			for b.Loop() {
				r, err := bb.encoder(bytes.NewReader(data))
				require.NoError(b, err)

				n, err := io.Copy(io.Discard, r)
				require.NoError(b, err)

				b.ReportMetric(float64(n), "size_compressed/op")
				b.ReportMetric(float64(len(data))/float64(n), "compression_ratio/op")
			}
		})
	}
}
