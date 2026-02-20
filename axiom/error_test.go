package axiom_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/axiom"
)

type (
	is interface{ Is(error) bool }
	as interface{ As(any) bool }
)

var (
	_ error = (*axiom.HTTPError)(nil)
	_ error = (*axiom.LimitError)(nil)

	_ is = (*axiom.HTTPError)(nil)
	_ is = (*axiom.LimitError)(nil)

	_ as = (*axiom.LimitError)(nil)
)

func TestLimitError_As(t *testing.T) {
	httpErr := axiom.HTTPError{
		Status:  http.StatusTooManyRequests,
		Message: "Too Many Requests",
		TraceID: "abc123",
	}
	limitErr := axiom.LimitError{HTTPError: httpErr}

	tests := []struct {
		name string
		err  error
	}{
		{"direct", limitErr},
		{"wrapped", fmt.Errorf("running query: %w", limitErr)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			act, ok := errors.AsType[axiom.HTTPError](tt.err)
			require.True(t, ok)

			// The projection must carry every field, not just the type.
			assert.Equal(t, httpErr, act)
		})
	}

	t.Run("still matches its own type", func(t *testing.T) {
		act, ok := errors.AsType[axiom.LimitError](error(limitErr))
		require.True(t, ok)

		assert.Equal(t, limitErr, act)
	})
}

// The projection must not leak into equality. A limit error is not the same as
// any other error that shares its status code. Adding an Unwrap method to
// [axiom.LimitError] breaks this.
func TestLimitError_Is(t *testing.T) {
	limitErr := axiom.LimitError{
		HTTPError: axiom.HTTPError{Status: http.StatusTooManyRequests},
		Limit:     axiom.Limit{Remaining: 0},
	}

	assert.False(t, errors.Is(limitErr, axiom.HTTPError{Status: http.StatusTooManyRequests}))
	assert.Nil(t, errors.Unwrap(limitErr))

	for _, sentinel := range []error{
		axiom.ErrUnauthorized,
		axiom.ErrUnauthenticated,
		axiom.ErrNotFound,
		axiom.ErrExists,
	} {
		assert.False(t, errors.Is(limitErr, sentinel), "matched %v", sentinel)
	}

	other := limitErr
	other.Limit.Remaining = 99
	assert.False(t, errors.Is(limitErr, other))
}
