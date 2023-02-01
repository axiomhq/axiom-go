package query

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAggregationOp_Unmarshal(t *testing.T) {
	var act struct {
		Op AggregationOp `json:"name"`
	}
	err := json.Unmarshal([]byte(`{ "name": "count" }`), &act)
	require.NoError(t, err)

	assert.Equal(t, OpCount, act.Op)
}

func TestAggregationOp_String(t *testing.T) {
	// Check outer bounds.
	assert.Equal(t, OpUnknown, AggregationOp(0))
	assert.Contains(t, (OpMakeListIf + 1).String(), "AggregationOp(")

	for op := OpUnknown; op <= OpMakeListIf; op++ {
		s := op.String()
		assert.NotEmpty(t, s)
		assert.NotContains(t, s, "AggregationOp(")
	}
}

func TestAggregationOpFromString(t *testing.T) {
	for op := OpCount; op <= OpMakeListIf; op++ {
		s := op.String()

		parsedOp, err := aggregationOpFromString(s)
		if assert.NoError(t, err) {
			assert.NotEmpty(t, s)
			assert.Equal(t, op, parsedOp)
		}
	}

	op, err := aggregationOpFromString("abc")
	assert.Equal(t, OpUnknown, op)
	assert.EqualError(t, err, "unknown aggregation operation: abc")
}
