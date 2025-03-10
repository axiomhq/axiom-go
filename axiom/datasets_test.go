package axiom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/axiom/ingest"
	"github.com/axiomhq/axiom-go/axiom/query"
	"github.com/axiomhq/axiom-go/axiom/querylegacy"
	"github.com/axiomhq/axiom-go/internal/test/testhelper"
)

const actQueryResp = `{
	"tables": [
		{
			"name": "0",
			"sources": [
				{
					"name": "test"
				}
			],
			"fields": [
				{
					"name": "_time",
					"type": "string"
				},
				{
					"name": "_sysTime",
					"type": "string"
				},
				{
					"name": "_rowId",
					"type": "string"
				},
				{
					"name": "agent",
					"type": "string"
				},
				{
					"name": "bytes",
					"type": "float64"
				},
				{
					"name": "referrer",
					"type": "string"
				},
				{
					"name": "remote_ip",
					"type": "string"
				},
				{
					"name": "remote_user",
					"type": "string"
				},
				{
					"name": "request",
					"type": "string"
				},
				{
					"name": "response",
					"type": "float64"
				},
				{
					"name": "time",
					"type": "string"
				}
			],
			"range": {
				"field": "_time",
				"start": "2023-03-21T13:38:51.735448191Z",
				"end": "2023-03-28T13:38:51.735448191Z"
			},
			"columns": [
				[
					"2020-11-19T11:06:31.569475746Z",
					"2020-11-19T11:06:31.569479846Z"
				],
				[
					"2020-11-19T11:06:31.581384524Z",
					"2020-11-19T11:06:31.581384524Z"
				],
				[
					"c776x1uafkpu-4918f6cb9000095-0",
					"c776x1uafnvq-4918f6cb9000095-1"
				],
				[
					"Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)",
					"Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)"
				],
				[
					0,
					0
				],
				[
					"-",
					"-"
				],
				[
					"93.180.71.3",
					"93.180.71.3"
				],
				[
					"-",
					"-"
				],
				[
					"GET /downloads/product_1 HTTP/1.1",
					"GET /downloads/product_1 HTTP/1.1"
				],
				[
					304,
					304
				],
				[
					"17/May/2015:08:05:32 +0000",
					"17/May/2015:08:05:23 +0000"
				]
			]
		}
	],
	"status": {
		"minCursor": "c776x1uafkpu-4918f6cb9000095-0",
		"maxCursor": "c776x1uafnvq-4918f6cb9000095-1",
		"elapsedTime": 542114,
		"rowsExamined": 142655,
		"rowsMatched": 142655
	}
}`

const actLegacyQueryResp = `{
	"status": {
		"minCursor": "c776x1uafkpu-4918f6cb9000095-0",
		"maxCursor": "c776x1uafnvq-4918f6cb9000095-1",
		"elapsedTime": 542114,
		"blocksExamined": 4,
		"rowsExamined": 142655,
		"rowsMatched": 142655,
		"numGroups": 0,
		"isPartial": false,
		"cacheStatus": 1,
		"minBlockTime": "2020-11-19T11:06:31.569475746Z",
		"maxBlockTime": "2020-11-27T12:06:38.966791794Z"
	},
	"matches": [
		{
			"_time": "2020-11-19T11:06:31.569475746Z",
			"_sysTime": "2020-11-19T11:06:31.581384524Z",
			"_rowId": "c776x1uafkpu-4918f6cb9000095-0",
			"data": {
				"agent": "Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)",
				"bytes": 0,
				"referrer": "-",
				"remote_ip": "93.180.71.3",
				"remote_user": "-",
				"request": "GET /downloads/product_1 HTTP/1.1",
				"response": 304,
				"time": "17/May/2015:08:05:32 +0000"
			}
		},
		{
			"_time": "2020-11-19T11:06:31.569479846Z",
			"_sysTime": "2020-11-19T11:06:31.581384524Z",
			"_rowId": "c776x1uafnvq-4918f6cb9000095-1",
			"data": {
				"agent": "Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)",
				"bytes": 0,
				"referrer": "-",
				"remote_ip": "93.180.71.3",
				"remote_user": "-",
				"request": "GET /downloads/product_1 HTTP/1.1",
				"response": 304,
				"time": "17/May/2015:08:05:23 +0000"
			}
		}
	]
}`

