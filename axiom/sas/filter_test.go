package sas

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axiomhq/axiom-go/axiom/querylegacy"
)

func TestFilter(t *testing.T) {
	exp := querylegacy.Filter{
		Op:            querylegacy.OpEqual,
		Field:         "customer",
		Value:         "vercel",
		CaseSensitive: true,
		Children: []querylegacy.Filter{
			{
				Op:            querylegacy.OpEqual,
				Field:         "project",
				Value:         "project-123",
				CaseSensitive: false,
			},
		},
	}

	act := filterFromQueryFilter(exp)
	equalFilter(t, exp, act)

	queryAct := act.toQueryFilter()
	assert.Equal(t, exp, queryAct)
}

func equalFilter(t *testing.T, exp querylegacy.Filter, act filter) {
	assert.Equal(t, exp.Op, act.Op)
	assert.Equal(t, exp.Field, act.Field)
	assert.Equal(t, exp.Value, act.Value)
	assert.Equal(t, exp.CaseSensitive, act.CaseSensitive)
	if assert.Equal(t, len(exp.Children), len(act.Children)) {
		for i := range exp.Children {
			equalFilter(t, exp.Children[i], act.Children[i])
		}
	}
}
