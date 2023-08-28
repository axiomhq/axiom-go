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

func TestGzipEncoder(t *testing.T) {
	exp := "Some fox jumps over a fence."

	r, err := GzipEncoder()(strings.NewReader(exp))
	require.NoError(t, err)
	require.NotNil(t, r)

	gzr, err := gzip.NewReader(r)
	require.NoError(t, err)
	defer func() { require.NoError(t, gzr.Close()) }()

	act, err := io.ReadAll(gzr)
	require.NoError(t, err)

	assert.Equal(t, exp, string(act))
}

func TestZstdEncoder(t *testing.T) {
	exp := "Some fox jumps over a fence."

	r, err := ZstdEncoder()(strings.NewReader(exp))
	require.NoError(t, err)

	zsr, err := zstd.NewReader(r)
	require.NoError(t, err)
	defer zsr.Close()

	act, err := io.ReadAll(zsr)
	require.NoError(t, err)

	assert.Equal(t, exp, string(act))
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
			for i := 0; i < b.N; i++ {
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
