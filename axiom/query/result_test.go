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
		"continuationToken": "",
		"isEstimate": false,
		"minBlockTime": "0001-01-01T00:00:00Z",
		"maxBlockTime": "0001-01-01T00:00:00Z",
		"messages": null
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

func TestMessageCode_Unmarshal(t *testing.T) {
	var act struct {
		MessageCode MessageCode `json:"code"`
	}
	err := json.Unmarshal([]byte(`{ "code": "missing_column" }`), &act)
	require.NoError(t, err)

	assert.Equal(t, MissingColumn, act.MessageCode)
}

func TestMessageCode_String(t *testing.T) {
	// Check outer bounds.
	assert.Equal(t, MessageCode(0).String(), "MessageCode(0)")
	assert.Contains(t, (VirtualFieldFinalizeError - 1).String(), "MessageCode(")
	assert.Contains(t, (DefaultLimitWarning + 1).String(), "MessageCode(")

	for typ := VirtualFieldFinalizeError; typ <= DefaultLimitWarning; typ++ {
		s := typ.String()
		assert.NotEmpty(t, s)
		assert.NotContains(t, s, "MessageCode(")
	}
}

func TestMessagePriority_Unmarshal(t *testing.T) {
	var act struct {
		MessagePriority MessagePriority `json:"priority"`
	}
	err := json.Unmarshal([]byte(`{ "priority": "info" }`), &act)
	require.NoError(t, err)

	assert.Equal(t, Info, act.MessagePriority)
}

func TestMessagePriority_String(t *testing.T) {
	// Check outer bounds.
	assert.Equal(t, MessagePriority(0).String(), "MessagePriority(0)")
	assert.Contains(t, (Trace - 1).String(), "MessagePriority(")
	assert.Contains(t, (Fatal + 1).String(), "MessagePriority(")

	for typ := Trace; typ <= Fatal; typ++ {
		s := typ.String()
		assert.NotEmpty(t, s)
		assert.NotContains(t, s, "MessagePriority(")
	}
}
