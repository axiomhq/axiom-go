package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatus_UnmarshalJSON(t *testing.T) {
	exp := Status{
		ElapsedTime: time.Second,
	}

	var act Status
	err := act.UnmarshalJSON([]byte(`{ "elapsedTime": 1000000 }`))
	require.NoError(t, err)

	assert.Equal(t, exp, act)
}
