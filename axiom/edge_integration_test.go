package axiom_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/ingest"
	"github.com/axiomhq/axiom-go/axiom/query"
)

var (
	edgeURL           string
	edgeToken         string
	edgeDatasetRegion string
)

func init() {
	edgeURL = os.Getenv("AXIOM_EDGE_URL")
	edgeToken = os.Getenv("AXIOM_EDGE_TOKEN")
	edgeDatasetRegion = os.Getenv("AXIOM_EDGE_DATASET_REGION")
}

// EdgeTestSuite tests ingest and query operations using edge endpoints.
type EdgeTestSuite struct {
	IntegrationTestSuite

	edgeClient *axiom.Client
	dataset    *axiom.Dataset
}

func TestEdgeTestSuite(t *testing.T) {
	suite.Run(t, new(EdgeTestSuite))
}

func (s *EdgeTestSuite) SetupSuite() {
	s.IntegrationTestSuite.SetupSuite()

	// Create edge client with edge configuration
	var edgeOptions []axiom.Option
	if edgeURL != "" {
		s.T().Logf("using edge URL %q", edgeURL)
		edgeOptions = append(edgeOptions, axiom.SetEdgeURL(edgeURL))
	}

	// Use dedicated edge token if provided (edge requires API token, not personal token)
	if edgeToken != "" {
		s.T().Log("using dedicated edge token")
		edgeOptions = append(edgeOptions, axiom.SetToken(edgeToken))
	}

	var err error
	s.edgeClient, err = newClient(edgeOptions...)
	s.Require().NoError(err)
	s.Require().NotNil(s.edgeClient)
}

func (s *EdgeTestSuite) SetupTest() {
	s.IntegrationTestSuite.SetupTest()

	// Create test dataset using the main client (not edge - dataset creation isn't supported on edge)
	req := axiom.DatasetCreateRequest{
		Name:        "test-axiom-go-edge-" + datasetSuffix,
		Description: "This is a test dataset for edge integration tests.",
	}

	// Set dataset region if configured (required for edge routing)
	if edgeDatasetRegion != "" {
		req.Region = edgeDatasetRegion
		s.T().Logf("creating dataset with region %q", edgeDatasetRegion)
	}

	var err error
	s.dataset, err = s.client.Datasets.Create(s.ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(s.dataset)
}

func (s *EdgeTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(s.ctx), time.Second*15)
	defer cancel()

	// Delete dataset using main client (not edge)
	if s.dataset != nil {
		err := s.client.Datasets.Delete(ctx, s.dataset.ID)
		s.NoError(err)
	}

	s.IntegrationTestSuite.TearDownTest()
}

func (s *EdgeTestSuite) TestEdgeIngest() {
	ingestStatus, err := s.edgeClient.Datasets.Ingest(s.ctx, s.dataset.ID, strings.NewReader(ingestData), axiom.JSON, axiom.Identity)
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(2, ingestStatus.Ingested)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)
}

func (s *EdgeTestSuite) TestEdgeIngestEvents() {
	// Test IngestEvents via edge endpoint
	ingestStatus, err := s.edgeClient.Datasets.IngestEvents(s.ctx, s.dataset.ID, ingestEvents)
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(2, ingestStatus.Ingested)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)
}

func (s *EdgeTestSuite) TestEdgeIngestWithLabels() {
	ingestStatus, err := s.edgeClient.Datasets.Ingest(s.ctx, s.dataset.ID, strings.NewReader(ingestData), axiom.JSON, axiom.Identity,
		ingest.SetEventLabel("edge", "true"),
		ingest.SetEventLabel("region", "test"),
	)
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(2, ingestStatus.Ingested)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)
}

func (s *EdgeTestSuite) TestEdgeIngestGzip() {
	r, err := axiom.GzipEncoder()(strings.NewReader(ingestData))
	s.Require().NoError(err)

	ingestStatus, err := s.edgeClient.Datasets.Ingest(s.ctx, s.dataset.ID, r, axiom.JSON, axiom.Gzip)
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(2, ingestStatus.Ingested)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)
}

