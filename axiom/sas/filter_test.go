package sas

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axiomhq/axiom-go/axiom/query"
)

func TestFilter(t *testing.T) {
	exp := query.Filter{
		Op:            query.OpEqual,
		Field:         "customer",
		Value:         "vercel",
		CaseSensitive: true,
		Children: []query.Filter{
			{
				Op:            query.OpEqual,
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

func equalFilter(t *testing.T, exp query.Filter, act filter) {
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
