package axiom

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		require.NoError(t, err)
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
		Description: "",
		Created:     mustTimeParse(t, time.RFC3339Nano, "2020-11-17T22:29:00.521238198Z"),
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `{
			"id": "test",
			"name": "test",
			"created": "2020-11-17T22:29:00.521238198Z"
		}`)
		require.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/datasets/test", hf)
	defer teardown()

	res, err := client.Datasets.Get(context.Background(), "test")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestDatasetsService_Info(t *testing.T) {
	exp := &DatasetInfo{
		DisplayName:          "test",
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
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `{
			"displayName": "test",
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
			]
		}`)
		require.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/datasets/test/info", hf)
	defer teardown()

	res, err := client.Datasets.Info(context.Background(), "test")
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

		_, err := fmt.Fprint(w, `{
			"id": "test",
			"name": "test",
			"description": "This is a test description",
			"created": "2020-11-18T21:30:20.623322799Z"
		}`)
		require.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/datasets", hf)
	defer teardown()

	res, err := client.Datasets.Create(context.Background(), CreateDatasetRequest{
		Name:        "test",
		Description: "This is a test description",
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

		_, err := fmt.Fprint(w, `{
			"ingested": 2,
			"failed": 0,
			"failures": [],
			"processedBytes": 630,
			"blocksCreated": 0,
			"walLength": 2
		}`)
		require.NoError(t, err)
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

// TODO(lukasmalkmus): Write a test that contains some failures in the server
// response.
