package apl

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormat_EncodeValues(t *testing.T) {
	tests := []struct {
		input Format
		exp   string
	}{
		{Legacy, "legacy"},
		// {0, "Format(0)"}, // HINT(lukasmalkmus): Maybe we want to sort this out by raising an error?
	}
	for _, tt := range tests {
		t.Run(tt.input.String(), func(t *testing.T) {
			v := &url.Values{}
			err := tt.input.EncodeValues("test", v)
			require.NoError(t, err)

			assert.Equal(t, tt.exp, v.Get("test"))
		})
	}
}

func TestFormat_String(t *testing.T) {
	// Check outer bounds.
	// assert.Equal(t, Format(0).String(), "Format(0)")
	// assert.Contains(t, (Legacy - 1).String(), "Format(")
	assert.Contains(t, (Legacy + 1).String(), "Format(")

	for c := Legacy; c <= Legacy; c++ {
		s := c.String()
		assert.NotEmpty(t, s)
		assert.NotContains(t, s, "Format(")
	}
}
