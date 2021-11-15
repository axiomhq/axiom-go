package axiom

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/axiom/apl"
	"github.com/axiomhq/axiom-go/axiom/query"
)

const actQueryResp = `{
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
	}`

const actAPLQueryResp = `{
		"request": {
			"startTime": "2021-07-20T16:34:57.911170243Z",
			"endTime": "2021-08-19T16:34:57.885821616Z",
			"resolution": "",
			"aggregations": null,
			"groupBy": null,
			"order": null,
			"limit": 1000,
			"virtualFields": null,
			"project": null,
			"cursor": "",
			"includeCursor": false
		},
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
	}`

var expQueryRes = &query.Result{
	Status: query.Status{
		ElapsedTime:    542114 * time.Microsecond,
		BlocksExamined: 4,
		RowsExamined:   142655,
		RowsMatched:    142655,
		NumGroups:      0,
		IsPartial:      false,
		MinBlockTime:   parseTimeOrPanic(time.RFC3339Nano, "2020-11-19T11:06:31.569475746Z"),
		MaxBlockTime:   parseTimeOrPanic(time.RFC3339Nano, "2020-11-27T12:06:38.966791794Z"),
	},
	Matches: []query.Entry{
		{
			Time:    parseTimeOrPanic(time.RFC3339Nano, "2020-11-19T11:06:31.569475746Z"),
			SysTime: parseTimeOrPanic(time.RFC3339Nano, "2020-11-19T11:06:31.581384524Z"),
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
			Time:    parseTimeOrPanic(time.RFC3339Nano, "2020-11-19T11:06:31.569479846Z"),
			SysTime: parseTimeOrPanic(time.RFC3339Nano, "2020-11-19T11:06:31.581384524Z"),
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
	SavedQueryID: "fyTFUldK4Z5219rWaz",
}

var expAPLQueryRes = &apl.Result{
	Request: &query.Query{
		StartTime: parseTimeOrPanic(time.RFC3339Nano, "2021-07-20T16:34:57.911170243Z"),
		EndTime:   parseTimeOrPanic(time.RFC3339Nano, "2021-08-19T16:34:57.885821616Z"),
		Limit:     1000,
	},
	Result: expQueryRes,
}

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
						Name:        "_sysTime",
						Description: "",
						Type:        "integer",
						Unit:        "",
						Hidden:      false,
					},
					{
						Name:        "_time",
						Description: "",
						Type:        "integer",
						Unit:        "",
						Hidden:      false,
					},
					{
						Name:        "path",
						Description: "",
						Type:        "string",
						Unit:        "",
						Hidden:      false,
					},
					{
						Name:        "size",
						Description: "",
						Type:        "integer",
						Unit:        "",
						Hidden:      false,
					},
					{
						Name:        "status",
						Description: "",
						Type:        "integer",
						Unit:        "",
						Hidden:      false,
					},
				},
				CreatedBy: "f83e245a-afdc-47ad-a765-4addd1994321",
				CreatedAt: mustTimeParse(t, time.RFC3339Nano, "2020-11-18T21:30:20.623322799Z"),
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
						Name:        "_sysTime",
						Description: "",
						Type:        "integer",
						Unit:        "",
						Hidden:      false,
					},
					{
						Name:        "_time",
						Description: "",
						Type:        "integer",
						Unit:        "",
						Hidden:      false,
					},
					{
						Name:        "path",
						Description: "",
						Type:        "string",
						Unit:        "",
						Hidden:      false,
					},
					{
						Name:        "size",
						Description: "",
						Type:        "integer",
						Unit:        "",
						Hidden:      false,
					},
					{
						Name:        "status",
						Description: "",
						Type:        "integer",
						Unit:        "",
						Hidden:      false,
					},
				},
				CreatedBy: "f83e245a-afdc-47ad-a765-4addd1994321",
				CreatedAt: mustTimeParse(t, time.RFC3339Nano, "2020-11-18T21:30:20.623322799Z"),
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
							"type": "integer",
							"unit": "",
            				"hidden": false,
            				"description": ""
						},
						{
							"name": "_time",
							"type": "integer",
							"unit": "",
            				"hidden": false,
            				"description": ""
						},
						{
							"name": "path",
							"type": "string",
							"unit": "",
            				"hidden": false,
            				"description": ""
						},
						{
							"name": "size",
							"type": "integer",
							"unit": "",
            				"hidden": false,
            				"description": ""
						},
						{
							"name": "status",
							"type": "integer",
							"unit": "",
            				"hidden": false,
            				"description": ""
						}
					],
					"who": "f83e245a-afdc-47ad-a765-4addd1994321",
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
							"type": "integer",
							"unit": "",
            				"hidden": false,
            				"description": ""
						},
						{
							"name": "_time",
							"type": "integer",
							"unit": "",
            				"hidden": false,
            				"description": ""
						},
						{
							"name": "path",
							"type": "string",
							"unit": "",
            				"hidden": false,
            				"description": ""
						},
						{
							"name": "size",
							"type": "integer",
							"unit": "",
            				"hidden": false,
            				"description": ""
						},
						{
							"name": "status",
							"type": "integer",
							"unit": "",
            				"hidden": false,
            				"description": ""
						}
					],
					"who": "f83e245a-afdc-47ad-a765-4addd1994321",
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
			CreatedBy:   "f83e245a-afdc-47ad-a765-4addd1994321",
			CreatedAt:   mustTimeParse(t, time.RFC3339Nano, "2020-11-17T22:29:00.521238198Z"),
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `[
			{
				"id": "test",
				"name": "test",
				"who": "f83e245a-afdc-47ad-a765-4addd1994321",
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
		CreatedBy:   "f83e245a-afdc-47ad-a765-4addd1994321",
		CreatedAt:   mustTimeParse(t, time.RFC3339Nano, "2020-11-17T22:29:00.521238198Z"),
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `{
			"id": "test",
			"name": "test",
			"description": "This is a test description",
			"who": "f83e245a-afdc-47ad-a765-4addd1994321",
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
		CreatedBy:   "f83e245a-afdc-47ad-a765-4addd1994321",
		CreatedAt:   mustTimeParse(t, time.RFC3339Nano, "2020-11-18T21:30:20.623322799Z"),
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		_, err := fmt.Fprint(w, `{
			"id": "test",
			"name": "test",
			"description": "This is a test description",
			"who": "f83e245a-afdc-47ad-a765-4addd1994321",
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
		CreatedBy:   "f83e245a-afdc-47ad-a765-4addd1994321",
		CreatedAt:   mustTimeParse(t, time.RFC3339Nano, "2020-11-18T21:30:20.623322799Z"),
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		_, err := fmt.Fprint(w, `{
			"id": "test",
			"name": "test",
			"description": "This is the new description",
			"who": "f83e245a-afdc-47ad-a765-4addd1994321",
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
				Name:        "_sysTime",
				Description: "",
				Type:        "integer",
				Unit:        "",
				Hidden:      false,
			},
			{
				Name:        "_time",
				Description: "",
				Type:        "integer",
				Unit:        "",
				Hidden:      false,
			},
			{
				Name:        "path",
				Description: "",
				Type:        "string",
				Unit:        "",
				Hidden:      false,
			},
			{
				Name:        "size",
				Description: "",
				Type:        "integer",
				Unit:        "",
				Hidden:      false,
			},
			{
				Name:        "status",
				Description: "",
				Type:        "integer",
				Unit:        "",
				Hidden:      false,
			},
		},
		CreatedBy: "f83e245a-afdc-47ad-a765-4addd1994321",
		CreatedAt: mustTimeParse(t, time.RFC3339Nano, "2020-11-18T21:30:20.623322799Z"),
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
					"type": "integer",
            		"unit": "",
            		"hidden": false,
            		"description": ""
				},
				{
					"name": "_time",
					"type": "integer",
            		"unit": "",
            		"hidden": false,
            		"description": ""
				},
				{
					"name": "path",
					"type": "string",
            		"unit": "",
            		"hidden": false,
            		"description": ""
				},
				{
					"name": "size",
					"type": "integer",
            		"unit": "",
            		"hidden": false,
            		"description": ""
				},
				{
					"name": "status",
					"type": "integer",
					"unit": "",
            		"hidden": false,
            		"description": ""
				}
			],
			"who": "f83e245a-afdc-47ad-a765-4addd1994321",
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