func (s *EdgeTestSuite) TestEdgeIngestZstd() {
	r, err := axiom.ZstdEncoder()(strings.NewReader(ingestData))
	s.Require().NoError(err)

	ingestStatus, err := s.edgeClient.Datasets.Ingest(s.ctx, s.dataset.ID, r, axiom.JSON, axiom.Zstd)
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(2, ingestStatus.Ingested)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)
}

func (s *EdgeTestSuite) TestEdgeQuery() {
	// First, ingest some data via edge
	ingestStatus, err := s.edgeClient.Datasets.IngestEvents(s.ctx, s.dataset.ID, ingestEvents)
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)
	s.EqualValues(2, ingestStatus.Ingested)

	// Wait a moment for data to be queryable
	time.Sleep(time.Second * 2)

	now := time.Now().Truncate(time.Second)
	startTime := now.Add(-time.Minute)
	endTime := now.Add(time.Minute)

	// Query via edge endpoint
	apl := fmt.Sprintf("['%s']", s.dataset.ID)
	queryResult, err := s.edgeClient.Datasets.Query(s.ctx, apl,
		query.SetStartTime(startTime),
		query.SetEndTime(endTime),
	)
	s.Require().NoError(err)
	s.Require().NotNil(queryResult)

	s.NotZero(queryResult.Status.ElapsedTime)
	s.GreaterOrEqual(queryResult.Status.RowsExamined, uint64(2))
	s.GreaterOrEqual(queryResult.Status.RowsMatched, uint64(2))
}

func (s *EdgeTestSuite) TestEdgeIngestAndQueryRoundTrip() {
	// Full round-trip test: ingest via edge, query via edge

	// Ingest unique test events
	testEvents := []axiom.Event{
		{
			ingest.TimestampField: time.Now(),
			"test_id":             "edge-roundtrip-1",
			"value":               42,
		},
		{
			ingest.TimestampField: time.Now(),
			"test_id":             "edge-roundtrip-2",
			"value":               84,
		},
	}

	ingestStatus, err := s.edgeClient.Datasets.IngestEvents(s.ctx, s.dataset.ID, testEvents)
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)
	s.EqualValues(2, ingestStatus.Ingested)

	// Wait for data to be queryable
	time.Sleep(time.Second * 2)

	now := time.Now().Truncate(time.Second)
	startTime := now.Add(-time.Minute)
	endTime := now.Add(time.Minute)

	// Query for specific test data via edge
	apl := fmt.Sprintf("['%s'] | where test_id startswith 'edge-roundtrip'", s.dataset.ID)
	queryResult, err := s.edgeClient.Datasets.Query(s.ctx, apl,
		query.SetStartTime(startTime),
		query.SetEndTime(endTime),
	)
	s.Require().NoError(err)
	s.Require().NotNil(queryResult)

	s.EqualValues(2, queryResult.Status.RowsMatched)
}

func (s *EdgeTestSuite) TestEdgeIngestChannel() {
	// Test IngestChannel via edge endpoint
	eventCh := make(chan axiom.Event, 2)
	go func() {
		for _, e := range ingestEvents {
			eventCh <- e
		}
		close(eventCh)
	}()

	ingestStatus, err := s.edgeClient.Datasets.IngestChannel(s.ctx, s.dataset.ID, eventCh)
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(2, ingestStatus.Ingested)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)
}

// TestMainClientStillWorks verifies that non-edge operations still work via the
// main client when edge configuration is present.
func (s *EdgeTestSuite) TestMainClientStillWorks() {
	// List datasets (not supported on edge, should use main endpoint)
	datasets, err := s.client.Datasets.List(s.ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(datasets)

	// Get specific dataset
	dataset, err := s.client.Datasets.Get(s.ctx, s.dataset.ID)
	s.Require().NoError(err)
	s.Require().NotNil(dataset)
	s.Equal(s.dataset.ID, dataset.ID)
}