var (
	expQueryRes = &query.Result{
		Tables: []query.Table{
			{
				Name: "0",
				Sources: []query.Source{
					{
						Name: "test",
					},
				},
				Fields: []query.Field{
					{
						Name: "_time",
						Type: "string",
					},
					{
						Name: "_sysTime",
						Type: "string",
					},
					{
						Name: "_rowId",
						Type: "string",
					},
					{
						Name: "agent",
						Type: "string",
					},
					{
						Name: "bytes",
						Type: "float64",
					},
					{
						Name: "referrer",
						Type: "string",
					},
					{
						Name: "remote_ip",
						Type: "string",
					},
					{
						Name: "remote_user",
						Type: "string",
					},
					{
						Name: "request",
						Type: "string",
					},
					{
						Name: "response",
						Type: "float64",
					},
					{
						Name: "time",
						Type: "string",
					},
				},
				Range: &query.Range{
					Field: "_time",
					Start: parseTimeOrPanic("2023-03-21T13:38:51.735448191Z"),
					End:   parseTimeOrPanic("2023-03-28T13:38:51.735448191Z"),
				},
				Columns: []query.Column{
					[]any{
						"2020-11-19T11:06:31.569475746Z",
						"2020-11-19T11:06:31.569479846Z",
					},
					[]any{
						"2020-11-19T11:06:31.581384524Z",
						"2020-11-19T11:06:31.581384524Z",
					},
					[]any{
						"c776x1uafkpu-4918f6cb9000095-0",
						"c776x1uafnvq-4918f6cb9000095-1",
					},
					[]any{
						"Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)",
						"Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)",
					},
					[]any{
						float64(0),
						float64(0),
					},
					[]any{
						"-",
						"-",
					},
					[]any{
						"93.180.71.3",
						"93.180.71.3",
					},
					[]any{
						"-",
						"-",
					},
					[]any{
						"GET /downloads/product_1 HTTP/1.1",
						"GET /downloads/product_1 HTTP/1.1",
					},
					[]any{
						float64(304),
						float64(304),
					},
					[]any{
						"17/May/2015:08:05:32 +0000",
						"17/May/2015:08:05:23 +0000",
					},
				},
			},
		},
		Status: query.Status{
			ElapsedTime:  time.Microsecond * 542_114,
			MinCursor:    "c776x1uafkpu-4918f6cb9000095-0",
			MaxCursor:    "c776x1uafnvq-4918f6cb9000095-1",
			RowsExamined: 142655,
			RowsMatched:  142655,
		},
		TraceID: "abc",
	}

	expLegacyQueryRes = &querylegacy.Result{
		Status: querylegacy.Status{
			ElapsedTime:    time.Microsecond * 542_114,
			MinCursor:      "c776x1uafkpu-4918f6cb9000095-0",
			MaxCursor:      "c776x1uafnvq-4918f6cb9000095-1",
			BlocksExamined: 4,
			RowsExamined:   142655,
			RowsMatched:    142655,
			NumGroups:      0,
			IsPartial:      false,
			MinBlockTime:   parseTimeOrPanic("2020-11-19T11:06:31.569475746Z"),
			MaxBlockTime:   parseTimeOrPanic("2020-11-27T12:06:38.966791794Z"),
		},
		Matches: []querylegacy.Entry{
			{
				Time:    parseTimeOrPanic("2020-11-19T11:06:31.569475746Z"),
				SysTime: parseTimeOrPanic("2020-11-19T11:06:31.581384524Z"),
				RowID:   "c776x1uafkpu-4918f6cb9000095-0",
				Data: map[string]any{
					"agent":       "Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)",
					"bytes":       float64(0),
					"referrer":    "-",
					"remote_ip":   "93.180.71.3",
					"remote_user": "-",
					"request":     "GET /downloads/product_1 HTTP/1.1",
					"response":    float64(304),
					"time":        "17/May/2015:08:05:32 +0000",
				},
			},
			{
				Time:    parseTimeOrPanic("2020-11-19T11:06:31.569479846Z"),
				SysTime: parseTimeOrPanic("2020-11-19T11:06:31.581384524Z"),
				RowID:   "c776x1uafnvq-4918f6cb9000095-1",
				Data: map[string]any{
					"agent":       "Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)",
					"bytes":       float64(0),
					"referrer":    "-",
					"remote_ip":   "93.180.71.3",
					"remote_user": "-",
					"request":     "GET /downloads/product_1 HTTP/1.1",
					"response":    float64(304),
					"time":        "17/May/2015:08:05:23 +0000",
				},
			},
		},
		SavedQueryID: "fyTFUldK4Z5219rWaz",
		TraceID:      "abc",
	}
)

