package testdata

import (
	"bytes"
	_ "embed"
	"io"
	"testing"

	"github.com/klauspost/compress/gzip"
	"github.com/stretchr/testify/require"
)

//go:embed large-file.json.gz
var testdata []byte

// Load and decompress the test data from the file.
func Load(tb testing.TB) []byte {
	gzr, err := gzip.NewReader(bytes.NewReader(testdata))
	require.NoError(tb, err)
	defer func() {
		require.NoError(tb, gzr.Close())
	}()

	b, err := io.ReadAll(gzr)
	require.NoError(tb, err)

	return b
}
