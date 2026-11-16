package query

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAggregationOp_Unmarshal(t *testing.T) {
	tests := []struct {
		input string
		want  AggregationOp
	}{
		{`{ "name": "count" }`, OpCount},
		{`{ "name": "takeany" }`, OpTakeAny},
		{`{ "name": "futureagg" }`, AggregationOp("futureagg")},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var act struct {
				Op AggregationOp `json:"name"`
			}
			require.NoError(t, json.Unmarshal([]byte(tt.input), &act))

			assert.Equal(t, tt.want, act.Op)
		})
	}
}

func TestAggregation_JSONRoundTrip(t *testing.T) {
	exp := Aggregation{Op: OpTopk, Fields: []string{"players"}, Args: []any{float64(10)}}

	b, err := json.Marshal(exp)
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"topk","fields":["players"],"args":[10]}`, string(b))

	var act Aggregation
	require.NoError(t, json.Unmarshal(b, &act))
	assert.Equal(t, exp, act)
}

func TestAggregationOp_String(t *testing.T) {
	assert.Equal(t, "stdev", OpStandardDeviation.String())
}
