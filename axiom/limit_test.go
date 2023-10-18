package axiom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimitScope_String(t *testing.T) {
	// Check outer bounds.
	assert.Equal(t, LimitScopeUnknown, LimitScope(0))
	assert.Contains(t, (LimitScopeAnonymous + 1).String(), "LimitScope(")

	for u := LimitScopeUnknown; u <= LimitScopeAnonymous; u++ {
		s := u.String()
		assert.NotEmpty(t, s)
		assert.NotContains(t, s, "LimitScope(")
	}
}

func TestLimitScopeFromString(t *testing.T) {
	for l := LimitScopeUnknown; l <= LimitScopeAnonymous; l++ {
		parsed, err := limitScopeFromString(l.String())
		assert.NoError(t, err)
		assert.Equal(t, l, parsed)
	}
}
