package axiom

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/axiomhq/axiom-go/axiom/query"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatasetsService_Stats(t *testing.T) {
	exp := &DatasetStats{
		Datasets: []*DatasetInfo{
			{
				Name:                 "test",
				NumBlocks:            1,
				NumEvents:            68459,
				NumFields:            8,
				InputBytes:           10383386,
				InputBytesHuman:      "10 MB",
				CompressedBytes:      2509224,
				CompressedBytesHuman: "2.5 MB",
				MinTime:              mustTimeParse(t, time.RFC3339, "2020-11-17T22:30:59Z"),
				MaxTime:              mustTimeParse(t, time.RFC3339, "2020-11-18T17:31:55Z"),
				Fields: []Field{
					{
						Name: "_sysTime",
						Type: "integer",
					},
					{
						Name: "_time",
						Type: "integer",
					},
					{
						Name: "path",
						Type: "string",
					},
					{
						Name: "size",
						Type: "integer",
					},
					{
						Name: "status",
						Type: "integer",
					},
				},
				Created: mustTimeParse(t, time.RFC3339Nano, "2020-11-18T21:30:20.623322799Z"),
			},
			{
				Name:                 "test1",
				NumBlocks:            1,
				NumEvents:            68459,
				NumFields:            8,
				InputBytes:           10383386,
				InputBytesHuman:      "10 MB",
				CompressedBytes:      2509224,
				CompressedBytesHuman: "2.5 MB",
				MinTime:              mustTimeParse(t, time.RFC3339, "2020-11-17T22:30:59Z"),
				MaxTime:              mustTimeParse(t, time.RFC3339, "2020-11-18T17:31:55Z"),
				Fields: []Field{
					{
						Name: "_sysTime",
						Type: "integer",
					},
					{
						Name: "_time",
						Type: "integer",
					},
					{
						Name: "path",
						Type: "string",
					},
					{
						Name: "size",
						Type: "integer",
					},
					{
						Name: "status",
						Type: "integer",
					},
				},
				Created: mustTimeParse(t, time.RFC3339Nano, "2020-11-18T21:30:20.623322799Z"),
			},
		},
		NumBlocks:            2,
		NumEvents:            136918,
		InputBytes:           666337356,
		InputBytesHuman:      "666 MB",
		CompressedBytes:      19049348,
		CompressedBytesHuman: "19 MB",
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `{
			"datasets": [
				{
					"name": "test",
					"numBlocks": 1,
					"numEvents": 68459,
					"numFields": 8,
					"inputBytes": 10383386,
					"inputBytesHuman": "10 MB",
					"compressedBytes": 2509224,
					"compressedBytesHuman": "2.5 MB",
					"minTime": "2020-11-17T22:30:59Z",
					"maxTime": "2020-11-18T17:31:55Z",
					"fields": [
						{
							"name": "_sysTime",
							"type": "integer"
						},
						{
							"name": "_time",
							"type": "integer"
						},
						{
							"name": "path",
							"type": "string"
						},
						{
							"name": "size",
							"type": "integer"
						},
						{
							"name": "status",
							"type": "integer"
						}
					],
					"created": "2020-11-18T21:30:20.623322799Z"
				},
				{
					"name": "test1",
					"numBlocks": 1,
					"numEvents": 68459,
					"numFields": 8,
					"inputBytes": 10383386,
					"inputBytesHuman": "10 MB",
					"compressedBytes": 2509224,
					"compressedBytesHuman": "2.5 MB",
					"minTime": "2020-11-17T22:30:59Z",
					"maxTime": "2020-11-18T17:31:55Z",
					"fields": [
						{
							"name": "_sysTime",
							"type": "integer"
						},
						{
							"name": "_time",
							"type": "integer"
						},
						{
							"name": "path",
							"type": "string"
						},
						{
							"name": "size",
							"type": "integer"
						},
						{
							"name": "status",
							"type": "integer"
						}
					],
					"created": "2020-11-18T21:30:20.623322799Z"
				}
			],
			"numBlocks": 2,
			"numEvents": 136918,
			"inputBytes": 666337356,
			"inputBytesHuman": "666 MB",
			"compressedBytes": 19049348,
			"compressedBytesHuman": "19 MB"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/datasets/_stats", hf)
	defer teardown()

	res, err := client.Datasets.Stats(context.Background())
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestDatasetsService_List(t *testing.T) {
	exp := []*Dataset{
		{
			ID:          "test",
			Name:        "test",
			Description: "",
			Created:     mustTimeParse(t, time.RFC3339Nano, "2020-11-17T22:29:00.521238198Z"),
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `[
			{
				"id": "test",
				"name": "test",
				"created": "2020-11-17T22:29:00.521238198Z"
			}
		]`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/datasets", hf)
	defer teardown()

	res, err := client.Datasets.List(context.Background())
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestDatasetsService_Get(t *testing.T) {
	exp := &Dataset{
		ID:          "test",
		Name:        "test",
		Description: "This is a test description",
		Created:     mustTimeParse(t, time.RFC3339Nano, "2020-11-17T22:29:00.521238198Z"),
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `{
			"id": "test",
			"name": "test",
			"description": "This is a test description",
			"created": "2020-11-17T22:29:00.521238198Z"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/datasets/test", hf)
	defer teardown()

	res, err := client.Datasets.Get(context.Background(), "test")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestDatasetsService_Create(t *testing.T) {
	exp := &Dataset{
		ID:          "test",
		Name:        "test",
		Description: "This is a test description",
		Created:     mustTimeParse(t, time.RFC3339Nano, "2020-11-18T21:30:20.623322799Z"),
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		_, err := fmt.Fprint(w, `{
			"id": "test",
			"name": "test",
			"description": "This is a test description",
			"created": "2020-11-18T21:30:20.623322799Z"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/datasets", hf)
	defer teardown()

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
		Created:     mustTimeParse(t, time.RFC3339Nano, "2020-11-18T21:30:20.623322799Z"),
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		_, err := fmt.Fprint(w, `{
			"id": "test",
			"name": "test",
			"description": "This is the new description",
			"created": "2020-11-18T21:30:20.623322799Z"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/datasets/test", hf)
	defer teardown()

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

	client, teardown := setup(t, "/api/v1/datasets/test", hf)
	defer teardown()

	err := client.Datasets.Delete(context.Background(), "test")
	require.NoError(t, err)
}

func TestDatasetsService_Info(t *testing.T) {
	exp := &DatasetInfo{
		Name:                 "test",
		NumBlocks:            1,
		NumEvents:            68459,
		NumFields:            8,
		InputBytes:           10383386,
		InputBytesHuman:      "10 MB",
		CompressedBytes:      2509224,
		CompressedBytesHuman: "2.5 MB",
		MinTime:              mustTimeParse(t, time.RFC3339, "2020-11-17T22:30:59Z"),
		MaxTime:              mustTimeParse(t, time.RFC3339, "2020-11-18T17:31:55Z"),
		Fields: []Field{
			{
				Name: "_sysTime",
				Type: "integer",
			},
			{
				Name: "_time",
				Type: "integer",
			},
			{
				Name: "path",
				Type: "string",
			},
			{
				Name: "size",
				Type: "integer",
			},
			{
				Name: "status",
				Type: "integer",
			},
		},
		Created: mustTimeParse(t, time.RFC3339Nano, "2020-11-18T21:30:20.623322799Z"),
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `{
			"name": "test",
			"numBlocks": 1,
			"numEvents": 68459,
			"numFields": 8,
			"inputBytes": 10383386,
			"inputBytesHuman": "10 MB",
			"compressedBytes": 2509224,
			"compressedBytesHuman": "2.5 MB",
			"minTime": "2020-11-17T22:30:59Z",
			"maxTime": "2020-11-18T17:31:55Z",
			"fields": [
				{
					"name": "_sysTime",
					"type": "integer"
				},
				{
					"name": "_time",
					"type": "integer"
				},
				{
					"name": "path",
					"type": "string"
				},
				{
					"name": "size",
					"type": "integer"
				},
				{
					"name": "status",
					"type": "integer"
				}
			],
			"created": "2020-11-18T21:30:20.623322799Z"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/datasets/test/info", hf)
	defer teardown()

	res, err := client.Datasets.Info(context.Background(), "test")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestDatasetsService_History(t *testing.T) {
	exp := &HistoryQuery{
		ID:      "GHP2ufS7OYwMeBhXHj",
		Kind:    "analytics",
		Dataset: "test",
		Owner:   "f83e245a-afdc-47ad-a765-4addd1994333",
		Query: query.Query{
			StartTime: mustTimeParse(t, time.RFC3339, "2020-11-18T13:00:00.000Z"),
			EndTime:   mustTimeParse(t, time.RFC3339, "2020-11-25T14:00:00.000Z"),
			Limit:     100,
		},
		Created: mustTimeParse(t, time.RFC3339, "2020-12-08T13:28:52.78954814Z"),
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `{
			"created": "2020-12-08T13:28:52.78954814Z",
			"dataset": "test",
			"id": "GHP2ufS7OYwMeBhXHj",
			"kind": "analytics",
			"query": {
				"startTime": "2020-11-18T13:00:00.000Z",
				"endTime": "2020-11-25T14:00:00.000Z",
				"limit": 100
			},
			"who": "f83e245a-afdc-47ad-a765-4addd1994333"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/datasets/_history/GHP2ufS7OYwMeBhXHj", hf)
	defer teardown()

	res, err := client.Datasets.History(context.Background(), "GHP2ufS7OYwMeBhXHj")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}
func TestDatasetsService_Ingest(t *testing.T) {
	exp := &IngestStatus{
		Ingested:       2,
		Failed:         0,
		Failures:       []*IngestFailure{},
		ProcessedBytes: 630,
		BlocksCreated:  0,
		WALLength:      2,
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

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

	client, teardown := setup(t, "/api/v1/datasets/test/ingest", hf)
	defer teardown()

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

	res, err := client.Datasets.Ingest(context.Background(), "test", r, JSON, Identity, IngestOptions{})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestDatasetsService_IngestEvents(t *testing.T) {
	exp := &IngestStatus{
		Ingested:       2,
		Failed:         0,
		Failures:       []*IngestFailure{},
		ProcessedBytes: 630,
		BlocksCreated:  0,
		WALLength:      2,
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/x-ndjson", r.Header.Get("content-type"))
		assert.Equal(t, "gzip", r.Header.Get("content-encoding"))

		gzr, err := gzip.NewReader(r.Body)
		require.NoError(t, err)

		assertValidJSON(t, gzr)
		assert.NoError(t, gzr.Close())

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

	client, teardown := setup(t, "/api/v1/datasets/test/ingest", hf)
	defer teardown()

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

	res, err := client.Datasets.IngestEvents(context.Background(), "test", IngestOptions{}, events...)
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

// TODO(lukasmalkmus): Write an ingest test that contains some failures in the
// server response.

func TestDatasetsService_Query(t *testing.T) {
	exp := &query.Result{
		Status: query.Status{
			ElapsedTime:    542114 * time.Microsecond,
			BlocksExamined: 4,
			RowsExamined:   142655,
			RowsMatched:    142655,
			NumGroups:      0,
			IsPartial:      false,
			MinBlockTime:   mustTimeParse(t, time.RFC3339Nano, "2020-11-19T11:06:31.569475746Z"),
			MaxBlockTime:   mustTimeParse(t, time.RFC3339Nano, "2020-11-27T12:06:38.966791794Z"),
		},
		Matches: []query.Entry{
			{
				Time:    mustTimeParse(t, time.RFC3339Nano, "2020-11-19T11:06:31.569475746Z"),
				SysTime: mustTimeParse(t, time.RFC3339Nano, "2020-11-19T11:06:31.581384524Z"),
				RowID:   "c776x1uafkpu-4918f6cb9000095-0",
				Data: map[string]interface{}{
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
				Time:    mustTimeParse(t, time.RFC3339Nano, "2020-11-19T11:06:31.569479846Z"),
				SysTime: mustTimeParse(t, time.RFC3339Nano, "2020-11-19T11:06:31.581384524Z"),
				RowID:   "c776x1uafnvq-4918f6cb9000095-1",
				Data: map[string]interface{}{
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
		Buckets: query.Timeseries{
			Series: []query.Interval{},
			Totals: []query.EntryGroup{},
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		assert.Equal(t, "1s", r.URL.Query().Get("streaming-duration"))
		assert.Equal(t, "true", r.URL.Query().Get("no-cache"))

		_, err := fmt.Fprint(w, `{
			"status": {
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
			],
			"buckets": {
				"series": [],
				"totals": []
			}
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/datasets/test/query", hf)
	defer teardown()

	res, err := client.Datasets.Query(context.Background(), "test", query.Query{
		StartTime: mustTimeParse(t, time.RFC3339Nano, "2020-11-26T11:18:00Z"),
		EndTime:   mustTimeParse(t, time.RFC3339Nano, "2020-11-17T11:18:00Z"),
	}, query.Options{
		StreamingDuration: time.Second,
		NoCache:           true,
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestGZIPStreamer(t *testing.T) {
	exp := "Some fox jumps over a fence."

	r, err := GZIPStreamer(strings.NewReader(exp), gzip.BestSpeed)
	require.NoError(t, err)

	gzr, err := gzip.NewReader(r)
	require.NoError(t, err)
	defer func() {
		closeErr := gzr.Close()
		require.NoError(t, closeErr)
	}()

	act, err := ioutil.ReadAll(gzr)
	require.NoError(t, err)

	assert.Equal(t, exp, string(act))
}

func assertValidJSON(t *testing.T, r io.Reader) bool {
	dec := json.NewDecoder(r)
	for dec.More() {
		var v interface{}
		err := dec.Decode(&v)
		if !assert.NoError(t, err) {
			return false
		} else if !assert.NotEmpty(t, v) {
			return false
		}
	}

	return true
}
