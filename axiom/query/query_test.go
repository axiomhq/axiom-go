package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQuery_MarshalJSON is a primitive test that makes sure the resolution of a
// query is properly marshalled into a string that is "auto" on zero resolution.
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
