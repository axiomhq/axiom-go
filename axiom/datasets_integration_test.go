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
	"github.com/axiomhq/axiom-go/axiom/query"
	"github.com/axiomhq/axiom-go/axiom/querylegacy"
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

	now := time.Now()
	startTime := now.Add(-time.Minute)
	endTime := now.Add(time.Minute)

	// Run a query and make sure we see some results.
	simpleQuery := querylegacy.Query{
		StartTime: startTime,
		EndTime:   endTime,
	}
	simpleQueryResult, err := s.client.Datasets.QueryLegacy(s.ctx, s.dataset.ID, simpleQuery, querylegacy.Options{
		SaveKind: querylegacy.Analytics,
	})
	s.Require().NoError(err)
	s.Require().NotNil(simpleQueryResult)

	// This needs to pass in order for the history query test to have an input.
	s.Require().NotEmpty(simpleQueryResult.SavedQueryID)

	// s.EqualValues(1, simpleQueryResult.Status.BlocksExamined) // FIXME(lukasmalkmus): For some reason we get "2" here?!
	s.EqualValues(8, simpleQueryResult.Status.RowsExamined)
	s.EqualValues(8, simpleQueryResult.Status.RowsMatched)
	s.Len(simpleQueryResult.Matches, 8)

	// Run another query but using APL.
	aplQuery := query.Query(fmt.Sprintf("['%s']", s.dataset.ID))
	aplQueryResult, err := s.client.Datasets.Query(s.ctx, aplQuery, query.Options{
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

	// Run a more complex querylegacy.
	complexQuery := querylegacy.Query{
		StartTime: startTime,
		EndTime:   endTime,
		Aggregations: []querylegacy.Aggregation{
			{
				Alias: "event_count",
				Op:    querylegacy.OpCount,
				Field: "*",
			},
		},
		GroupBy: []string{"success", "remote_ip"},
		Filter: querylegacy.Filter{
			Op:    querylegacy.OpEqual,
			Field: "response",
			Value: 304,
		},
		Order: []querylegacy.Order{
			{
				Field: "success",
				Desc:  true,
			},
			{
				Field: "remote_ip",
				Desc:  false,
			},
		},
		VirtualFields: []querylegacy.VirtualField{
			{
				Alias:      "success",
				Expression: "response < 400",
			},
		},
		Projections: []querylegacy.Projection{
			{
				Field: "remote_ip",
				Alias: "ip",
			},
		},
	}
	complexQueryResult, err := s.client.Datasets.QueryLegacy(s.ctx, s.dataset.ID, complexQuery, querylegacy.Options{})
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
}
