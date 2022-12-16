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
}

func (s *DatasetsTestSuite) TearDownSuite() {
	s.IntegrationTestSuite.TearDownSuite()
}

func (s *DatasetsTestSuite) SetupTest() {
	s.IntegrationTestSuite.SetupTest()

	var err error
	s.dataset, err = s.client.Datasets.Create(s.suiteCtx, axiom.DatasetCreateRequest{
		Name:        "test-axiom-go-datasets-" + datasetSuffix,
		Description: "This is a test dataset for datasets integration tests.",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.dataset)
}

func (s *DatasetsTestSuite) TearDownTest() {
	// Teardown routines use their own context to avoid not being run at all
	// when the suite gets cancelled or times out.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := s.client.Datasets.Delete(ctx, s.dataset.ID)
	s.NoError(err)

	s.IntegrationTestSuite.TearDownTest()
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
	ingestStatus, err := s.client.Datasets.Ingest(s.ctx, s.dataset.ID, r, axiom.JSON, axiom.Identity)
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(ingestStatus.Ingested, 2)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)
	s.EqualValues(ingested.Len(), ingestStatus.ProcessedBytes)

	// ... but gzip encoded...
	resetBuffer(axiom.GzipEncoder())
	ingestStatus, err = s.client.Datasets.Ingest(s.ctx, s.dataset.ID, r, axiom.JSON, axiom.Gzip)
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(ingestStatus.Ingested, 2)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)
	s.EqualValues(ingested.Len(), ingestStatus.ProcessedBytes)

	// ... but zstd encoded...
	resetBuffer(axiom.ZstdEncoder())
	ingestStatus, err = s.client.Datasets.Ingest(s.ctx, s.dataset.ID, r, axiom.JSON, axiom.Zstd)
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(ingestStatus.Ingested, 2)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)
	s.EqualValues(ingested.Len(), ingestStatus.ProcessedBytes)

	// ... and a map.
	ingestStatus, err = s.client.Datasets.IngestEvents(s.ctx, s.dataset.ID, ingestEvents)
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(ingestStatus.Ingested, 2)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)

	now := time.Now().Truncate(time.Second)
	startTime := now.Add(-time.Minute)
	endTime := now.Add(time.Minute)

	// Run a simple APL query.
	apl := fmt.Sprintf("['%s']", s.dataset.ID)
	queryResult, err := s.client.Datasets.Query(s.ctx, apl,
		query.SetStartTime(startTime),
		query.SetEndTime(endTime),
	)
	s.Require().NoError(err)
	s.Require().NotNil(queryResult)

	s.EqualValues(1, queryResult.Status.BlocksExamined)
	s.EqualValues(8, queryResult.Status.RowsExamined)
	s.EqualValues(8, queryResult.Status.RowsMatched)
	s.Len(queryResult.Matches, 8)

	// Also run a legacy query and make sure we see some results.
	legacyQueryResult, err := s.client.Datasets.QueryLegacy(s.ctx, s.dataset.ID, querylegacy.Query{
		StartTime: startTime,
		EndTime:   endTime,
	}, querylegacy.Options{})
	s.Require().NoError(err)
	s.Require().NotNil(legacyQueryResult)

	s.EqualValues(1, legacyQueryResult.Status.BlocksExamined)
	s.EqualValues(8, legacyQueryResult.Status.RowsExamined)
	s.EqualValues(8, legacyQueryResult.Status.RowsMatched)
	s.Len(legacyQueryResult.Matches, 8)

	// Run a more complex legacy query.
	complexLegacyQuery := querylegacy.Query{
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

	complexLegacyQueryResult, err := s.client.Datasets.QueryLegacy(s.ctx, s.dataset.ID, complexLegacyQuery, querylegacy.Options{})
	s.Require().NoError(err)
	s.Require().NotNil(complexLegacyQueryResult)

	s.EqualValues(8, complexLegacyQueryResult.Status.RowsExamined)
	s.EqualValues(8, complexLegacyQueryResult.Status.RowsMatched)
	if s.Len(complexLegacyQueryResult.Buckets.Totals, 2) {
		agg := complexLegacyQueryResult.Buckets.Totals[0].Aggregations[0]
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

func (s *DatasetsTestSuite) TestCursor() {
	// Let's ingest some data.
	now := time.Now().Truncate(time.Second)
	_, err := s.client.Datasets.IngestEvents(s.ctx, s.dataset.ID, []axiom.Event{
		{ // Oldest
			"_time": now.Add(-time.Second * 3),
			"foo":   "bar",
		},
		{
			"_time": now.Add(-time.Second * 2),
			"foo":   "baz",
		},
		{ // Newest
			"_time": now.Add(-time.Second * 1),
			"foo":   "buz",
		},
	})
	s.Require().NoError(err)

	startTime := now.Add(-time.Minute)
	endTime := now.Add(time.Minute)

	// Query all events.
	apl := fmt.Sprintf("['%s'] | sort by _time", s.dataset.ID)
	queryResult, err := s.client.Datasets.Query(s.ctx, apl,
		query.SetStartTime(startTime),
		query.SetEndTime(endTime),
	)
	s.Require().NoError(err)

	if s.Len(queryResult.Matches, 3) {
		s.Equal("buz", queryResult.Matches[0].Data["foo"])
		s.Equal("baz", queryResult.Matches[1].Data["foo"])
		s.Equal("bar", queryResult.Matches[2].Data["foo"])
	}

	midRowID := queryResult.Matches[1].RowID

	// Query events with a cursor in descending order.
	apl = fmt.Sprintf("['%s'] | sort by _time desc", s.dataset.ID)
	queryResult, err = s.client.Datasets.Query(s.ctx, apl,
		query.SetStartTime(startTime),
		query.SetEndTime(endTime),
		query.SetCursor(midRowID),
	)
	s.Require().NoError(err)

	if s.Len(queryResult.Matches, 2) {
		s.Equal("buz", queryResult.Matches[0].Data["foo"])
		s.Equal("baz", queryResult.Matches[1].Data["foo"])
	}

	// Query events with a cursor in ascending order.
	apl = fmt.Sprintf("['%s'] | sort by _time asc", s.dataset.ID)
	queryResult, err = s.client.Datasets.Query(s.ctx, apl,
		query.SetStartTime(startTime),
		query.SetEndTime(endTime),
		query.SetCursor(midRowID),
	)
	s.Require().NoError(err)

	if s.Len(queryResult.Matches, 2) {
		s.Equal("bar", queryResult.Matches[0].Data["foo"])
		s.Equal("baz", queryResult.Matches[1].Data["foo"])
	}
}
