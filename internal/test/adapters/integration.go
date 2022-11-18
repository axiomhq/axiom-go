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
	"github.com/axiomhq/axiom-go/axiom/querylegacy"
	"github.com/axiomhq/axiom-go/internal/config"
	"github.com/axiomhq/axiom-go/internal/test/testhelper"
)

var datasetSuffix = os.Getenv("AXIOM_DATASET_SUFFIX")

// IntegrationTestFunc is a function that provides a client that is configured
// with an API token for a unique test dataset. The client should be passed to
// the adapter to be tested as well as the target dataset.
type IntegrationTestFunc func(ctx context.Context, dataset string, client *axiom.Client)

// IntegrationTest tests the given adapter with the given test function. It
// takes care of setting up all surroundings for the integration test.
func IntegrationTest(t *testing.T, adapterName string, testFunc IntegrationTestFunc) {
	cfg := config.Default()
	if err := cfg.IncorporateEnvironment(); err != nil {
		t.Fatal(err)
	} else if err = cfg.Validate(); err != nil {
		t.Fatal(err)
	}

	// Clear the environment to avoid unexpected behavior.
	testhelper.SafeClearEnv(t)

	if adapterName == "" {
		t.Fatal("adapter integration test needs the name of the adapter")
	}

	if datasetSuffix == "" {
		datasetSuffix = "local"
	}

	deadline := time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	t.Cleanup(cancel)

	startTime := time.Now()
	endtime, ok := ctx.Deadline()
	if !ok {
		endtime = startTime.Add(deadline)
	}

	userAgent := fmt.Sprintf("axiom-go-adapter-%s-integration-test/%s", adapterName, datasetSuffix)
	client, err := axiom.NewClient(
		axiom.SetNoEnv(),
		axiom.SetURL(cfg.BaseURL().String()),
		axiom.SetToken(cfg.Token()),
		axiom.SetOrganizationID(cfg.OrganizationID()),
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

	// time.Sleep(time.Second * 30)

	// Make sure the dataset is not empty.
	res, err := client.Datasets.QueryLegacy(ctx, dataset.ID, querylegacy.Query{
		StartTime: startTime,
		EndTime:   endtime,
	}, querylegacy.Options{})
	require.NoError(t, err)

	assert.NotZero(t, len(res.Matches), "dataset should not be empty")
}

func teardownContext(t *testing.T, timeout time.Duration) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)
	return ctx
}
