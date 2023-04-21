package testhelper

import (
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/sjson"
)

// JSONEqExp is like assert.JSONEq() but excludes the given fields (given in
// sjson notation: https://github.com/tidwall/sjson).
func JSONEqExp(t assert.TestingT, expected string, actual string, excludedFields []string, msgAndArgs ...any) bool {
	type tHelper interface {
		Helper()
	}

	if h, ok := t.(tHelper); ok {
		h.Helper()
	}

	for _, excludedField := range excludedFields {
		var err error
		if expected, err = sjson.Delete(expected, excludedField); err != nil {
			return assert.Error(t, err)
		} else if actual, err = sjson.Delete(actual, excludedField); err != nil {
			return assert.Error(t, err)
		}
	}

	return assert.JSONEq(t, expected, actual, msgAndArgs...)
}
