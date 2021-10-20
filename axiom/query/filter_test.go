package query

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterOp_Marshal(t *testing.T) {
	exp := `{
		"op": "and"
	}`

	b, err := json.Marshal(struct {
		Op FilterOp `json:"op"`
	}{
		Op: OpAnd,
	})
	require.NoError(t, err)
	require.NotEmpty(t, b)

	assert.JSONEq(t, exp, string(b))
}

func TestFilterOp_Unmarshal(t *testing.T) {
	var act struct {
		Op FilterOp `json:"op"`
	}
	err := json.Unmarshal([]byte(`{ "op": "and" }`), &act)
	require.NoError(t, err)

	assert.Equal(t, OpAnd, act.Op)
}

func TestFilterOp_String(t *testing.T) {
	// Check outer bounds.
	assert.Empty(t, FilterOp(0).String(), "")
	assert.Empty(t, UnknownFilterOp.String())
	assert.Contains(t, (OpNotContains + 1).String(), "FilterOp(")

	for c := OpAnd; c <= OpNotContains; c++ {
		s := c.String()
		assert.NotEmpty(t, s)
		assert.NotContains(t, s, "FilterOp(")
	}
}
