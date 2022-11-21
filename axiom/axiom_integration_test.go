//go:build integration

package axiom_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/internal/test/testhelper"
)

func TestValidateCredentials(t *testing.T) {
	testhelper.SafeClearEnv(t)

	t.Setenv("AXIOM_TOKEN", accessToken)
	t.Setenv("AXIOM_ORG_ID", orgID)
	t.Setenv("AXIOM_URL", deploymentURL)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)

	err := axiom.ValidateCredentials(ctx)
	require.NoError(t, err)
}
