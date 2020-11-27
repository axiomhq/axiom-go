// +build integration

package axiom_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
)

const ingestData = `[
	{
		"time": "17/May/2015:08:05:32 +0000",
		"remote_ip": "93.180.71.3",
		"remote_user": "-",
		"request": "GET /downloads/product_1 HTTP/1.1",
		"response": 304,
		"bytes": 0,
		"referrer": "-",
		"agent": "Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)"
	},
	{
		"time": "17/May/2015:08:05:32 +0000",
		"remote_ip": "93.180.71.3",
		"remote_user": "-",
		"request": "GET /downloads/product_1 HTTP/1.1",
		"response": 304,
		"bytes": 0,
		"referrer": "-",
		"agent": "Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)"
	}
]`

var ingestEvents = []axiom.Event{
	{
		"time":        "17/May/2015:08:05:32 +0000",
		"remote_ip":   "93.180.71.3",
		"remote_user": "-",
		"request":     "GET /downloads/product_1 HTTP/1.1",
		"response":    304,
		"bytes":       0,
		"referrer":    "-",
		"agent":       "Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)",
	},
	{
		"time":        "17/May/2015:08:05:32 +0000",
		"remote_ip":   "93.180.71.3",
		"remote_user": "-",
		"request":     "GET /downloads/product_1 HTTP/1.1",
		"response":    304,
		"bytes":       0,
		"referrer":    "-",
		"agent":       "Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)",
	},
}

// DatasetsTestSuite tests all methods of the Axiom Datasets API against a live
// deployment.
type DatasetsTestSuite struct {
	IntegrationTestSuite

	dataset *axiom.Dataset
}

func TestDatasetsTestSuite(t *testing.T) {
	suite.Run(t, new(DatasetsTestSuite))
}

func (s *DatasetsTestSuite) SetupSuite() {
	s.IntegrationTestSuite.SetupSuite()

	var err error
	s.dataset, err = s.client.Datasets.Create(s.suiteCtx, axiom.DatasetCreateRequest{
		Name:        "test",
		Description: "This is a test dataset",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.dataset)
}

func (s *DatasetsTestSuite) TearDownSuite() {
	// Teardown routines use their own context to avoid not being run at all
	// when the suite gets cancelled or times out.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := s.client.Datasets.Delete(ctx, s.dataset.ID)
	s.NoError(err)

	s.IntegrationTestSuite.TearDownSuite()
}

func (s *DatasetsTestSuite) TestUpdate() {
	s.T().Skip("Activate if awkward API validation (matching ID in param and body) has been fixed.")

	dataset, err := s.client.Datasets.Update(s.ctx, s.dataset.ID, axiom.DatasetUpdateRequest{
		Description: "This is a soon to be filled test dataset",
	})
	s.Require().NoError(err)
	s.Require().NotNil(dataset)

	s.dataset = dataset
}

func (s *DatasetsTestSuite) TestGet() {
	dataset, err := s.client.Datasets.Get(s.ctx, s.dataset.ID)
	s.Require().NoError(err)
	s.Require().NotNil(dataset)

	s.Equal(s.dataset, dataset)
}

func (s *DatasetsTestSuite) TestList() {
	datasets, err := s.client.Datasets.List(s.ctx)
	s.Require().NoError(err)
	s.Require().NotNil(datasets)

	s.Contains(datasets, s.dataset)
}

func (s *DatasetsTestSuite) TestInfoAndStats() {
	s.T().Skip("Enable as soon as the API response has been fixed!")

	datasetInfo, err := s.client.Datasets.Info(s.ctx, s.dataset.ID)
	s.Require().NoError(err)
	s.Require().NotNil(datasetInfo)

	s.Equal(datasetInfo.Name, s.dataset.Name)
	s.Equal(datasetInfo.NumEvents, 2)

	datasetStats, err := s.client.Datasets.Stats(s.ctx)
	s.Require().NoError(err)
	s.Require().NotNil(datasetStats)

	s.Contains(datasetStats, datasetInfo)
}

func (s *DatasetsTestSuite) TestIngest() {
	var (
		ingested bytes.Buffer
		r        = io.TeeReader(strings.NewReader(ingestData), &ingested)
	)
	ingestStatus, err := s.client.Datasets.Ingest(s.ctx, s.dataset.ID, r, axiom.JSON, axiom.Identity, axiom.IngestOptions{})
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(ingestStatus.Ingested, 2)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)
	s.EqualValues(ingestStatus.ProcessedBytes, ingested.Len())
}

func (s *DatasetsTestSuite) TestIngestEvents() {
	ingestStatus, err := s.client.Datasets.IngestEvents(s.ctx, s.dataset.ID, axiom.IngestOptions{}, ingestEvents...)
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(ingestStatus.Ingested, 2)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)
}

func (s *DatasetsTestSuite) TestQuery() {
	s.T().Skip("Activate if we know why the result set is empty.")

	queryResult, err := s.client.Datasets.Query(s.ctx, s.dataset.ID, axiom.Query{
		StartTime: time.Time{},
		EndTime:   time.Now(),
	}, axiom.QueryOptions{})
	s.Require().NoError(err)
	s.Require().NotNil(queryResult)

	s.EqualValues(0, queryResult.Status.BlocksExamined)
	s.EqualValues(4, queryResult.Status.RowsExamined)
	s.EqualValues(4, queryResult.Status.RowsMatched)
	s.Len(queryResult.Matches, 4)
}
