package axiom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/axiomhq/axiom-go/axiom/query"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonitorsService_List(t *testing.T) {
	exp := []*Monitor{
		{
			ID:          "nGxDh3TGuidQJgJW3s",
			Dataset:     "test",
			Name:        "Test",
			Description: "A test monitor",
			Query: query.Query{
				StartTime:  mustTimeParse(t, time.RFC3339, "2020-11-30T14:28:29Z"),
				EndTime:    mustTimeParse(t, time.RFC3339, "2020-11-30T14:33:29Z"),
				Resolution: time.Second,
			},
			Threshold:     1000,
			Comparison:    AboveOrEqual,
			Frequency:     time.Minute,
			Duration:      5 * time.Minute,
			LastCheckTime: mustTimeParse(t, time.RFC3339, "2020-11-30T14:37:13Z"),
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `[
			{
				"id": "nGxDh3TGuidQJgJW3s",
				"name": "Test",
				"description": "A test monitor",
				"disabledUntil": "0001-01-01T00:00:00Z",
				"query": {
					"startTime": "2020-11-30T14:28:29Z",
					"endTime": "2020-11-30T14:33:29Z",
					"resolution": "1s"
				},
				"dataset": "test",
				"threshold": 1000,
				"comparison": "AboveOrEqual",
				"frequencyMinutes": 1,
				"durationMinutes": 5,
				"lastCheckTime": "2020-11-30T14:37:13Z"
			}
		]`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/monitors", hf)
	defer teardown()

	res, err := client.Monitors.List(context.Background())
	require.NoError(t, err)

	assert.EqualValues(t, exp, res)
}

func TestMonitorsService_Get(t *testing.T) {
	exp := &Monitor{
		ID:          "nGxDh3TGuidQJgJW3s",
		Dataset:     "test",
		Name:        "Test",
		Description: "A test monitor",
		Query: query.Query{
			StartTime:  mustTimeParse(t, time.RFC3339, "2020-11-30T14:28:29Z"),
			EndTime:    mustTimeParse(t, time.RFC3339, "2020-11-30T14:33:29Z"),
			Resolution: time.Second,
		},
		Threshold:     1000,
		Comparison:    AboveOrEqual,
		Frequency:     time.Minute,
		Duration:      5 * time.Minute,
		LastCheckTime: mustTimeParse(t, time.RFC3339, "2020-11-30T14:37:13Z"),
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `{
			"id": "nGxDh3TGuidQJgJW3s",
			"name": "Test",
			"description": "A test monitor",
			"disabledUntil": "0001-01-01T00:00:00Z",
			"query": {
				"startTime": "2020-11-30T14:28:29Z",
				"endTime": "2020-11-30T14:33:29Z",
				"resolution": "1s"
			},
			"dataset": "test",
			"threshold": 1000,
			"comparison": "AboveOrEqual",
			"frequencyMinutes": 1,
			"durationMinutes": 5,
			"lastCheckTime": "2020-11-30T14:37:13Z"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/monitors/nGxDh3TGuidQJgJW3s", hf)
	defer teardown()

	res, err := client.Monitors.Get(context.Background(), "nGxDh3TGuidQJgJW3s")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestMonitorsService_Create(t *testing.T) {
	exp := &Monitor{
		ID:          "lrR66wmzYm9NKtq0rz",
		Name:        "Test",
		Description: "A test monitor",
		Dataset:     "test",
		Comparison:  Below,
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		_, err := fmt.Fprint(w, `{
			"id": "lrR66wmzYm9NKtq0rz",
			"name": "Test",
			"description": "A test monitor",
			"disabledUntil": "0001-01-01T00:00:00Z",
			"query": {
				"startTime": "0001-01-01T00:00:00Z",
				"endTime": "0001-01-01T00:00:00Z",
				"resolution": ""
			},
			"dataset": "test",
			"threshold": 0,
			"comparison": "Below",
			"frequencyMinutes": 0,
			"durationMinutes": 0,
			"lastCheckTime": "0001-01-01T00:00:00Z"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/monitors", hf)
	defer teardown()

	res, err := client.Monitors.Create(context.Background(), Monitor{
		Name:        "Test",
		Description: "A test monitor",
		Dataset:     "test",
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestMonitorsService_Update(t *testing.T) {
	exp := &Monitor{
		ID:          "lrR66wmzYm9NKtq0rz",
		Name:        "Test",
		Description: "A very good test monitor",
		Dataset:     "test",
		Comparison:  Below,
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		_, err := fmt.Fprint(w, `{
			"id": "lrR66wmzYm9NKtq0rz",
			"name": "Test",
			"description": "A very good test monitor",
			"disabledUntil": "0001-01-01T00:00:00Z",
			"query": {
				"startTime": "0001-01-01T00:00:00Z",
				"endTime": "0001-01-01T00:00:00Z",
				"resolution": ""
			},
			"dataset": "test",
			"threshold": 0,
			"comparison": "Below",
			"frequencyMinutes": 0,
			"durationMinutes": 0,
			"lastCheckTime": "0001-01-01T00:00:00Z"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/monitors/lrR66wmzYm9NKtq0rz", hf)
	defer teardown()

	res, err := client.Monitors.Update(context.Background(), "lrR66wmzYm9NKtq0rz", Monitor{
		Name:        "Test",
		Description: "A very good test monitor",
		Dataset:     "test",
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestMonitorsService_Delete(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}

	client, teardown := setup(t, "/api/v1/monitors/lrR66wmzYm9NKtq0rz", hf)
	defer teardown()

	err := client.Monitors.Delete(context.Background(), "lrR66wmzYm9NKtq0rz")
	require.NoError(t, err)
}

func TestComparison_Marshal(t *testing.T) {
	exp := `{
		"comparison": "AboveOrEqual"
	}`

	b, err := json.Marshal(struct {
		Comparison Comparison `json:"comparison"`
	}{
		Comparison: AboveOrEqual,
	})
	require.NoError(t, err)
	require.NotEmpty(t, b)

	assert.JSONEq(t, exp, string(b))
}

func TestComparison_Unmarshal(t *testing.T) {
	var act struct {
		Comparison Comparison `json:"comparison"`
	}
	err := json.Unmarshal([]byte(`{ "comparison": "AboveOrEqual" }`), &act)
	require.NoError(t, err)

	assert.Equal(t, AboveOrEqual, act.Comparison)
}

func TestComparison_String(t *testing.T) {
	// Check outer bounds.
	assert.Contains(t, (Below - 1).String(), "Comparison(")
	assert.Contains(t, (AboveOrEqual + 1).String(), "Comparison(")

	for c := Below; c <= AboveOrEqual; c++ {
		s := c.String()
		assert.NotEmpty(t, s)
		assert.NotContains(t, s, "Comparison(")
	}
}

func TestMonitor(t *testing.T) {
	exp := Monitor{
		ID:          "lrR66wmzYm9NKtq0rz",
		Dataset:     "test",
		Name:        "Test",
		Description: "A test monitor",
		Threshold:   21.25,
		Comparison:  AboveOrEqual,
	}

	b, err := json.Marshal(exp)
	require.NoError(t, err)
	require.NotEmpty(t, b)

	var act Monitor
	err = json.Unmarshal(b, &act)
	require.NoError(t, err)

	assert.Equal(t, exp, act)
}

func TestMonitor_MarshalJSON(t *testing.T) {
	exp := `{
		"id": "",
		"dataset": "",
		"name": "",
		"description": "",
		"disabledUntil": "0001-01-01T00:00:00Z",
		"query": {
			"startTime": "0001-01-01T00:00:00Z",
			"endTime": "0001-01-01T00:00:00Z",
			"aggregations": null,
			"filter": {
				"op": "",
				"field": "",
				"value": null,
				"caseInsensitive": false,
				"children": null
			},
			"groupBy": null,
			"order": null,
			"limit": 0,
			"virtualFields": null,
			"cursor": "",
			"resolution": "auto"
		},
		"threshold": 0,
		"comparison": "Comparison(0)",
		"noDataCloseWaitMinutes": 1,
		"frequencyMinutes": 2,
		"durationMinutes": 3,
		"notifiers": null,
		"lastCheckTime": "0001-01-01T00:00:00Z",
		"lastCheckState": null
	}`

	act, err := Monitor{
		NoDataCloseWait: time.Minute,
		Frequency:       2 * time.Minute,
		Duration:        3 * time.Minute,
	}.MarshalJSON()
	require.NoError(t, err)
	require.NotEmpty(t, act)

	fmt.Println(string(act))

	assert.JSONEq(t, exp, string(act))
}

func TestMonitor_UnmarshalJSON(t *testing.T) {
	exp := Monitor{
		NoDataCloseWait: time.Minute,
		Frequency:       2 * time.Minute,
		Duration:        3 * time.Minute,
	}

	var act Monitor
	err := act.UnmarshalJSON([]byte(`{ "noDataCloseWaitMinutes": 1, "frequencyMinutes": 2, "durationMinutes": 3 }`))
	require.NoError(t, err)

	assert.Equal(t, exp, act)
}
