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
	"github.com/axiomhq/axiom-go/axiom/query"
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
		Name:        "test-" + randString(),
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

func (s *DatasetsTestSuite) Test() {
	// Let's update the dataset.
	dataset, err := s.client.Datasets.Update(s.ctx, s.dataset.ID, axiom.DatasetUpdateRequest{
		Description: "This is a soon to be filled test dataset",
	})
	s.Require().NoError(err)
	s.Require().NotNil(dataset)

	s.dataset = dataset

	// Get the dataset and make sure it matches what we have updated it to.
	dataset, err = s.client.Datasets.Get(s.ctx, s.dataset.ID)
	s.Require().NoError(err)
	s.Require().NotNil(dataset)

	s.Equal(s.dataset, dataset)

	// List all datasets and make sure the created dataset is part of that
	// list.
	datasets, err := s.client.Datasets.List(s.ctx)
	s.Require().NoError(err)
	s.Require().NotNil(datasets)

	s.Contains(datasets, s.dataset)

	// Let's ingest some data from a reader source...
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
	s.EqualValues(ingested.Len(), ingestStatus.ProcessedBytes)

	// ... and a map.
	ingestStatus, err = s.client.Datasets.IngestEvents(s.ctx, s.dataset.ID, axiom.IngestOptions{}, ingestEvents...)
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(ingestStatus.Ingested, 2)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)

	// Make sure we don't overtake the server.
	time.Sleep(2 * time.Second)

	// Get the dataset info and make sure four events have been ingested.
	datasetInfo, err := s.client.Datasets.Info(s.ctx, s.dataset.ID)
	s.Require().NoError(err)
	s.Require().NotNil(datasetInfo)

	s.Equal(s.dataset.Name, datasetInfo.Name)
	s.EqualValues(4, datasetInfo.NumEvents)
	s.NotEmpty(datasetInfo.Fields)

	// Get the stats of all datasets and make sure our dataset info is included
	// in that list.
	datasetStats, err := s.client.Datasets.Stats(s.ctx)
	s.Require().NoError(err)
	s.Require().NotNil(datasetStats)

	var contains bool
	for _, stat := range datasetStats.Datasets {
		if contains = stat.Name == dataset.Name; contains {
			break
		}
	}
	s.True(contains, "stats do not contain the dataset we created for this test")

	// Run a query and make sure we see some results.
	queryResult, err := s.client.Datasets.Query(s.ctx, s.dataset.ID, query.Query{
		StartTime: time.Now().UTC().Add(-time.Minute),
		EndTime:   time.Now().UTC(),
	}, query.Options{})
	s.Require().NoError(err)
	s.Require().NotNil(queryResult)

	s.EqualValues(1, queryResult.Status.BlocksExamined)
	s.EqualValues(4, queryResult.Status.RowsExamined)
	s.EqualValues(4, queryResult.Status.RowsMatched)
	s.Len(queryResult.Matches, 4)
}

func (s *DatasetsTestSuite) TestHistory() {
	// HINT(lukasmalkmus): This test initializes a new client to make sure
	// strict decoding is never set to this method. After this test, is gets
	// set to its previous state.
	// This is in place because the API returns a slightly different model with
	// a lot of empty fields which are never set for a history query. Those are
	// not part of the client side model for ease of use.
	s.newClient()
	defer func() {
		if strictDecoding {
			err := s.client.Options(axiom.SetStrictDecoding())
			s.Require().NoError(err)
		}
	}()

	query, err := s.client.Datasets.History(s.ctx, historyQueryID)
	s.Require().NoError(err)
	s.Require().NotNil(query)
}
