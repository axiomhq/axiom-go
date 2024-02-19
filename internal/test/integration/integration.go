package integration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/internal/config"
)

// Setup marks the calling test as an integration test. Integration tests are
// skipped if not explicitly enabled via AXIOM_INTEGRATION_TESTS. The test fails
// if integration tests are explicitly enabled but if no Axiom environment is
// configured via the environment. Returns a valid configuration for the
// integration test. Should be called early in the test function but must be
// called before [SafeClearEnv].
func Setup(tb testing.TB) config.Config {
	tb.Helper()

	// If not explicitly enabled, skip the test.
	if os.Getenv("AXIOM_INTEGRATION_TESTS") == "" {
		tb.Skip(
			"skipping integration tests;",
			"set AXIOM_INTEGRATION_TESTS=true AXIOM_URL=<URL> AXIOM_TOKEN=<TOKEN> AXIOM_ORG_ID=<ORG_ID> to run this test",
		)
	}

	// Get a default configuration and incorporate environment variables. Fail
	// if the resulting configuration is invalid.
	cfg := config.Default()
	require.NoError(tb, cfg.IncorporateEnvironment())
	require.NoError(tb, cfg.Validate(), "invalid configuration")

	return cfg
}
