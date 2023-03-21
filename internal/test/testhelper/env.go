package testhelper

import (
	"os"
	"strings"
	"testing"
)

// SafeClearEnv clears the environment but restores it when the test finishes.
func SafeClearEnv(tb testing.TB) {
	env := os.Environ()
	os.Clearenv()
	tb.Cleanup(func() {
		os.Clearenv()
		for _, e := range env {
			pair := strings.SplitN(e, "=", 2)
			if err := os.Setenv(pair[0], pair[1]); err != nil {
				tb.Logf("Error setting %q: %v", e, err)
				tb.Fail()
			}
		}
	})
}
