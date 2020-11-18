package axiom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardsService_List(t *testing.T) {
	exp := []*Dashboard{
		{
			ID:          "buTFUddK4X5845Qwzv",
			Name:        "Test",
			Description: "A test dashboard",
			Owner:       "e9cffaad-60e7-4b04-8d27-185e1808c38c",
			Charts: []interface{}{
				map[string]interface{}{
					"id":        "5b28c014-8247-4271-a310-7c5953574614",
					"name":      "Total",
					"type":      "TimeSeries",
					"datasetId": "test",
					"query": map[string]interface{}{
						"aggregations": []interface{}{
							map[string]interface{}{
								"op":    "count",
								"field": "",
							},
						},
						"resolution": "15s",
					},
					"modified": float64(1605882074936),
				},
			},
			Layout: []interface{}{
				map[string]interface{}{
					"w":      float64(6),
					"h":      float64(4),
					"x":      float64(0),
					"y":      float64(0),
					"i":      "5b28c014-8247-4271-a310-7c5953574614",
					"minW":   float64(4),
					"minH":   float64(4),
					"moved":  false,
					"static": false,
				},
			},
			RefreshTime:     15 * time.Second,
			SchemaVersion:   2,
			TimeWindowStart: "qr-now-30m",
			TimeWindowEnd:   "qr-now",
			Version:         "1605882077469288241",
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		assert.Equal(t, "1", r.URL.Query().Get("limit"))
		assert.Equal(t, "1", r.URL.Query().Get("offset"))

		_, err := fmt.Fprint(w, `[
			{
				"name": "Test",
				"owner": "e9cffaad-60e7-4b04-8d27-185e1808c38c",
				"description": "A test dashboard",
				"charts": [
					{
						"id": "5b28c014-8247-4271-a310-7c5953574614",
						"name": "Total",
						"type": "TimeSeries",
						"datasetId": "test",
						"query": {
							"aggregations": [
								{
									"op": "count",
									"field": ""
								}
							],
							"resolution": "15s"
						},
						"modified": 1605882074936
					}
				],
				"layout": [
					{
						"w": 6,
						"h": 4,
						"x": 0,
						"y": 0,
						"i": "5b28c014-8247-4271-a310-7c5953574614",
						"minW": 4,
						"minH": 4,
						"moved": false,
						"static": false
					}
				],
				"refreshTime": 15,
				"schemaVersion": 2,
				"timeWindowStart": "qr-now-30m",
				"timeWindowEnd": "qr-now",
				"id": "buTFUddK4X5845Qwzv",
				"version": "1605882077469288241"
			}
		]`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/dashboards", hf)
	defer teardown()

	res, err := client.Dashboards.List(context.Background(), ListOptions{
		Limit:  1,
		Offset: 1,
	})
	require.NoError(t, err)

	assert.EqualValues(t, exp, res)
}

func TestDashboardsService_Get(t *testing.T) {
	exp := &Dashboard{
		ID:          "buTFUddK4X5845Qwzv",
		Name:        "Test",
		Description: "A test dashboard",
		Owner:       "e9cffaad-60e7-4b04-8d27-185e1808c38c",
		Charts: []interface{}{
			map[string]interface{}{
				"id":        "5b28c014-8247-4271-a310-7c5953574614",
				"name":      "Total",
				"type":      "TimeSeries",
				"datasetId": "test",
				"query": map[string]interface{}{
					"aggregations": []interface{}{
						map[string]interface{}{
							"op":    "count",
							"field": "",
						},
					},
					"resolution": "15s",
				},
				"modified": float64(1605882074936),
			},
		},
		Layout: []interface{}{
			map[string]interface{}{
				"w":      float64(6),
				"h":      float64(4),
				"x":      float64(0),
				"y":      float64(0),
				"i":      "5b28c014-8247-4271-a310-7c5953574614",
				"minW":   float64(4),
				"minH":   float64(4),
				"moved":  false,
				"static": false,
			},
		},
		RefreshTime:     15 * time.Second,
		SchemaVersion:   2,
		TimeWindowStart: "qr-now-30m",
		TimeWindowEnd:   "qr-now",
		Version:         "1605882077469288241",
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `{
			"name": "Test",
			"owner": "e9cffaad-60e7-4b04-8d27-185e1808c38c",
			"description": "A test dashboard",
			"charts": [
				{
					"id": "5b28c014-8247-4271-a310-7c5953574614",
					"name": "Total",
					"type": "TimeSeries",
					"datasetId": "test",
					"query": {
						"aggregations": [
							{
								"op": "count",
								"field": ""
							}
						],
						"resolution": "15s"
					},
					"modified": 1605882074936
				}
			],
			"layout": [
				{
					"w": 6,
					"h": 4,
					"x": 0,
					"y": 0,
					"i": "5b28c014-8247-4271-a310-7c5953574614",
					"minW": 4,
					"minH": 4,
					"moved": false,
					"static": false
				}
			],
			"refreshTime": 15,
			"schemaVersion": 2,
			"timeWindowStart": "qr-now-30m",
			"timeWindowEnd": "qr-now",
			"id": "buTFUddK4X5845Qwzv",
			"version": "1605882077469288241"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/dashboards/test", hf)
	defer teardown()

	res, err := client.Dashboards.Get(context.Background(), "test")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestDashboardsService_Create(t *testing.T) {
	exp := &Dashboard{
		ID:          "buTFUddK4X5845Qwzv",
		Name:        "Test",
		Description: "A test dashboard",
		Owner:       "e9cffaad-60e7-4b04-8d27-185e1808c38c",
		Charts: []interface{}{
			map[string]interface{}{
				"id":        "5b28c014-8247-4271-a310-7c5953574614",
				"name":      "Total",
				"type":      "TimeSeries",
				"datasetId": "test",
				"query": map[string]interface{}{
					"aggregations": []interface{}{
						map[string]interface{}{
							"op":    "count",
							"field": "",
						},
					},
					"resolution": "15s",
				},
				"modified": float64(1605882074936),
			},
		},
		Layout: []interface{}{
			map[string]interface{}{
				"w":      float64(6),
				"h":      float64(4),
				"x":      float64(0),
				"y":      float64(0),
				"i":      "5b28c014-8247-4271-a310-7c5953574614",
				"minW":   float64(4),
				"minH":   float64(4),
				"moved":  false,
				"static": false,
			},
		},
		RefreshTime:     15 * time.Second,
		SchemaVersion:   2,
		TimeWindowStart: "qr-now-30m",
		TimeWindowEnd:   "qr-now",
		Version:         "1605882077469288241",
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		_, err := fmt.Fprint(w, `{
			"name": "Test",
			"owner": "e9cffaad-60e7-4b04-8d27-185e1808c38c",
			"description": "A test dashboard",
			"charts": [
				{
					"id": "5b28c014-8247-4271-a310-7c5953574614",
					"name": "Total",
					"type": "TimeSeries",
					"datasetId": "test",
					"query": {
						"aggregations": [
							{
								"op": "count",
								"field": ""
							}
						],
						"resolution": "15s"
					},
					"modified": 1605882074936
				}
			],
			"layout": [
				{
					"w": 6,
					"h": 4,
					"x": 0,
					"y": 0,
					"i": "5b28c014-8247-4271-a310-7c5953574614",
					"minW": 4,
					"minH": 4,
					"moved": false,
					"static": false
				}
			],
			"refreshTime": 15,
			"schemaVersion": 2,
			"timeWindowStart": "qr-now-30m",
			"timeWindowEnd": "qr-now",
			"id": "buTFUddK4X5845Qwzv",
			"version": "1605882077469288241"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/dashboards", hf)
	defer teardown()

	res, err := client.Dashboards.Create(context.Background(), Dashboard{
		Name:        "Test",
		Description: "A test dashboard",
		Owner:       "e9cffaad-60e7-4b04-8d27-185e1808c38c",
		Charts: []interface{}{
			map[string]interface{}{
				"id":        "5b28c014-8247-4271-a310-7c5953574614",
				"name":      "Total",
				"type":      "TimeSeries",
				"datasetId": "test",
				"query": map[string]interface{}{
					"aggregations": []interface{}{
						map[string]interface{}{
							"op":    "count",
							"field": "",
						},
					},
					"resolution": "15s",
				},
				"modified": float64(1605882074936),
			},
		},
		Layout: []interface{}{
			map[string]interface{}{
				"w":      float64(6),
				"h":      float64(4),
				"x":      float64(0),
				"y":      float64(0),
				"i":      "5b28c014-8247-4271-a310-7c5953574614",
				"minW":   float64(4),
				"minH":   float64(4),
				"moved":  false,
				"static": false,
			},
		},
		RefreshTime:     15 * time.Second,
		SchemaVersion:   2,
		TimeWindowStart: "qr-now-30m",
		TimeWindowEnd:   "qr-now",
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestDashboardsService_Update(t *testing.T) {
	exp := &Dashboard{
		ID:          "buTFUddK4X5845Qwzv",
		Name:        "Test",
		Description: "An updated test dashboard",
		Owner:       "e9cffaad-60e7-4b04-8d27-185e1808c38c",
		Charts: []interface{}{
			map[string]interface{}{
				"id":        "5b28c014-8247-4271-a310-7c5953574614",
				"name":      "Total",
				"type":      "TimeSeries",
				"datasetId": "test",
				"query": map[string]interface{}{
					"aggregations": []interface{}{
						map[string]interface{}{
							"op":    "count",
							"field": "",
						},
					},
					"resolution": "15s",
				},
				"modified": float64(1605882074936),
			},
		},
		Layout: []interface{}{
			map[string]interface{}{
				"w":      float64(6),
				"h":      float64(4),
				"x":      float64(0),
				"y":      float64(0),
				"i":      "5b28c014-8247-4271-a310-7c5953574614",
				"minW":   float64(4),
				"minH":   float64(4),
				"moved":  false,
				"static": false,
			},
		},
		RefreshTime:     15 * time.Second,
		SchemaVersion:   2,
		TimeWindowStart: "qr-now-30m",
		TimeWindowEnd:   "qr-now",
		Version:         "1605882077469288241",
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		_, err := fmt.Fprint(w, `{
			"name": "Test",
			"owner": "e9cffaad-60e7-4b04-8d27-185e1808c38c",
			"description": "An updated test dashboard",
			"charts": [
				{
					"id": "5b28c014-8247-4271-a310-7c5953574614",
					"name": "Total",
					"type": "TimeSeries",
					"datasetId": "test",
					"query": {
						"aggregations": [
							{
								"op": "count",
								"field": ""
							}
						],
						"resolution": "15s"
					},
					"modified": 1605882074936
				}
			],
			"layout": [
				{
					"w": 6,
					"h": 4,
					"x": 0,
					"y": 0,
					"i": "5b28c014-8247-4271-a310-7c5953574614",
					"minW": 4,
					"minH": 4,
					"moved": false,
					"static": false
				}
			],
			"refreshTime": 15,
			"schemaVersion": 2,
			"timeWindowStart": "qr-now-30m",
			"timeWindowEnd": "qr-now",
			"id": "buTFUddK4X5845Qwzv",
			"version": "1605882077469288241"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/dashboards/buTFUddK4X5845Qwzv", hf)
	defer teardown()

	res, err := client.Dashboards.Update(context.Background(), "buTFUddK4X5845Qwzv", Dashboard{
		Name:        "Test",
		Description: "An updated test dashboard",
		Owner:       "e9cffaad-60e7-4b04-8d27-185e1808c38c",
		Charts: []interface{}{
			map[string]interface{}{
				"id":        "5b28c014-8247-4271-a310-7c5953574614",
				"name":      "Total",
				"type":      "TimeSeries",
				"datasetId": "test",
				"query": map[string]interface{}{
					"aggregations": []interface{}{
						map[string]interface{}{
							"op":    "count",
							"field": "",
						},
					},
					"resolution": "15s",
				},
				"modified": float64(1605882074936),
			},
		},
		Layout: []interface{}{
			map[string]interface{}{
				"w":      float64(6),
				"h":      float64(4),
				"x":      float64(0),
				"y":      float64(0),
				"i":      "5b28c014-8247-4271-a310-7c5953574614",
				"minW":   float64(4),
				"minH":   float64(4),
				"moved":  false,
				"static": false,
			},
		},
		RefreshTime:     15 * time.Second,
		SchemaVersion:   2,
		TimeWindowStart: "qr-now-30m",
		TimeWindowEnd:   "qr-now",
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestDashboardsService_Delete(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}

	client, teardown := setup(t, "/api/v1/dashboards/buTFUddK4X5845Qwzv", hf)
	defer teardown()

	err := client.Dashboards.Delete(context.Background(), "buTFUddK4X5845Qwzv")
	require.NoError(t, err)
}

func TestDashboard(t *testing.T) {
	exp := Dashboard{
		ID:            "buTFUddK4X5845Qwzv",
		Name:          "Test",
		Description:   "A test dashboard",
		RefreshTime:   time.Second,
		SchemaVersion: 2,
		Version:       "1605882077469288241",
	}

	b, err := json.Marshal(exp)
	require.NoError(t, err)
	require.NotEmpty(t, b)

	var act Dashboard
	err = json.Unmarshal(b, &act)
	require.NoError(t, err)

	assert.Equal(t, exp, act)
}

func TestDashboard_MarshalJSON(t *testing.T) {
	exp := `{
		"name": "",
		"refreshTime": 1,
		"owner": "",
		"description": "",
		"charts": null,
		"layout": null,
		"schemaVersion": 0,
		"timeWindowStart": "",
		"timeWindowEnd": "",
		"id": "",
		"version": ""
	}`

	act, err := Dashboard{
		RefreshTime: time.Second,
	}.MarshalJSON()
	require.NoError(t, err)
	require.NotEmpty(t, act)

	assert.JSONEq(t, exp, string(act))
}

func TestDashboard_UnmarshalJSON(t *testing.T) {
	exp := Dashboard{
		RefreshTime: time.Second,
	}

	var act Dashboard
	err := act.UnmarshalJSON([]byte(`{ "refreshTime": 1 }`))
	require.NoError(t, err)

	assert.Equal(t, exp, act)
}
