package querylegacy

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/internal/test/testhelper"
)

var (
	expStatus = Status{
		ElapsedTime:       time.Second,
		BlocksExamined:    10,
		RowsExamined:      100_000,
		RowsMatched:       2,
		NumGroups:         1,
		IsPartial:         true,
		ContinuationToken: "123",
		IsEstimate:        true,
		MinBlockTime:      parseTimeOrPanic("2022-08-15T10:55:53Z"),
		MaxBlockTime:      parseTimeOrPanic("2022-08-15T11:55:53Z"),
		Messages: []Message{
			{
				Priority: Error,
				Code:     MissingColumn,
				Count:    2,
				Text:     "missing column",
			},
			{
				Priority: Warn,
				Code:     CompilerWarning,
				Count:    1,
				Text:     "some apl compiler warning",
			},
		},
		MinCursor: "c776x1uafkpu-4918f6cb9000095-0",
		MaxCursor: "c776x1uafnvq-4918f6cb9000095-1",
	}

	expStatusJSON = `{
		"elapsedTime": 1000000,
		"blocksExamined": 10,
		"rowsExamined": 100000,
		"rowsMatched": 2,
		"numGroups": 1,
		"isPartial": true,
		"continuationToken": "123",
		"isEstimate": true,
		"minBlockTime": "2022-08-15T10:55:53Z",
		"maxBlockTime": "2022-08-15T11:55:53Z",
		"messages": [
			{
				"priority": "error",
				"code": "missing_column",
				"count": 2,
				"msg": "missing column"
			},
			{
				"priority": "warn",
				"code": "apl_convertingfromtypestotypes_1",
				"count": 1,
				"msg": "some apl compiler warning"
			}
		],
		"minCursor": "c776x1uafkpu-4918f6cb9000095-0",
		"maxCursor": "c776x1uafnvq-4918f6cb9000095-1"
	}`
)

func TestStatus(t *testing.T) {
	b, err := json.Marshal(expStatus)
	require.NoError(t, err)
	require.NotEmpty(t, b)

	var act Status
	err = json.Unmarshal(b, &act)
	require.NoError(t, err)

	assert.Equal(t, expStatus, act)
}

func TestStatus_MarshalJSON(t *testing.T) {
	act, err := expStatus.MarshalJSON()
	require.NoError(t, err)
	require.NotEmpty(t, act)

	testhelper.JSONEqExp(t, expStatusJSON, string(act), []string{"messages.1.code"})
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
	assert.Empty(t, MessageCode(0).String())
	assert.Empty(t, emptyMessageCode.String())
	assert.Equal(t, emptyMessageCode, MessageCode(0))
	assert.Contains(t, (CompilerWarning + 1).String(), "MessageCode(")

	for mc := VirtualFieldFinalizeError; mc <= CompilerWarning; mc++ {
		s := mc.String()
		assert.NotEmpty(t, s)
		assert.NotContains(t, s, "MessageCode(")
	}
}

func TestMessageCodeFromString(t *testing.T) {
	for mc := VirtualFieldFinalizeError; mc <= CompilerWarning; mc++ {
		parsed, err := messageCodeFromString(mc.String())
		assert.NoError(t, err)
		assert.Equal(t, mc, parsed)
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
	assert.Empty(t, MessagePriority(0).String())
	assert.Empty(t, emptyMessagePriority.String())
	assert.Equal(t, emptyMessagePriority, MessagePriority(0))
	assert.Contains(t, (Fatal + 1).String(), "MessagePriority(")

	for mp := Trace; mp <= Fatal; mp++ {
		s := mp.String()
		assert.NotEmpty(t, s)
		assert.NotContains(t, s, "MessagePriority(")
	}
}

func TestMessagePriorityFromString(t *testing.T) {
	for mp := Trace; mp <= Fatal; mp++ {
		parsed, err := messagePriorityFromString(mp.String())
		assert.NoError(t, err)
		assert.Equal(t, mp, parsed)
	}
}

func parseTimeOrPanic(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return t
}
