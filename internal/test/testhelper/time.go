package testhelper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// MustTimeParse parses the given time string using the given layout.
func MustTimeParse(tb testing.TB, layout, value string) time.Time {
	ts, err := time.Parse(layout, value)
	require.NoError(tb, err)
	return ts
}
