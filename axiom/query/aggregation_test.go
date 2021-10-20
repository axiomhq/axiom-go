package query

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAggregationOp_Marshal(t *testing.T) {
	exp := `{
		"op": "count"
	}`

	b, err := json.Marshal(struct {
		Op AggregationOp `json:"op"`
	}{
		Op: OpCount,
	})
	require.NoError(t, err)
	require.NotEmpty(t, b)

	assert.JSONEq(t, exp, string(b))
}

func TestAggregationOp_Unmarshal(t *testing.T) {
	var act struct {
		Op AggregationOp `json:"op"`
	}
	err := json.Unmarshal([]byte(`{ "op": "count" }`), &act)
	require.NoError(t, err)

	assert.Equal(t, OpCount, act.Op)
}

func TestAggregationOp_String(t *testing.T) {
	// Check outer bounds.
	assert.Empty(t, AggregationOp(0).String())
	assert.Empty(t, UnknownAggregationOp.String())
	assert.Contains(t, (OpStandardDeviation + 1).String(), "AggregationOp(")

	for c := OpCount; c <= OpStandardDeviation; c++ {
		s := c.String()
		assert.NotEmpty(t, s)
		assert.NotContains(t, s, "AggregationOp(")
	}
}
