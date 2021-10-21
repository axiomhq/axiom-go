//nolint:dupl // Fine to have a bit of duplication in a test file.
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
	assert.Empty(t, FilterOp(0).String())
	assert.Empty(t, emptyFilterOp.String())
	assert.Equal(t, emptyFilterOp, FilterOp(0))
	assert.Contains(t, (OpNotContains + 1).String(), "FilterOp(")

	for op := OpAnd; op <= OpNotContains; op++ {
		s := op.String()
		assert.NotEmpty(t, s)
		assert.NotContains(t, s, "FilterOp(")
	}
}

func TestFilterOpFromString(t *testing.T) {
	for op := OpAnd; op <= OpNotContains; op++ {
		s := op.String()

		parsedOp, err := filterOpFromString(s)
		assert.NoError(t, err)

		assert.NotEmpty(t, s)
		assert.Equal(t, op, parsedOp)
	}
}
