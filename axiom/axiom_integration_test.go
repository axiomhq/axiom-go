//go:build integration
// +build integration

package axiom_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/axiom"
)

func TestValidateCredentials(t *testing.T) {
	os.Clearenv()

	os.Setenv("AXIOM_TOKEN", accessToken)
	os.Setenv("AXIOM_ORG_ID", orgID)
	os.Setenv("AXIOM_URL", deploymentURL)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := axiom.ValidateCredentials(ctx)
	require.NoError(t, err)
}
