//go:build integration

package sas_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/query"
	"github.com/axiomhq/axiom-go/axiom/sas"
	"github.com/axiomhq/axiom-go/internal/config"
	"github.com/axiomhq/axiom-go/internal/test/testhelper"
)

const ingestData = `[
	{
		"time": "17/May/2015:08:05:30 +0000",
		"remote_ip": "93.180.71.1",
		"remote_user": "-",
		"request": "GET /downloads/product_1 HTTP/1.1",
		"response": 304,
		"bytes": 0,
		"referrer": "-",
		"agent": "Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)"
	},
	{
		"time": "17/May/2015:08:05:31 +0000",
		"remote_ip": "93.180.71.2",
		"remote_user": "-",
		"request": "GET /downloads/product_1 HTTP/1.1",
		"response": 304,
		"bytes": 0,
		"referrer": "-",
		"agent": "Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)"
	}
]`

func TestSAS(t *testing.T) {
	cfg := config.Default()
	if err := cfg.IncorporateEnvironment(); err != nil {
		t.Fatal(err)
	} else if err = cfg.Validate(); err != nil {
		t.Fatal(err)
	}

	datasetSuffix := os.Getenv("AXIOM_DATASET_SUFFIX")
	if datasetSuffix == "" {
		datasetSuffix = "local"
	}

	// Clear the environment to avoid unexpected behavior.
	testhelper.SafeClearEnv(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	t.Cleanup(cancel)

	userAgent := fmt.Sprintf("axiom-go-sas-integration-test/%s", datasetSuffix)
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
		Name:        fmt.Sprintf("test-axiom-go-sas-%s", datasetSuffix),
		Description: "This is a test dataset for adapter integration tests.",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		teardownCtx := teardownContext(t, time.Second*15)
		deleteErr := client.Datasets.Delete(teardownCtx, dataset.ID)
		assert.NoError(t, deleteErr)
	})

	// Ingest some test data.
	ingestRes, err := client.Ingest(ctx, dataset.ID, strings.NewReader(ingestData), axiom.JSON, axiom.Identity)
	require.NoError(t, err)
	assert.EqualValues(t, 2, ingestRes.Ingested)

	// List the keys we're going to use for creating the SAS.
	keys, err := client.Organizations.ViewSigningKeys(ctx, cfg.OrganizationID())
	require.NoError(t, err)

	signature, err := sas.Create(keys.Primary, sas.Params{
		OrganizationID: cfg.OrganizationID(),
		Dataset:        dataset.ID,
		Filter:         `remote_ip == "93.180.71.1"`,
		MinStartTime:   "ago(5m)",
		MaxEndTime:     "now",
	})
	require.NoError(t, err)
	require.NotEmpty(t, signature)

	// Now use the SAS for authentication.
	err = client.Options(axiom.SetToken(signature))
	require.NoError(t, err)

	queryRes, err := client.Query(ctx, fmt.Sprintf("['%s'] | count", dataset.ID),
		query.SetStartTime("ago(2m)"),
		query.SetEndTime("ago(1m)"),
	)
	require.NoError(t, err)
	require.NotEmpty(t, queryRes)
}

func teardownContext(t *testing.T, timeout time.Duration) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)
	return ctx
}
