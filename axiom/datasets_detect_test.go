package axiom

import (
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectContentTypePreservesInput(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input string
		want  ContentType
	}{
		{"csv spaces", "  message,count\nnext,2\n", CSV},
		{"csv tab", "\tmessage,count\nnext,2\n", CSV},
		{"csv unicode whitespace", "\u0085\u00a0\u2003\u3000message,count\n", CSV},
		{"csv quoted", "\"  message\",count\n", CSV},
		{"csv unicode letter", "\u2003\u00e9v\u00e9nement,count\n", CSV},
		{"csv refill", strings.Repeat(" ", 8192) + "message,count\n", CSV},
		{"csv split rune", strings.Repeat(" ", 4095) + "\u3000message,count\n", CSV},
		{"json", " \t\r\n[{\"message\":\"hello\"}]\n", JSON},
		{"ndjson", " \t\r\n{\"message\":\"hello\"}\n{\"message\":\"next\"}\n", NDJSON},
		{"body bytes", " message,\xff\x00\n", CSV},
	} {
		t.Run(tt.name, func(t *testing.T) {
			for _, reader := range []struct {
				name string
				wrap func(io.Reader) io.Reader
			}{
				{"ordinary", func(r io.Reader) io.Reader { return r }},
				{"one byte", iotest.OneByteReader},
				{"half read", iotest.HalfReader},
				{"data and eof", iotest.DataErrReader},
			} {
				t.Run(reader.name, func(t *testing.T) {
					r, got, err := DetectContentType(reader.wrap(strings.NewReader(tt.input)))
					require.NoError(t, err)
					assert.Equal(t, tt.want, got)
					data, err := io.ReadAll(r)
					require.NoError(t, err)
					assert.Equal(t, tt.input, string(data))
				})
			}
		})
	}
}

func TestDetectContentTypeErrors(t *testing.T) {
	for _, tt := range []struct{ name, input, want string }{
		{"empty", "", "couldn't find beginning of supported ingestion format"},
		{"only whitespace", " \t\r\n\u3000", "couldn't find beginning of supported ingestion format"},
		{"invalid", " \t123", "cannot determine content type"},
		{"invalid utf8", " \xff", "cannot determine content type"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, typ, err := DetectContentType(strings.NewReader(tt.input))
			assert.EqualError(t, err, tt.want)
			assert.Nil(t, r)
			assert.Zero(t, typ)
		})
	}

	readErr := errors.New("read failed")
	for _, prefix := range []string{"", " \u3000"} {
		r, typ, err := DetectContentType(io.MultiReader(strings.NewReader(prefix), iotest.ErrReader(readErr)))
		assert.ErrorIs(t, err, readErr)
		assert.Nil(t, r)
		assert.Zero(t, typ)
	}

	t.Run("error after detection", func(t *testing.T) {
		const input = "  message,count\n"
		r, typ, err := DetectContentType(&contentTypeErrorReader{data: input, err: readErr})
		require.NoError(t, err)
		assert.Equal(t, CSV, typ)
		data, err := io.ReadAll(r)
		assert.ErrorIs(t, err, readErr)
		assert.Equal(t, input, string(data))
	})
}

func TestDetectContentTypeLeavesBodyUnread(t *testing.T) {
	input := "  [" + strings.Repeat(" ", 1<<20) + "]"
	reader := &io.LimitedReader{R: strings.NewReader(input), N: int64(len(input))}
	r, typ, err := DetectContentType(reader)
	require.NoError(t, err)
	assert.Equal(t, JSON, typ)
	assert.GreaterOrEqual(t, reader.N, int64(len(input)-4096))
	data, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.Equal(t, input, string(data))
}

// Return data and its terminal error together to exercise bufio's pending error.
type contentTypeErrorReader struct {
	data string
	err  error
}

func (r *contentTypeErrorReader) Read(p []byte) (int, error) {
	n := copy(p, r.data)
	r.data = r.data[n:]
	if len(r.data) == 0 {
		return n, r.err
	}
	return n, nil
}
