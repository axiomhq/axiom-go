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
	"github.com/axiomhq/axiom-go/axiom/ingest"
	"github.com/axiomhq/axiom-go/axiom/query"
	"github.com/axiomhq/axiom-go/axiom/querylegacy"
)

const (
	ingestData = `[
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

	csvIngestData = `17/May/2015:08:05:30 +0000,93.180.71.1,-,GET /downloads/product_1 HTTP/1.1,304,0,-,Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21
		17/May/2015:08:05:31 +0000,93.180.71.2,-,GET /downloads/product_1 HTTP/1.1,304,0,-,Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)`

	csvIngestDataHeader = `time,remote_ip,remote_user,request,response,bytes,referrer,agent
		17/May/2015:08:05:30 +0000,93.180.71.1,-,GET /downloads/product_1 HTTP/1.1,304,0,-,Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21
		17/May/2015:08:05:31 +0000,93.180.71.2,-,GET /downloads/product_1 HTTP/1.1,304,0,-,Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)`
)

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

	// Setup once per test.
	dataset *axiom.Dataset
}

func TestDatasetsTestSuite(t *testing.T) {
	suite.Run(t, new(DatasetsTestSuite))
}

func (s *DatasetsTestSuite) SetupTest() {
	s.IntegrationTestSuite.SetupTest()

	var err error
	s.dataset, err = s.client.Datasets.Create(s.ctx, axiom.DatasetCreateRequest{
		Name:        "test-axiom-go-datasets-" + datasetSuffix,
		Description: "This is a test dataset for datasets integration tests.",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.dataset)
}

func (s *DatasetsTestSuite) TearDownTest() {
	// Teardown routines use their own context to avoid not being run at all
	// when the suite gets cancelled or times out.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(s.ctx), time.Second*15)
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

		resetBuffer = func(data string, contentEncoders ...axiom.ContentEncoder) {
			ingested.Reset()
			r = io.TeeReader(strings.NewReader(data), &ingested)

			for _, contentEncoder := range contentEncoders {
				var ceErr error
				r, ceErr = contentEncoder(r)
				s.Require().NoError(ceErr)
			}
		}
	)

	resetBuffer(ingestData)
	ingestStatus, err := s.client.Datasets.Ingest(s.ctx, s.dataset.ID, r, axiom.JSON, axiom.Identity, ingest.SetEventLabel("region", "eu-west-1"))
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(ingestStatus.Ingested, 2)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)
	s.EqualValues(ingested.Len()+22, ingestStatus.ProcessedBytes) // 22 bytes extra for the event label

	// ... but gzip encoded...
	resetBuffer(ingestData, axiom.GzipEncoder())
	ingestStatus, err = s.client.Datasets.Ingest(s.ctx, s.dataset.ID, r, axiom.JSON, axiom.Gzip)
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(ingestStatus.Ingested, 2)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)
	s.EqualValues(ingested.Len(), ingestStatus.ProcessedBytes)

	// ... but zstd encoded...
	resetBuffer(ingestData, axiom.ZstdEncoder())
	ingestStatus, err = s.client.Datasets.Ingest(s.ctx, s.dataset.ID, r, axiom.JSON, axiom.Zstd)
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(ingestStatus.Ingested, 2)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)
	s.EqualValues(ingested.Len(), ingestStatus.ProcessedBytes)

	// ... and from a map source...
	ingestStatus, err = s.client.Datasets.IngestEvents(s.ctx, s.dataset.ID, ingestEvents)
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(ingestStatus.Ingested, 2)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)
	s.EqualValues(448, int(ingestStatus.ProcessedBytes))

	// ... and from a channel source ...
	ingestStatus, err = s.client.Datasets.IngestChannel(s.ctx, s.dataset.ID, getEventChan())
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(ingestStatus.Ingested, 2)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)
	s.EqualValues(448, int(ingestStatus.ProcessedBytes))

	// ... and from a CSV reader source with header...
	resetBuffer(csvIngestDataHeader)
	ingestStatus, err = s.client.Datasets.Ingest(s.ctx, s.dataset.ID, r, axiom.CSV, axiom.Identity)
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(ingestStatus.Ingested, 2)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)
	s.EqualValues(325, int(ingestStatus.ProcessedBytes))

	// ... and from a CSV reader source without header.
	resetBuffer(csvIngestData)
	ingestStatus, err = s.client.Datasets.Ingest(s.ctx, s.dataset.ID, r, axiom.CSV, axiom.Identity,
		ingest.SetCSVFields("time", "remote_ip", "remote_user", "request", "response", "bytes", "referrer", "agent"))
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(ingestStatus.Ingested, 2)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)
	s.EqualValues(258, int(ingestStatus.ProcessedBytes))

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

	s.NotZero(queryResult.Status.ElapsedTime)
	s.EqualValues(14, queryResult.Status.RowsExamined)
	s.EqualValues(14, queryResult.Status.RowsMatched)
	if s.Len(queryResult.Tables, 1) {
		table := queryResult.Tables[0]

		if s.Len(table.Sources, 1) {
			s.Equal(s.dataset.ID, table.Sources[0].Name)
		}

		// FIXME(lukasmalkmus): Tabular results format is not yet returning the
		// _rowID column.
		s.Len(table.Fields, 11)  // 8 event fields + 1 label field + 2 system fields
		s.Len(table.Columns, 11) // 8 event fields + 1 label field + 2 system fields
		// s.Len(table.Fields, 12) // 8 event fields + 1 label field + 3 system fields
		// s.Len(table.Columns, 12) // 8 event fields + 1 label field + 3 system fields
	}

	// ... and a slightly more complex (analytic) APL query.
	apl = fmt.Sprintf("['%s'] | summarize topk(remote_ip, 1)", s.dataset.ID)
	queryResult, err = s.client.Datasets.Query(s.ctx, apl,
		query.SetStartTime(startTime),
		query.SetEndTime(endTime),
	)
	s.Require().NoError(err)
	s.Require().NotNil(queryResult)

	s.NotZero(queryResult.Status.ElapsedTime)
	s.EqualValues(14, queryResult.Status.RowsExamined)
	s.EqualValues(14, queryResult.Status.RowsMatched)
	if s.Len(queryResult.Tables, 1) {
		table := queryResult.Tables[0]

		if s.Len(table.Sources, 1) {
			s.Equal(s.dataset.ID, table.Sources[0].Name)
		}

		if s.Len(table.Fields, 1) && s.NotNil(table.Fields[0].Aggregation) {
			agg := table.Fields[0].Aggregation

			s.Equal(query.OpTopk, agg.Op)
			s.Equal([]string{"remote_ip"}, agg.Fields)
			s.Equal([]any{1.}, agg.Args)
		}

		if s.Len(table.Columns, 1) && s.Len(table.Columns[0], 1) {
			v := table.Columns[0][0].([]any)
			m := v[0].(map[string]any)

			s.Equal("93.180.71.1", m["key"])
			s.Equal(7., m["count"])
			s.Equal(0., m["error"])
		}
	}

	// Also run a legacy query and make sure we see some results.
	legacyQueryResult, err := s.client.Datasets.QueryLegacy(s.ctx, s.dataset.ID, querylegacy.Query{
		StartTime: startTime,
		EndTime:   endTime,
	}, querylegacy.Options{})
	s.Require().NoError(err)
	s.Require().NotNil(legacyQueryResult)

	s.NotZero(queryResult.Status.ElapsedTime)
	s.EqualValues(14, legacyQueryResult.Status.RowsExamined)
	s.EqualValues(14, legacyQueryResult.Status.RowsMatched)
	s.Len(legacyQueryResult.Matches, 14)

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
			Op:    querylegacy.OpExists,
			Field: "response",
			Children: []querylegacy.Filter{
				{
					Op:    querylegacy.OpContains,
					Field: "request",
					Value: "GET",
				},
			},
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
				Expression: "toint(response) < 400",
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

	s.EqualValues(14, complexLegacyQueryResult.Status.RowsExamined)
	s.EqualValues(14, complexLegacyQueryResult.Status.RowsMatched)
	if s.Len(complexLegacyQueryResult.Buckets.Totals, 2) {
		agg := complexLegacyQueryResult.Buckets.Totals[0].Aggregations[0]
		s.EqualValues("event_count", agg.Alias)
		s.EqualValues(7, agg.Value)
	}

	// Trim the dataset down to a minimum.
	err = s.client.Datasets.Trim(s.ctx, s.dataset.ID, time.Second)
	s.Require().NoError(err)
}

func (s *DatasetsTestSuite) TestCursor() {
	// Let's ingest some data.
	now := time.Now().Truncate(time.Second)
	_, err := s.client.Datasets.IngestEvents(s.ctx, s.dataset.ID, []axiom.Event{
		{ // Oldest
			ingest.TimestampField: now.Add(-time.Second * 3),
			"foo":                 "bar",
		},
		{
			ingest.TimestampField: now.Add(-time.Second * 2),
			"foo":                 "baz",
		},
		{ // Newest
			ingest.TimestampField: now.Add(-time.Second * 1),
			"foo":                 "buz",
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

	// FIXME(lukasmalkmus): Tabular results format is not yet returning the
	// _rowID column.
	s.T().Skip()

	// HINT(lukasmalkmus): Expecting four columns: _time, _sysTime, _rowID, foo.
	// This is only checked once for the first query result to verify the
	// dataset scheme. The following queries will only check the results in the
	// columns.
	s.Require().Len(queryResult.Tables, 1)
	s.Require().Len(queryResult.Tables[0].Columns, 4)
	s.Require().Len(queryResult.Tables[0].Columns[0], 3)

	if s.Len(queryResult.Tables, 1) {
		s.Equal("buz", queryResult.Tables[0].Columns[2][0])
		s.Equal("baz", queryResult.Tables[0].Columns[2][1])
		s.Equal("bar", queryResult.Tables[0].Columns[2][2])
	}

	// HINT(lukasmalkmus): In a real-world scenario, the cursor would be
	// retrieved from the query status MinCursor or MaxCursor fields, depending
	// on the queries sort order.
	midRowID := queryResult.Tables[0].Columns[0][2].(string)

	// Query events with a cursor in descending order...
	apl = fmt.Sprintf("['%s'] | sort by _time desc", s.dataset.ID)
	queryResult, err = s.client.Datasets.Query(s.ctx, apl,
		query.SetStartTime(startTime),
		query.SetEndTime(endTime),
		query.SetCursor(midRowID, false),
	)
	s.Require().NoError(err)

	// "buz" and "baz" skipped by the cursor, only "bar" is returned. The cursor
	// is exclusive, so "baz" is not included.
	if s.Len(queryResult.Tables[0].Columns[0], 1) {
		s.Equal("bar", queryResult.Tables[0].Columns[0][0])
	}

	// ...again, but with the cursor inclusive.
	queryResult, err = s.client.Datasets.Query(s.ctx, apl,
		query.SetStartTime(startTime),
		query.SetEndTime(endTime),
		query.SetCursor(midRowID, true),
	)
	s.Require().NoError(err)

	// "buz" skipped by the cursor, only "baz" and "bar" is returned. The cursor
	// is inclusive, so "baz" is included.
	if s.Len(queryResult.Tables[0].Columns[0], 2) {
		s.Equal("baz", queryResult.Tables[0].Columns[0][0])
		s.Equal("bar", queryResult.Tables[0].Columns[0][1])
	}

	// Query events with a cursor in ascending order...
	apl = fmt.Sprintf("['%s'] | sort by _time asc", s.dataset.ID)
	queryResult, err = s.client.Datasets.Query(s.ctx, apl,
		query.SetStartTime(startTime),
		query.SetEndTime(endTime),
		query.SetCursor(midRowID, false),
	)
	s.Require().NoError(err)

	// "bar" and "baz" skipped by the cursor, only "buz" is returned. The cursor
	// is exclusive, so "baz" is not included.
	if s.Len(queryResult.Tables[0].Columns[0], 1) {
		s.Equal("buz", queryResult.Tables[0].Columns[0][0])
	}

	// ...again, but with the cursor inclusive.
	queryResult, err = s.client.Datasets.Query(s.ctx, apl,
		query.SetStartTime(startTime),
		query.SetEndTime(endTime),
		query.SetCursor(midRowID, true),
	)
	s.Require().NoError(err)

	// "bar" skipped by the cursor, only "baz" and "buz" is returned. The cursor
	// is inclusive, so "baz" is included.
	if s.Len(queryResult.Tables[0].Columns[0], 2) {
		s.Equal("baz", queryResult.Tables[0].Columns[0][0])
		s.Equal("buz", queryResult.Tables[0].Columns[0][1])
	}
}

func getEventChan() <-chan axiom.Event {
	eventCh := make(chan axiom.Event)
	go func() {
		for _, e := range ingestEvents {
			eventCh <- e
		}
		close(eventCh)
	}()
	return eventCh
}
