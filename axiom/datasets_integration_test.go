//go:build integration

package axiom_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/apl"
	"github.com/axiomhq/axiom-go/axiom/query"
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

var ingestEvents = []axiom.Event{
	{
		"time":        "17/May/2015:08:05:32 +0000",
		"remote_ip":   "93.180.71.1",
		"remote_user": "-",
		"request":     "GET /downloads/product_1 HTTP/1.1",
		"response":    304,
		"bytes":       0,
		"referrer":    "-",
		"agent":       "Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)",
	},
	{
		"time":        "17/May/2015:08:05:33 +0000",
		"remote_ip":   "93.180.71.2",
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
		Name:        "test-axiom-go-datasets-" + datasetSuffix,
		Description: "This is a test dataset for datasets integration tests.",
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
	s.Require().NotEmpty(datasets)

	s.Contains(datasets, s.dataset)

	// Let's ingest some data from a reader source...
	var (
		ingested bytes.Buffer
		r        io.Reader

		resetBuffer = func(contentEncoders ...axiom.ContentEncoder) {
			ingested.Reset()
			r = io.TeeReader(strings.NewReader(ingestData), &ingested)

			for _, contentEncoder := range contentEncoders {
				var ceErr error
				r, ceErr = contentEncoder(r)
				s.Require().NoError(ceErr)
			}
		}
	)
	resetBuffer()
	ingestStatus, err := s.client.Datasets.Ingest(s.ctx, s.dataset.ID, r, axiom.JSON, axiom.Identity, axiom.IngestOptions{})
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(ingestStatus.Ingested, 2)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)
	s.EqualValues(ingested.Len(), ingestStatus.ProcessedBytes)

	// ... but gzip encoded...
	resetBuffer(axiom.GzipEncoder)
	ingestStatus, err = s.client.Datasets.Ingest(s.ctx, s.dataset.ID, r, axiom.JSON, axiom.Gzip, axiom.IngestOptions{})
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(ingestStatus.Ingested, 2)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)
	s.EqualValues(ingested.Len(), ingestStatus.ProcessedBytes)

	// ... but zstd encoded...
	resetBuffer(axiom.ZstdEncoder)
	ingestStatus, err = s.client.Datasets.Ingest(s.ctx, s.dataset.ID, r, axiom.JSON, axiom.Zstd, axiom.IngestOptions{})
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

	// Make sure we aren't to fast for the server.
	time.Sleep(15 * time.Second)

	// Get the dataset info and make sure four events have been ingested.
	datasetInfo, err := s.client.Datasets.Info(s.ctx, s.dataset.ID)
	s.Require().NoError(err)
	s.Require().NotNil(datasetInfo)

	s.Equal(s.dataset.Name, datasetInfo.Name)
	s.EqualValues(8, datasetInfo.NumEvents)
	s.NotEmpty(datasetInfo.Fields)

	// Get the statistics of all datasets.
	datasetStats, err := s.client.Datasets.Stats(s.ctx)
	s.Require().NoError(err)
	s.Require().NotNil(datasetStats)

	// Get the fields of all datasets and make sure the fields of our dataset
	// info match those.
	datasetFields, err := s.client.Datasets.Fields(s.ctx)
	s.Require().NoError(err)
	s.Require().NotNil(datasetFields)

	if fields := datasetFields[dataset.Name]; s.NotNil(fields, "no fields for dataset %s", dataset.Name) {
		s.Equal(datasetInfo.Fields, fields, "dataset info fields do not match dataset %s entry in the global list of dataset fields", dataset.Name)
	}

	// Update a field of our dataset.
	field, err := s.client.Datasets.UpdateField(s.ctx, s.dataset.ID, "response", axiom.FieldUpdateRequest{
		Description: "HTTP status code returned as part of the response",
	})
	s.Require().NoError(err)
	s.Require().NotNil(field)

	s.Equal("response", field.Name)
	s.Equal("HTTP status code returned as part of the response", field.Description)
	s.Equal("integer", field.Type)

	// Run a query and make sure we see some results.
	queryResult, err := s.client.Datasets.Query(s.ctx, s.dataset.ID, query.Query{
		StartTime: time.Now().UTC().Add(-time.Minute),
		EndTime:   time.Now().UTC(),
	}, query.Options{
		SaveKind: query.Analytics,
	})
	s.Require().NoError(err)
	s.Require().NotNil(queryResult)

	// This needs to pass in order for the history query test to have an input.
	s.Require().NotEmpty(queryResult.SavedQueryID)

	// s.EqualValues(1, queryResult.Status.BlocksExamined) // FIXME(lukasmalkmus): For some reason we get "2" here?!
	s.EqualValues(8, queryResult.Status.RowsExamined)
	s.EqualValues(8, queryResult.Status.RowsMatched)
	s.Len(queryResult.Matches, 8)

	// Run another query but using APL.
	q := apl.Query(fmt.Sprintf("['%s']", s.dataset.ID))
	aplQueryResult, err := s.client.Datasets.APLQuery(s.ctx, q, apl.Options{
		Save: true,
	})
	s.Require().NoError(err)
	s.Require().NotNil(aplQueryResult)

	// This needs to pass in order for the history query test to have an input.
	s.Require().NotEmpty(aplQueryResult.SavedQueryID)

	// s.EqualValues(1, aplQueryResult.Status.BlocksExamined) // FIXME(lukasmalkmus): For some reason we get "2" here?!
	s.EqualValues(8, aplQueryResult.Status.RowsExamined)
	s.EqualValues(8, aplQueryResult.Status.RowsMatched)
	s.Len(aplQueryResult.Matches, 8)
	s.Contains(aplQueryResult.Datasets, s.dataset.ID)

	// Run a more complex query.
	complexQueryResult, err := s.client.Datasets.Query(s.ctx, s.dataset.ID, query.Query{
		StartTime: time.Now().UTC().Add(-time.Minute),
		EndTime:   time.Now().UTC(),
		Aggregations: []query.Aggregation{
			{
				Alias: "event_count",
				Op:    query.OpCount,
				Field: "*",
			},
		},
		GroupBy: []string{"success", "remote_ip"},
		Filter: query.Filter{
			Op:    query.OpEqual,
			Field: "response",
			Value: 304,
		},
		Order: []query.Order{
			{
				Field: "success",
				Desc:  true,
			},
			{
				Field: "remote_ip",
				Desc:  false,
			},
		},
		VirtualFields: []query.VirtualField{
			{
				Alias:      "success",
				Expression: "response < 400",
			},
		},
		Projections: []query.Projection{
			{
				Field: "remote_ip",
				Alias: "ip",
			},
		},
	}, query.Options{})
	s.Require().NoError(err)
	s.Require().NotNil(complexQueryResult)

	s.EqualValues(8, complexQueryResult.Status.RowsExamined)
	s.EqualValues(8, complexQueryResult.Status.RowsMatched)
	if s.Len(complexQueryResult.Buckets.Totals, 2) {
		agg := complexQueryResult.Buckets.Totals[0].Aggregations[0]
		s.EqualValues("event_count", agg.Alias)
		s.EqualValues(4, agg.Value)
	}

	// Trim the dataset down to a minimum.
	trimResult, err := s.client.Datasets.Trim(s.ctx, s.dataset.ID, time.Second)
	s.Require().NoError(err)
	s.Require().NotNil(trimResult)

	// HINT(lukasmalkmus): There are no blocks to trim in this test.
	s.EqualValues(0, trimResult.BlocksDeleted)

	// HINT(lukasmalkmus): Disable strict decoding for the query history. This
	// is in place because the API returns a slightly different model with a lot
	// of empty fields which are never set for a history query. Those are not
	// part of the client side model for ease of use.
	err = s.client.Options(axiom.SetStrictDecoding(false))
	s.Require().NoError(err)
	defer func() {
		optsErr := s.client.Options(axiom.SetStrictDecoding(strictDecoding))
		s.Require().NoError(optsErr)
	}()

	// Give the server some time to store the queries as they are processed
	// asynchronously.
	time.Sleep(time.Second * 15)

	historyQuery, err := s.client.Datasets.History(s.ctx, queryResult.SavedQueryID)
	s.Require().NoError(err)
	s.Require().NotNil(historyQuery)

	s.Equal(queryResult.SavedQueryID, historyQuery.ID)
	s.Equal(query.Analytics, historyQuery.Kind)

	historyQuery, err = s.client.Datasets.History(s.ctx, aplQueryResult.SavedQueryID)
	s.Require().NoError(err)
	s.Require().NotNil(historyQuery)

	s.Equal(aplQueryResult.SavedQueryID, historyQuery.ID)
	s.Equal(query.APL, historyQuery.Kind)
}