func TestDatasetsService_Trim(t *testing.T) {
	exp := &TrimResult{
		BlocksDeleted: 1,
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		_, err := fmt.Fprint(w, `{
			"numDeleted": 1
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/datasets/test/trim", hf)
	defer teardown()

	res, err := client.Datasets.Trim(context.Background(), "test", time.Hour)
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestDatasetsService_History(t *testing.T) {
	exp := &query.History{
		ID:      "GHP2ufS7OYwMeBhXHj",
		Kind:    query.Analytics,
		Dataset: "test",
		Owner:   "f83e245a-afdc-47ad-a765-4addd1994321",
		Query: query.Query{
			StartTime: mustTimeParse(t, time.RFC3339, "2020-11-18T13:00:00.000Z"),
			EndTime:   mustTimeParse(t, time.RFC3339, "2020-11-25T14:00:00.000Z"),
			Limit:     100,
		},
		CreatedAt: mustTimeParse(t, time.RFC3339, "2020-12-08T13:28:52.78954814Z"),
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
			"who": "f83e245a-afdc-47ad-a765-4addd1994321"
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
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		assert.Equal(t, "time", r.URL.Query().Get("timestamp-field"))
		assert.Equal(t, "2/Jan/2006:15:04:05 +0000", r.URL.Query().Get("timestamp-format"))
		assert.Equal(t, ";", r.URL.Query().Get("csv-delimiter"))

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

	res, err := client.Datasets.Ingest(context.Background(), "test", r, JSON, Identity, IngestOptions{
		TimestampField:  "time",
		TimestampFormat: "2/Jan/2006:15:04:05 +0000",
		CSVDelimiter:    ";", // Obviously not valid for JSON, but perfectly fine to test for its presence in this test.
	})
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
		assert.Equal(t, "application/x-ndjson", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

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
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		assert.Equal(t, "1s", r.URL.Query().Get("streaming-duration"))
		assert.Equal(t, "true", r.URL.Query().Get("nocache"))
		assert.Equal(t, "analytics", r.URL.Query().Get("saveAsKind"))

		w.Header().Set("X-Axiom-History-Query-Id", "fyTFUldK4Z5219rWaz")

		_, err := fmt.Fprint(w, actQueryResp)
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
		SaveKind:          query.Analytics,
	})
	require.NoError(t, err)

	assert.Equal(t, expQueryRes, res)
}

func TestDatasetsService_Query_InvalidSaveKind(t *testing.T) {
	client, teardown := setup(t, "/api/v1/datasets/test/query", nil)
	defer teardown()

	_, err := client.Datasets.Query(context.Background(), "test", query.Query{}, query.Options{
		SaveKind: query.APL,
	})
	require.EqualError(t, err, `invalid query kind "apl": must be "analytics" or "stream"`)
}

func TestDatasetsService_APLQuery(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		assert.Equal(t, "true", r.URL.Query().Get("saveAsKind"))
		assert.Equal(t, "legacy", r.URL.Query().Get("format"))

		var req aplQueryRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if assert.NoError(t, err) {
			assert.Equal(t, "['test'] | where response == 304", req.Raw)
			assert.NotEmpty(t, req.StartTime)
			assert.Empty(t, req.EndTime)
		}

		w.Header().Set("X-Axiom-History-Query-Id", "fyTFUldK4Z5219rWaz")

		_, err = fmt.Fprint(w, actAPLQueryResp)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/datasets/_apl", hf)
	defer teardown()

	res, err := client.Datasets.APLQuery(context.Background(),
		"['test'] | where response == 304", apl.Options{
			StartTime: time.Now().Add(-5 * time.Minute),
			Save:      true,
		})
	require.NoError(t, err)

	assert.Equal(t, expAPLQueryRes, res)
}

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ContentType
		wantErr string
	}{
		{
			name:  JSON.String(),
			input: `[{"a":"b"}, {"c":"d"}]`,
			want:  JSON,
		},
		{
			name:  NDJSON.String(),
			input: `{"a":"b"}`,
			want:  NDJSON,
		},
		{
			name: NDJSON.String(),
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
			_, got, err := DetectContentType(strings.NewReader(tt.input))
			if tt.want > 0 {
				require.NoError(t, err)
				assert.Equal(t, tt.want.String(), got.String())
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func assertValidJSON(t *testing.T, r io.Reader) bool {
	dec := json.NewDecoder(r)
	for dec.More() {
		var v interface{}
		if err := dec.Decode(&v); !assert.NoError(t, err) {
			return false
		} else if !assert.NotEmpty(t, v) {
			return false
		}
	}
	return true
}

func parseTimeOrPanic(layout, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return t
}
