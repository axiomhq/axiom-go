package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatus_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		exp   Status
	}{
		{
			// The server sends the elapsed time in microseconds. This case also
			// covers the result quality flags defaulting to false, which is how
			// the server signals them: it omits "isEstimate" unless it is true.
			name:  "elapsed time",
			input: `{ "elapsedTime": 1000000 }`,
			exp:   Status{ElapsedTime: time.Second},
		},
		{
			name:  "estimate",
			input: `{ "elapsedTime": 1000000, "isEstimate": true }`,
			exp:   Status{ElapsedTime: time.Second, IsEstimate: true},
		},
		{
			name:  "partial",
			input: `{ "elapsedTime": 1000000, "isPartial": true }`,
			exp:   Status{ElapsedTime: time.Second, IsPartial: true},
		},
		{
			name:  "partial and estimate",
			input: `{ "elapsedTime": 1000000, "isPartial": true, "isEstimate": true }`,
			exp:   Status{ElapsedTime: time.Second, IsPartial: true, IsEstimate: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var act Status
			require.NoError(t, act.UnmarshalJSON([]byte(tt.input)))

			assert.Equal(t, tt.exp, act)
		})
	}
}
