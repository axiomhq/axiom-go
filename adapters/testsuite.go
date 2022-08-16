//go:build integration

package adapters

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/axiom"
)

var (
	accessToken   = os.Getenv("AXIOM_TOKEN")
	orgID         = os.Getenv("AXIOM_ORG_ID")
	deploymentURL = os.Getenv("AXIOM_URL")
	datasetSuffix = os.Getenv("AXIOM_DATASET_SUFFIX")
)

// TestFunc is a function that provides a client that is configured with an
// API token for a unique test dataset. The client should be passed to the
// adapter to be tested as well as the target dataset.
type TestFunc func(ctx context.Context, dataset string, client *axiom.Client)

// TestAdapter tests the given adapter with the given test function. It takes
// care of setting up all surroundings for the test.
func TestAdapter(t *testing.T, adapterName string, testFunc TestFunc) {
	t.Helper()

	// Clear the environment to avoid unexpected behavior.
	SafeClearEnv(t)

	if accessToken == "" || !axiom.IsPersonalToken(accessToken) {
		t.Fatal("adapter integration test needs a personal access token set")
	}
	if deploymentURL == "" {
		t.Fatal("adapter integration test needs the deployment url set")
	}
	if adapterName == "" {
		t.Fatal("adapter integration test needs the name of the adapter")
	}

	if datasetSuffix == "" {
		datasetSuffix = "local"
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	t.Cleanup(cancel)

	userAgent := fmt.Sprintf("axiom-go-adapter-%s-integration-test/%s", adapterName, datasetSuffix)
	client, err := axiom.NewClient(
		axiom.SetNoEnv(),
		axiom.SetURL(deploymentURL),
		axiom.SetAccessToken(accessToken),
		axiom.SetOrgID(orgID),
		axiom.SetUserAgent(userAgent),
	)
	require.NoError(t, err)

	// Get some info on the user that runs the test.
	testUser, err := client.Users.Current(ctx)
	require.NoError(t, err)

	t.Logf("using account %q", testUser.Name)

	// Create the dataset to use.
	dataset, err := client.Datasets.Create(ctx, axiom.DatasetCreateRequest{
		Name:        fmt.Sprintf("test-axiom-go-adapter-%s-%s", adapterName, datasetSuffix),
		Description: "This is a test dataset for adapter integration tests.",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		teardownCtx := teardownContext(t, time.Second*15)
		deleteErr := client.Datasets.Delete(teardownCtx, dataset.ID)
		assert.NoError(t, deleteErr)
	})

	// Run the test function with the test client.
	testFunc(ctx, dataset.ID, client)
}

func teardownContext(t *testing.T, timeout time.Duration) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)
	return ctx
}
