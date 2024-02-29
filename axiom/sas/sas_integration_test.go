//go:build integration

package sas_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/ingest"
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

	signingKey := os.Getenv("AXIOM_SIGNING_KEY")
	if signingKey == "" {
		t.Fatal("AXIOM_SIGNING_KEY must be set to a shared access signing key!")
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
		// Restore token.
		optsErr := client.Options(axiom.SetToken(cfg.Token()))
		require.NoError(t, optsErr)

		teardownCtx := teardownContext(t, time.Second*15)
		deleteErr := client.Datasets.Delete(teardownCtx, dataset.ID)
		assert.NoError(t, deleteErr)
	})

	// Ingest some test data.
	ingestRes, err := client.Ingest(ctx, dataset.ID, strings.NewReader(ingestData), axiom.JSON, axiom.Identity)
	require.NoError(t, err)
	require.EqualValues(t, 2, ingestRes.Ingested)

	// Ingest one event each with an earlier timestamp that will break the query
	// test if the signatures time range is not respected.
	now := time.Now()
	then := now.Add(-time.Hour)
	ingestRes, err = client.IngestEvents(ctx, dataset.ID, []axiom.Event{
		{
			ingest.TimestampField: then.Format(time.RFC3339Nano),
			"remote_ip":           "93.180.71.1",
		},
		{
			ingest.TimestampField: then.Format(time.RFC3339Nano),
			"remote_ip":           "93.180.71.2",
		},
	})
	require.NoError(t, err)
	require.EqualValues(t, 2, ingestRes.Ingested)

	options, err := sas.Create(signingKey, sas.Params{
		OrganizationID: cfg.OrganizationID(),
		Dataset:        dataset.ID,
		Filter:         `remote_ip == "93.180.71.1"`,
		MinStartTime:   "ago(5m)",
		MaxEndTime:     "now",
		ExpiryTime:     "endofday(now)",
	})
	require.NoError(t, err)
	require.NotEmpty(t, options)

	u := cfg.BaseURL().JoinPath("/v1/datasets/_apl")
	q := u.Query()
	q.Set("format", "legacy")
	u.RawQuery = q.Encode()

	r := fmt.Sprintf(`{
		"apl": "['%s'] | count",
		"startTime": "ago(1m)",
		"endTime": "now"
	}`, dataset.ID)
	req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(r))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	require.NoError(t, options.Attach(req))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, resp.Body.Close()) })

	require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status code")

	var res query.Result
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))

	assert.EqualValues(t, 1, res.Buckets.Totals[0].Aggregations[0].Value)

	// Now try to query and bypass the timerange via an APL 'where' statement.
	r = fmt.Sprintf(`{
		"apl": "['%s'] | where _time > ago(1d) | count",
		"startTime": "ago(1m)",
		"endTime": "now"
	}`, dataset.ID)
	req, err = http.NewRequest(http.MethodPost, u.String(), strings.NewReader(r))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	require.NoError(t, options.Attach(req))

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, resp.Body.Close()) })

	require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status code")

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))

	assert.EqualValues(t, 1, res.Buckets.Totals[0].Aggregations[0].Value)
}

func teardownContext(t *testing.T, timeout time.Duration) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)
	return ctx
}
