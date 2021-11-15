package axiom

import (
	"compress/gzip"
	"io"
	"strings"
	"testing"

	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipEncoder(t *testing.T) {
	exp := "Some fox jumps over a fence."

	r, err := GzipEncoder(strings.NewReader(exp))
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
	exp := "Some fox jumps over a fence."

	r, err := ZstdEncoder(strings.NewReader(exp))
	require.NoError(t, err)

	zr, err := zstd.NewReader(r)
	require.NoError(t, err)
	defer zr.Close()

	act, err := io.ReadAll(zr)
	require.NoError(t, err)

	assert.Equal(t, exp, string(act))
}
