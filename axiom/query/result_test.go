package query

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatus(t *testing.T) {
	exp := Status{
		ElapsedTime:    time.Second,
		BlocksExamined: 10,
		RowsExamined:   100000,
		MaxBlockTime:   time.Now().UTC(),
		MinBlockTime:   time.Now().UTC().Add(-time.Hour),
	}

	b, err := json.Marshal(exp)
	require.NoError(t, err)
	require.NotEmpty(t, b)

	var act Status
	err = json.Unmarshal(b, &act)
	require.NoError(t, err)

	assert.Equal(t, exp, act)
}

func TestStatus_MarshalJSON(t *testing.T) {
	exp := `{
		"elapsedTime": 1000000,
		"blocksExamined": 0,
		"rowsExamined": 0,
		"rowsMatched": 0,
		"numGroups": 0,
		"isPartial": false,
		"isEstimate": false,
		"minBlockTime": "0001-01-01T00:00:00Z",
		"maxBlockTime": "0001-01-01T00:00:00Z"
	}`

	act, err := Status{
		ElapsedTime: time.Second,
	}.MarshalJSON()
	require.NoError(t, err)
	require.NotEmpty(t, act)

	assert.JSONEq(t, exp, string(act))
}

func TestStatus_UnmarshalJSON(t *testing.T) {
	exp := Status{
		ElapsedTime: time.Second,
	}

	var act Status
	err := act.UnmarshalJSON([]byte(`{ "elapsedTime": 1000000 }`))
	require.NoError(t, err)

	assert.Equal(t, exp, act)
}
