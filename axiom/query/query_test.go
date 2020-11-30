package query

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuery(t *testing.T) {
	exp := Query{
		StartTime:  time.Now().UTC(),
		EndTime:    time.Now().UTC().Add(-time.Hour),
		Resolution: time.Second,
		GroupBy:    []string{"hello", "world"},
		Aggregations: []Aggregation{
			{
				Op:    OpAvg,
				Field: "hostname",
			},
		},
		Filter: Filter{
			Op: OpOr,
			Children: []Filter{
				{
					Field: "hostname",
					Op:    OpEqual,
					Value: "foo",
				},
				{
					Field: "hostname",
					Op:    OpEqual,
					Value: "bar",
				},
			},
		},
		Order: []Order{
			{
				Field: "_timestamp",
			},
		},
		VirtualFields: []VirtualField{
			{
				Alias:      "virtA",
				Expression: "status*2",
			},
		},
		Cursor: "c28qdg7oec7w-40-20",
	}

	b, err := json.Marshal(exp)
	require.NoError(t, err)
	require.NotEmpty(t, b)

	var act Query
	err = json.Unmarshal(b, &act)
	require.NoError(t, err)

	assert.Equal(t, exp, act)
}

func TestQuery_MarshalJSON(t *testing.T) {
	tests := []struct {
		input time.Duration
		exp   string
	}{
		{time.Minute + 30*time.Second, "1m30s"},
		{time.Second, "1s"},
		{0, "auto"},
	}
	for _, tt := range tests {
		t.Run(tt.input.String(), func(t *testing.T) {
			q := Query{
				Resolution: tt.input,
			}

			act, err := q.MarshalJSON()
			require.NoError(t, err)
			require.NotEmpty(t, act)

			assert.Contains(t, string(act), tt.exp)
		})
	}
}

func TestQuery_UnarshalJSON(t *testing.T) {
	tests := []struct {
		input string
		exp   time.Duration
	}{
		{"1m30s", time.Minute + 30*time.Second},
		{"1s", time.Second},
		{"auto", 0},
		{"", 0},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			exp := Query{
				Resolution: tt.exp,
			}

			var act Query
			err := act.UnmarshalJSON([]byte(fmt.Sprintf(`{ "resolution": "%s" }`, tt.input)))
			require.NoError(t, err)

			assert.Equal(t, exp, act)
		})
	}
}
