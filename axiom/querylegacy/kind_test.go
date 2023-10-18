package querylegacy

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKind_EncodeValues(t *testing.T) {
	tests := []struct {
		input Kind
		exp   string
	}{
		{emptyKind, ""},
		{Analytics, "analytics"},
		{Stream, "stream"},
		{APL, "apl"},
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

func TestKind_Marshal(t *testing.T) {
	exp := `{
		"kind": "analytics"
	}`

	b, err := json.Marshal(struct {
		Kind Kind `json:"kind"`
	}{
		Kind: Analytics,
	})
	require.NoError(t, err)
	require.NotEmpty(t, b)

	assert.JSONEq(t, exp, string(b))
}

func TestKind_Unmarshal(t *testing.T) {
	var act struct {
		Kind Kind `json:"kind"`
	}
	err := json.Unmarshal([]byte(`{ "kind": "analytics" }`), &act)
	require.NoError(t, err)

	assert.Equal(t, Analytics, act.Kind)
}

func TestKind_String(t *testing.T) {
	// Check outer bounds.
	assert.Empty(t, Kind(0).String())
	assert.Empty(t, emptyKind.String())
	assert.Equal(t, emptyKind, Kind(0))
	assert.Contains(t, (APL + 1).String(), "Kind(")

	for k := Analytics; k <= APL; k++ {
		s := k.String()
		assert.NotEmpty(t, s)
		assert.NotContains(t, s, "Kind(")
	}
}

func TestKindFromString(t *testing.T) {
	for k := Analytics; k <= APL; k++ {
		parsed, err := kindFromString(k.String())
		assert.NoError(t, err)
		assert.Equal(t, k, parsed)
	}
}
