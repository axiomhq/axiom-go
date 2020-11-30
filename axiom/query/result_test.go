package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatus_UnmarshalJSON(t *testing.T) {
	exp := Status{
		ElapsedTime: 542114 * time.Microsecond,
	}

	var act Status
	err := act.UnmarshalJSON([]byte(`{ "elapsedTime": 542114 }`))
	require.NoError(t, err)

	assert.Equal(t, exp, act)
}
