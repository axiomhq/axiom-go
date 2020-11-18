package axiom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/axiom/query"
)

func TestStarredQueriesService_List(t *testing.T) {
	exp := []*StarredQuery{
		{
			ID:      "NBYj9rO5p4F5CtYEy6",
			Kind:    Analytics,
			Dataset: "test",
			Owner:   "610455ff-2b16-4e8a-a3c5-70adde1538ff",
			Name:    "avg(size) shown",
			Query: query.Query{
				StartTime:  mustTimeParse(t, time.RFC3339, "2020-11-24T16:23:15.000Z"),
				EndTime:    mustTimeParse(t, time.RFC3339, "2020-11-24T16:53:30.000Z"),
				Resolution: 15 * time.Second,
				Aggregations: []query.Aggregation{
					{
						Op:    "avg",
						Field: "size",
					},
				},
			},
			Metadata: map[string]string{
				"quickRange": "30m",
			},
			Created: mustTimeParse(t, time.RFC3339Nano, "2020-11-24T16:53:38.267775284Z"),
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		assert.Equal(t, "analytics", r.URL.Query().Get("kind"))
		assert.Equal(t, "team", r.URL.Query().Get("who"))
		assert.Equal(t, "test", r.URL.Query().Get("dataset"))
		assert.Equal(t, "1", r.URL.Query().Get("limit"))
		assert.Equal(t, "1", r.URL.Query().Get("offset"))

		_, err := fmt.Fprint(w, `[
			{
				"kind": "analytics",
				"dataset": "test",
				"name": "avg(size) shown",
				"who": "610455ff-2b16-4e8a-a3c5-70adde1538ff",
				"query": {
					"aggregations": [
						{
							"op": "avg",
							"field": "size"
						}
					],
					"startTime": "2020-11-24T16:23:15.000Z",
					"endTime": "2020-11-24T16:53:30.000Z",
					"resolution": "15s"
				},
				"metadata": {
					"quickRange": "30m"
				},
				"id": "NBYj9rO5p4F5CtYEy6",
				"created": "2020-11-24T16:53:38.267775284Z"
			}
		]`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/starred", hf)
	defer teardown()

	res, err := client.StarredQueries.List(context.Background(), StarredQueriesListOptions{
		Kind:    Analytics,
		Dataset: "test",
		Owner:   OwnedByTeam,
		ListOptions: ListOptions{
			Limit:  1,
			Offset: 1,
		},
	})
	require.NoError(t, err)

	assert.EqualValues(t, exp, res)
}

func TestStarredQueriesService_Get(t *testing.T) {
	exp := &StarredQuery{
		ID:      "NBYj9rO5p4F5CtYEy6",
		Kind:    Analytics,
		Dataset: "test",
		Owner:   "610455ff-2b16-4e8a-a3c5-70adde1538ff",
		Name:    "avg(size) shown",
		Query: query.Query{
			StartTime:  mustTimeParse(t, time.RFC3339, "2020-11-24T16:23:15.000Z"),
			EndTime:    mustTimeParse(t, time.RFC3339, "2020-11-24T16:53:30.000Z"),
			Resolution: 15 * time.Second,
			Aggregations: []query.Aggregation{
				{
					Op:    "avg",
					Field: "size",
				},
			},
		},
		Metadata: map[string]string{
			"quickRange": "30m",
		},
		Created: mustTimeParse(t, time.RFC3339Nano, "2020-11-24T16:53:38.267775284Z"),
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `{
			"kind": "analytics",
			"dataset": "test",
			"name": "avg(size) shown",
			"who": "610455ff-2b16-4e8a-a3c5-70adde1538ff",
			"query": {
				"aggregations": [
					{
						"op": "avg",
						"field": "size"
					}
				],
				"startTime": "2020-11-24T16:23:15.000Z",
				"endTime": "2020-11-24T16:53:30.000Z",
				"resolution": "15s"
			},
			"metadata": {
				"quickRange": "30m"
			},
			"id": "NBYj9rO5p4F5CtYEy6",
			"created": "2020-11-24T16:53:38.267775284Z"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/starred/NBYj9rO5p4F5CtYEy6", hf)
	defer teardown()

	res, err := client.StarredQueries.Get(context.Background(), "NBYj9rO5p4F5CtYEy6")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestStarredQueriesService_Create(t *testing.T) {
	exp := &StarredQuery{
		ID:      "NBYj9rO5p4F5CtYEy6",
		Kind:    Analytics,
		Dataset: "test",
		Owner:   "e9cffaad-60e7-4b04-8d27-185e1808c38c",
		Name:    "Everything",
		Query: query.Query{
			StartTime: mustTimeParse(t, time.RFC3339, "2020-11-24T16:23:15.000Z"),
			EndTime:   mustTimeParse(t, time.RFC3339, "2020-11-24T16:53:30.000Z"),
			Limit:     1000,
		},
		Metadata: map[string]string{
			"quickRange": "7d",
		},
		Created: mustTimeParse(t, time.RFC3339Nano, "2020-11-25T17:34:07.659355723Z"),
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		_, err := fmt.Fprint(w, `{
			"kind": "analytics",
			"dataset": "test",
			"name": "Everything",
			"who": "e9cffaad-60e7-4b04-8d27-185e1808c38c",
			"query": {
				"startTime": "2020-11-24T16:23:15.000Z",
				"endTime": "2020-11-24T16:53:30.000Z",
				"limit": 1000
			},
			"metadata": {
				"quickRange": "7d"
			},
			"id": "NBYj9rO5p4F5CtYEy6",
			"created": "2020-11-25T17:34:07.659355723Z"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/starred", hf)
	defer teardown()

	res, err := client.StarredQueries.Create(context.Background(), StarredQuery{
		Kind:    Analytics,
		Dataset: "test",
		Name:    "Everything",
		Query: query.Query{
			StartTime: mustTimeParse(t, time.RFC3339, "2020-11-24T16:23:15.000Z"),
			EndTime:   mustTimeParse(t, time.RFC3339, "2020-11-24T16:53:30.000Z"),
			Limit:     1000,
		},
		Metadata: map[string]string{
			"quickRange": "7d",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestStarredQueriesService_Update(t *testing.T) {
	exp := &StarredQuery{
		ID:      "NBYj9rO5p4F5CtYEy6",
		Kind:    Analytics,
		Dataset: "test",
		Owner:   "e9cffaad-60e7-4b04-8d27-185e1808c38c",
		Name:    "A fancy query name",
		Query: query.Query{
			StartTime: mustTimeParse(t, time.RFC3339, "2020-11-24T16:23:15.000Z"),
			EndTime:   mustTimeParse(t, time.RFC3339, "2020-11-24T16:53:30.000Z"),
			Limit:     100,
		},
		Metadata: map[string]string{
			"quickRange": "7d",
		},
		Created: mustTimeParse(t, time.RFC3339Nano, "2020-11-25T17:34:07.659355723Z"),
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		_, err := fmt.Fprint(w, `{
			"kind": "analytics",
			"dataset": "test",
			"name": "A fancy query name",
			"who": "e9cffaad-60e7-4b04-8d27-185e1808c38c",
			"query": {
				"startTime": "2020-11-24T16:23:15.000Z",
				"endTime": "2020-11-24T16:53:30.000Z",
				"limit": 100
			},
			"metadata": {
				"quickRange": "7d"
			},
			"id": "NBYj9rO5p4F5CtYEy6",
			"created": "2020-11-25T17:34:07.659355723Z"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/starred/NBYj9rO5p4F5CtYEy6", hf)
	defer teardown()

	res, err := client.StarredQueries.Update(context.Background(), "NBYj9rO5p4F5CtYEy6", StarredQuery{
		Kind:    Analytics,
		Dataset: "test",
		Name:    "A fancy query name",
		Query: query.Query{
			StartTime: mustTimeParse(t, time.RFC3339, "2020-11-24T16:23:15.000Z"),
			EndTime:   mustTimeParse(t, time.RFC3339, "2020-11-24T16:53:30.000Z"),
			Limit:     100,
		},
		Metadata: map[string]string{
			"quickRange": "7d",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestStarredQueriesService_Delete(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}

	client, teardown := setup(t, "/api/v1/starred/NBYj9rO5p4F5CtYEy6", hf)
	defer teardown()

	err := client.StarredQueries.Delete(context.Background(), "NBYj9rO5p4F5CtYEy6")
	require.NoError(t, err)
}

func TestQueryKind_EncodeValues(t *testing.T) {
	tests := []struct {
		input QueryKind
		exp   string
	}{
		{Analytics, "analytics"},
		{Stream, "stream"},
		{0, "QueryKind(0)"}, // HINT(lukasmalkmus): Maybe we want to sort this out by raising an error?
	}
	for _, tt := range tests {
		t.Run(tt.input.String(), func(t *testing.T) {
			v := &url.Values{}
			err := tt.input.EncodeValues("test", v)
			require.NoError(t, err)

			assert.Equal(t, tt.exp, v.Get("test"))
		})
	}
}

func TestQueryKind_Marshal(t *testing.T) {
	exp := `{
		"kind": "analytics"
	}`

	b, err := json.Marshal(struct {
		Kind QueryKind `json:"kind"`
	}{
		Kind: Analytics,
	})
	require.NoError(t, err)
	require.NotEmpty(t, b)

	assert.JSONEq(t, exp, string(b))
}

func TestQueryKind_Unmarshal(t *testing.T) {
	var act struct {
		Kind QueryKind `json:"kind"`
	}
	err := json.Unmarshal([]byte(`{ "kind": "analytics" }`), &act)
	require.NoError(t, err)

	assert.Equal(t, Analytics, act.Kind)
}

func TestQueryKind_String(t *testing.T) {
	// Check outer bounds.
	assert.Contains(t, (Analytics - 1).String(), "QueryKind(")
	assert.Contains(t, (Stream + 1).String(), "QueryKind(")

	for c := Analytics; c <= Stream; c++ {
		s := c.String()
		assert.NotEmpty(t, s)
		assert.NotContains(t, s, "QueryKind(")
	}
}