func TestDatasetsService_List(t *testing.T) {
	exp := []*Dataset{
		{
			ID:          "test",
			Name:        "test",
			Description: "",
			CreatedBy:   "f83e245a-afdc-47ad-a765-4addd1994321",
			CreatedAt:   testhelper.MustTimeParse(t, time.RFC3339Nano, "2020-11-17T22:29:00.521238198Z"),
			CanWrite:    true,
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `[
			{
				"id": "test",
				"name": "test",
				"who": "f83e245a-afdc-47ad-a765-4addd1994321",
				"created": "2020-11-17T22:29:00.521238198Z",
				"canWrite": true
			}
		]`)
		assert.NoError(t, err)
	}

	client := setup(t, "GET /v2/datasets", hf)

	res, err := client.Datasets.List(context.Background())
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestDatasetsService_Get(t *testing.T) {
	exp := &Dataset{
		ID:          "test",
		Name:        "test",
		Description: "This is a test description",
		CreatedBy:   "f83e245a-afdc-47ad-a765-4addd1994321",
		CreatedAt:   testhelper.MustTimeParse(t, time.RFC3339Nano, "2020-11-17T22:29:00.521238198Z"),
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
			"id": "test",
			"name": "test",
			"description": "This is a test description",
			"who": "f83e245a-afdc-47ad-a765-4addd1994321",
			"created": "2020-11-17T22:29:00.521238198Z"
		}`)
		assert.NoError(t, err)
	}

	client := setup(t, "GET /v2/datasets/test", hf)

	res, err := client.Datasets.Get(context.Background(), "test")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestDatasetsService_Create(t *testing.T) {
	exp := &Dataset{
		ID:          "test",
		Name:        "test",
		Description: "This is a test description",
		CreatedBy:   "f83e245a-afdc-47ad-a765-4addd1994321",
		CreatedAt:   testhelper.MustTimeParse(t, time.RFC3339Nano, "2020-11-18T21:30:20.623322799Z"),
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, mediaTypeJSON, r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
			"id": "test",
			"name": "test",
			"description": "This is a test description",
			"who": "f83e245a-afdc-47ad-a765-4addd1994321",
			"created": "2020-11-18T21:30:20.623322799Z"
		}`)
		assert.NoError(t, err)
	}

	client := setup(t, "POST /v2/datasets", hf)

	res, err := client.Datasets.Create(context.Background(), DatasetCreateRequest{
		Name:        "test",
		Description: "This is a test description",
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestDatasetsService_Update(t *testing.T) {
	exp := &Dataset{
		ID:          "test",
		Name:        "test",
		Description: "This is the new description",
		CreatedBy:   "f83e245a-afdc-47ad-a765-4addd1994321",
		CreatedAt:   testhelper.MustTimeParse(t, time.RFC3339Nano, "2020-11-18T21:30:20.623322799Z"),
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, mediaTypeJSON, r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
			"id": "test",
			"name": "test",
			"description": "This is the new description",
			"who": "f83e245a-afdc-47ad-a765-4addd1994321",
			"created": "2020-11-18T21:30:20.623322799Z"
		}`)
		assert.NoError(t, err)
	}

	client := setup(t, "PUT /v2/datasets/test", hf)

	res, err := client.Datasets.Update(context.Background(), "test", DatasetUpdateRequest{
		Description: "This is the new description",
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestDatasetsService_Delete(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}

	client := setup(t, "DELETE /v2/datasets/test", hf)

	err := client.Datasets.Delete(context.Background(), "test")
	require.NoError(t, err)
}

func TestDatasetsService_Trim(t *testing.T) {
	hf := func(_ http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
	}

	client := setup(t, "POST /v2/datasets/test/trim", hf)

	err := client.Datasets.Trim(context.Background(), "test", time.Hour)
	require.NoError(t, err)
}

// TestDatasetsService_Ingest tests the ingest functionality of the client. It
// also tests the event labels functionality by setting two individual labels.
func TestDatasetsService_Ingest(t *testing.T) {
	exp := &ingest.Status{
		Ingested:       2,
		Failed:         0,
		Failures:       []*ingest.Failure{},
		ProcessedBytes: 630,
		BlocksCreated:  0,
		WALLength:      2,
		TraceID:        "abc",
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, mediaTypeJSON, r.Header.Get("Content-Type"))

		assert.Equal(t, "time", r.URL.Query().Get("timestamp-field"))
		assert.Equal(t, "2/Jan/2006:15:04:05 +0000", r.URL.Query().Get("timestamp-format"))
		assert.Equal(t, ";", r.URL.Query().Get("csv-delimiter"))
		eventLabels := assertValidJSON(t, strings.NewReader(r.Header.Get("X-Axiom-Event-Labels")))
		assert.Equal(t, "eu-west-1", eventLabels[0].(map[string]any)["region"])
		assert.EqualValues(t, 1, eventLabels[0].(map[string]any)["instance"])

		_ = assertValidJSON(t, r.Body)

		w.Header().Set("Content-Type", mediaTypeJSON)
		w.Header().Set("X-Axiom-Trace-Id", "abc")
		_, err := fmt.Fprint(w, `{
			"ingested": 2,
			"failed": 0,
			"failures": [],
			"processedBytes": 630,
			"blocksCreated": 0,
			"walLength": 2
		}`)
		assert.NoError(t, err)
	}

	client := setup(t, "POST /v1/datasets/test/ingest", hf)

	r := strings.NewReader(`[
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
	]`)

	res, err := client.Datasets.Ingest(context.Background(), "test", r, JSON, Identity,
		ingest.SetTimestampField("time"),
		ingest.SetTimestampFormat("2/Jan/2006:15:04:05 +0000"),
		ingest.SetCSVDelimiter(";"), // Obviously not valid for JSON, but perfectly fine to test for its presence in this test.
		ingest.SetEventLabel("region", "eu-west-1"),
		ingest.SetEventLabel("instance", 1),
	)
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

// TestDatasetsService_IngestEvents tests the ingest functionality of the
// client. It also tests the event labels functionality by setting a set of
// labels.
func TestDatasetsService_IngestEvents(t *testing.T) {
	exp := &ingest.Status{
		Ingested:       2,
		Failed:         0,
		Failures:       []*ingest.Failure{},
		ProcessedBytes: 630,
		BlocksCreated:  0,
		WALLength:      2,
		TraceID:        "abc",
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, mediaTypeNDJSON, r.Header.Get("Content-Type"))
		assert.Equal(t, "zstd", r.Header.Get("Content-Encoding"))
		eventLabels := assertValidJSON(t, strings.NewReader(r.Header.Get("X-Axiom-Event-Labels")))
		assert.Equal(t, "eu-west-1", eventLabels[0].(map[string]any)["region"])
		assert.EqualValues(t, 1, eventLabels[0].(map[string]any)["instance"])

		zsr, err := zstd.NewReader(r.Body)
		require.NoError(t, err)

		events := assertValidJSON(t, zsr)
		assert.Len(t, events, 2)
		zsr.Close()

		w.Header().Set("Content-Type", mediaTypeJSON)
		w.Header().Set("X-Axiom-Trace-Id", "abc")
		_, err = fmt.Fprint(w, `{
			"ingested": 2,
			"failed": 0,
			"failures": [],
			"processedBytes": 630,
			"blocksCreated": 0,
			"walLength": 2
		}`)
		assert.NoError(t, err)
	}

	client := setup(t, "POST /v1/datasets/test/ingest", hf)

	events := []Event{
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

	res, err := client.Datasets.IngestEvents(context.Background(), "test", events, ingest.SetEventLabels(
		map[string]any{
			"region":   "eu-west-1",
			"instance": 1,
		},
	))
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

// TestDatasetsService_IngestEvents_Retry tests the retry ingest functionality
// of the client. It also tests the event labels functionality by setting no
// labels. It also tests for the presence of a trace ID in the response.
func TestDatasetsService_IngestEvents_Retry(t *testing.T) {
	exp := &ingest.Status{
		Ingested:       2,
		Failed:         0,
		Failures:       []*ingest.Failure{},
		ProcessedBytes: 630,
		BlocksCreated:  0,
		WALLength:      2,
		TraceID:        "abc",
	}

	hasErrored := false
	hf := func(w http.ResponseWriter, r *http.Request) {
		if !hasErrored {
			hasErrored = true
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, mediaTypeNDJSON, r.Header.Get("Content-Type"))
		assert.Equal(t, "zstd", r.Header.Get("Content-Encoding"))
		assert.Empty(t, r.Header.Get(headerEventLabels))

		zsr, err := zstd.NewReader(r.Body)
		require.NoError(t, err)

		events := assertValidJSON(t, zsr)
		assert.Len(t, events, 2)
		zsr.Close()

		w.Header().Set("Content-Type", mediaTypeJSON)
		w.Header().Set("X-Axiom-Trace-Id", "abc")
		_, err = fmt.Fprint(w, `{
			"ingested": 2,
			"failed": 0,
			"failures": [],
			"processedBytes": 630,
			"blocksCreated": 0,
			"walLength": 2
		}`)
		assert.NoError(t, err)
	}

	client := setup(t, "POST /v1/datasets/test/ingest", hf)

	events := []Event{
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

	res, err := client.Datasets.IngestEvents(context.Background(), "test", events)
	require.NoError(t, err)

	assert.Equal(t, exp, res)
	assert.True(t, hasErrored)
}

func TestDatasetsService_IngestChannel_Unbuffered(t *testing.T) {
	exp := &ingest.Status{
		Ingested:       2,
		Failed:         0,
		ProcessedBytes: 630,
		BlocksCreated:  0,
		WALLength:      2,
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, mediaTypeNDJSON, r.Header.Get("Content-Type"))
		assert.Equal(t, "zstd", r.Header.Get("Content-Encoding"))

		zsr, err := zstd.NewReader(r.Body)
		require.NoError(t, err)

		events := assertValidJSON(t, zsr)
		assert.Len(t, events, 2)
		zsr.Close()

		w.Header().Set("Content-Type", mediaTypeJSON)
		w.Header().Set("X-Axiom-Trace-Id", "abc")
		_, err = fmt.Fprint(w, `{
			"ingested": 2,
			"failed": 0,
			"failures": [],
			"processedBytes": 630,
			"blocksCreated": 0,
			"walLength": 2
		}`)
		assert.NoError(t, err)
	}

	client := setup(t, "POST /v1/datasets/test/ingest", hf)

	events := []Event{
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

	eventCh := make(chan Event)
	go func() {
		for _, e := range events {
			eventCh <- e
		}
		close(eventCh)
	}()

	res, err := client.Datasets.IngestChannel(context.Background(), "test", eventCh)
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestDatasetsService_IngestChannel_Buffered(t *testing.T) {
	exp := &ingest.Status{
		Ingested:       2,
		Failed:         0,
		ProcessedBytes: 1260,
		BlocksCreated:  0,
		WALLength:      2,
	}

	handlerInvokedCount := 0
	hf := func(w http.ResponseWriter, r *http.Request) {
		handlerInvokedCount++

		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, mediaTypeNDJSON, r.Header.Get("Content-Type"))
		assert.Equal(t, "zstd", r.Header.Get("Content-Encoding"))

		zsr, err := zstd.NewReader(r.Body)
		require.NoError(t, err)

		events := assertValidJSON(t, zsr)
		assert.Len(t, events, 1)
		zsr.Close()

		// For the sake of simplicity in this handler, we'll just return the
		// same WAL length for each request.
		w.Header().Set("Content-Type", mediaTypeJSON)
		w.Header().Set("X-Axiom-Trace-Id", "abc")
		_, err = fmt.Fprintf(w, `{
			"ingested": %d,
			"failed": 0,
			"failures": [],
			"processedBytes": %d,
			"blocksCreated": 0,
			"walLength": 2
		}`, len(events), len(events)*630)
		assert.NoError(t, err)
	}

	client := setup(t, "POST /v1/datasets/test/ingest", hf)

	events := []Event{
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

	eventCh := make(chan Event, 1)
	go func() {
		for _, e := range events {
			eventCh <- e
		}
		close(eventCh)
	}()

	res, err := client.Datasets.IngestChannel(context.Background(), "test", eventCh)
	require.NoError(t, err)

	assert.Equal(t, exp, res)
	assert.Equal(t, 2, handlerInvokedCount)
}

func TestDatasetsService_IngestChannel_UnbufferedSlow(t *testing.T) {
	exp := &ingest.Status{
		Ingested:       2,
		Failed:         0,
		ProcessedBytes: 1260,
		BlocksCreated:  0,
		WALLength:      2,
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, mediaTypeNDJSON, r.Header.Get("Content-Type"))
		assert.Equal(t, "zstd", r.Header.Get("Content-Encoding"))

		zsr, err := zstd.NewReader(r.Body)
		require.NoError(t, err)

		events := assertValidJSON(t, zsr)
		assert.Len(t, events, 1)
		zsr.Close()

		// For the sake of simplicity in this handler, we'll just return the
		// same WAL length for each request.
		w.Header().Set("Content-Type", mediaTypeJSON)
		w.Header().Set("X-Axiom-Trace-Id", "abc")
		_, err = fmt.Fprintf(w, `{
			"ingested": %d,
			"failed": 0,
			"failures": [],
			"processedBytes": %d,
			"blocksCreated": 0,
			"walLength": 2
		}`, len(events), len(events)*630)
		assert.NoError(t, err)
	}

	client := setup(t, "POST /v1/datasets/test/ingest", hf)

	events := []Event{
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

	eventCh := make(chan Event)
	go func() {
		for _, e := range events {
			eventCh <- e
			// Simulate a slow producer which should trigger the ticker based
			// batch flush in the IngestChannel method.
			time.Sleep(time.Second + time.Millisecond*250)
		}
		close(eventCh)
	}()

	res, err := client.Datasets.IngestChannel(context.Background(), "test", eventCh)
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestDatasetsService_IngestChannel_BufferedSlow(t *testing.T) {
	exp := &ingest.Status{
		Ingested:       2,
		Failed:         0,
		ProcessedBytes: 1260,
		BlocksCreated:  0,
		WALLength:      2,
	}

	handlerInvokedCount := 0
	hf := func(w http.ResponseWriter, r *http.Request) {
		handlerInvokedCount++

		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, mediaTypeNDJSON, r.Header.Get("Content-Type"))
		assert.Equal(t, "zstd", r.Header.Get("Content-Encoding"))

		zsr, err := zstd.NewReader(r.Body)
		require.NoError(t, err)

		events := assertValidJSON(t, zsr)
		assert.Len(t, events, 1)
		zsr.Close()

		// For the sake of simplicity in this handler, we'll just return the
		// same WAL length for each request.
		w.Header().Set("Content-Type", mediaTypeJSON)
		w.Header().Set("X-Axiom-Trace-Id", "abc")
		_, err = fmt.Fprintf(w, `{
			"ingested": %d,
			"failed": 0,
			"failures": [],
			"processedBytes": %d,
			"blocksCreated": 0,
			"walLength": 2
		}`, len(events), len(events)*630)
		assert.NoError(t, err)
	}

	client := setup(t, "POST /v1/datasets/test/ingest", hf)

	events := []Event{
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

	eventCh := make(chan Event, 2)
	go func() {
		for _, e := range events {
			eventCh <- e
			// Simulate a slow producer which should trigger the ticker based
			// batch flush in the IngestChannel method.
			time.Sleep(time.Second + time.Millisecond*250)
		}
		close(eventCh)
	}()

	res, err := client.Datasets.IngestChannel(context.Background(), "test", eventCh)
	require.NoError(t, err)

	assert.Equal(t, exp, res)
	assert.Equal(t, 2, handlerInvokedCount)
}

// TODO(lukasmalkmus): Write an ingest test that contains some failures in the
// server response.

func TestDatasetsService_Query(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, mediaTypeJSON, r.Header.Get("Content-Type"))

		assert.Equal(t, "tabular", r.URL.Query().Get("format"))

		var req aplQueryRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if assert.NoError(t, err) {
			assert.EqualValues(t, "['test'] | where response == 304", req.APL)
			assert.NotEmpty(t, req.StartTime)
			assert.Empty(t, req.EndTime)
		}

		w.Header().Set("Content-Type", mediaTypeJSON)
		w.Header().Set("X-Axiom-Trace-Id", "abc")
		_, err = fmt.Fprint(w, actQueryResp)
		assert.NoError(t, err)
	}

	client := setup(t, "POST /v2/datasets/query", hf)

	res, err := client.Datasets.Query(context.Background(),
		"['test'] | where response == 304",
		query.SetStartTime(time.Now().Add(-time.Minute*5)),
	)
	require.NoError(t, err)

	assert.Equal(t, expQueryRes, res)
}

// TODO(lukasmalkmus): Add test for a query with an aggregation.

func TestDatasetsService_QueryLegacy(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, mediaTypeJSON, r.Header.Get("Content-Type"))

		assert.Equal(t, "1s", r.URL.Query().Get("streaming-duration"))
		assert.Equal(t, "true", r.URL.Query().Get("nocache"))
		assert.Equal(t, "analytics", r.URL.Query().Get("saveAsKind"))

		w.Header().Set("X-Axiom-History-Query-Id", "fyTFUldK4Z5219rWaz")

		w.Header().Set("Content-Type", mediaTypeJSON)
		w.Header().Set("X-Axiom-Trace-Id", "abc")
		_, err := fmt.Fprint(w, actLegacyQueryResp)
		assert.NoError(t, err)
	}

	client := setup(t, "POST /v1/datasets/test/query", hf)

	res, err := client.Datasets.QueryLegacy(context.Background(), "test", querylegacy.Query{
		StartTime: testhelper.MustTimeParse(t, time.RFC3339Nano, "2020-11-26T11:18:00Z"),
		EndTime:   testhelper.MustTimeParse(t, time.RFC3339Nano, "2020-11-17T11:18:00Z"),
	}, querylegacy.Options{
		StreamingDuration: time.Second,
		NoCache:           true,
		SaveKind:          querylegacy.Analytics,
	})
	require.NoError(t, err)

	assert.Equal(t, expLegacyQueryRes, res)
}

func TestDatasetsService_QueryLegacyInvalid_InvalidSaveKind(t *testing.T) {
	client := setup(t, "POST /v1/datasets/test/query", nil)

	_, err := client.Datasets.QueryLegacy(context.Background(), "test", querylegacy.Query{}, querylegacy.Options{
		SaveKind: querylegacy.APL,
	})
	require.EqualError(t, err, `invalid query kind "apl": must be "analytics" or "stream"`)
}

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ContentType
		wantErr string
	}{
		// FIXME(lukasmalkmus): The function incorrectly detects the content
		// type as NDJSON as it strips newline characters from the input.
		// {
		// 	name:  "json - pretty",
		// 	input: "{\n\"a\":\"b\"\n}",
		// 	want:  JSON,
		// },
		{
			name:  "json - multiline",
			input: `[{"a":"b"}, {"c":"d"}]`,
			want:  JSON,
		},
		{
			name:  "ndjson - single line",
			input: `{"a":"b"}`,
			want:  NDJSON,
		},
		{
			name: "ndjson - single line",
			input: `{"a":"b"}
				{"c":"d"}`,
			want: NDJSON,
		},
		{
			name: CSV.String(),
			input: `Year,Make,Model,Length
				1997,Ford,E350,2.35
				2000,Mercury,Cougar,2.38`,
			want: CSV,
		},
		{
			name:    "eof",
			input:   "",
			wantErr: "couldn't find beginning of supported ingestion format",
		},
		{
			name:    "invalid",
			input:   "123",
			wantErr: "cannot determine content type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, got, err := DetectContentType(strings.NewReader(tt.input))
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			if b, err := io.ReadAll(r); assert.NoError(t, err) {
				assert.Equal(t, tt.input, string(b))
			}
			assert.Equal(t, tt.want.String(), got.String())
		})
	}
}

func assertValidJSON(t *testing.T, r io.Reader) []any {
	var (
		dec = json.NewDecoder(r)
		vs  []any
	)
	for dec.More() {
		var v any
		err := dec.Decode(&v)
		assert.NoError(t, err)
		assert.NotEmpty(t, v)
		vs = append(vs, v)
	}
	return vs
}

func parseTimeOrPanic(value string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		panic(err)
	}
	return t
}
