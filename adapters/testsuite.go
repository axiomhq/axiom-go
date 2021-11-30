//go:build integration
// +build integration

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
// ingest token for a unique test dataset. The client should be passed to the
// adapter to be tested as well as the target dataset.
type TestFunc func(ctx context.Context, dataset string, client *axiom.Client)

// TestAdapter tests the given adapter with the given test function. It takes
// care of setting up all surroundings for the test.
func TestAdapter(t *testing.T, adapterName string, testFunc TestFunc) {
	t.Helper()

	os.Unsetenv("AXIOM_DATASET")

	if accessToken == "" || !axiom.IsPersonalToken(accessToken) {
		t.Fatal("adapter integration test needs a personal access token set")
	}
	if adapterName == "" {
		t.Fatal("adapter integration test needs the name of the adapter")
	}

	if datasetSuffix = os.Getenv("AXIOM_DATASET_SUFFIX"); datasetSuffix == "" {
		datasetSuffix = "local"
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	t.Cleanup(cancel)

	// Setup the client that will be used to setup the test environment.
	client, err := newClient(
		axiom.SetUserAgent("axiom-go-adapter-integration-test/" + datasetSuffix),
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
		deleteErr := client.Datasets.Delete(teardownContext(t), dataset.ID)
		assert.NoError(t, deleteErr)
	})

	// Create the ingest token for the dataset.
	token, err := client.Tokens.Ingest.Create(ctx, axiom.TokenCreateUpdateRequest{
		Name:        fmt.Sprintf("test-axiom-go-adapter-%s-%s", adapterName, datasetSuffix),
		Description: "This is a test ingest token for adapter integration tests.",
		Scopes:      []string{dataset.ID},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		deleteErr := client.Tokens.Ingest.Delete(teardownContext(t), token.ID)
		assert.NoError(t, deleteErr)
	})

	rawToken, err := client.Tokens.Ingest.View(ctx, token.ID)
	require.NoError(t, err)

	// Create a client that uses the ingest token.
	testClient, err := newClient(
		axiom.SetUserAgent(fmt.Sprintf("axiom-go-adapter-%s-integration-test/%s", adapterName, datasetSuffix)),
		axiom.SetAccessToken(rawToken.Token),
	)
	require.NoError(t, err)

	// Run the test function with the test client.
	testFunc(ctx, dataset.ID, testClient)
}

func teardownContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	t.Cleanup(cancel)
	return ctx
}

func newClient(additionalOptions ...axiom.Option) (*axiom.Client, error) {
	options := []axiom.Option{axiom.SetNoEnv()}

	if deploymentURL != "" {
		options = append(options, axiom.SetURL(deploymentURL))
	}
	if accessToken != "" {
		options = append(options, axiom.SetAccessToken(accessToken))
	}
	if orgID != "" {
		options = append(options, axiom.SetOrgID(orgID))
	}

	options = append(options, additionalOptions...)

	return axiom.NewClient(options...)
}
