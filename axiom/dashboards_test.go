package axiom

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardsService_GetRaw(t *testing.T) {
	resp := `{
		"uid": "db_test",
                "id": "dash_123",
		"version": 2,
		"dashboard": {"name": "Test Dashboard"},
                "createdAt": "2026-01-01T00:00:00Z",
                "updatedAt": "2026-01-02T00:00:00Z",
                "createdBy": "usr_1",
                "updatedBy": "usr_2"
        }`

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, resp)
		assert.NoError(t, err)
	}

	client := setup(t, "GET /v2/dashboards/uid/db_test", hf)

	res, err := client.Dashboards.GetRaw(t.Context(), "db_test")
	require.NoError(t, err)
	assert.JSONEq(t, resp, string(res))
}

func TestDashboardsService_ListRaw(t *testing.T) {
	resp := `[
                {
                        "uid": "db_test",
                        "id": "dash_123",
                        "version": 1,
                        "dashboard": {"name": "Test Dashboard"},
                        "createdAt": "2026-01-01T00:00:00Z",
                        "updatedAt": "2026-01-01T00:00:00Z",
			"createdBy": "usr_1",
                        "updatedBy": "usr_1"
                }
        ]`

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "10", r.URL.Query().Get("limit"))
		assert.Equal(t, "5", r.URL.Query().Get("offset"))

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, resp)
		assert.NoError(t, err)
	}

	client := setup(t, "GET /v2/dashboards", hf)

	res, err := client.Dashboards.ListRaw(t.Context(), &DashboardsListOptions{Limit: 10, Offset: 5})
	require.NoError(t, err)
	assert.JSONEq(t, resp, string(res))
}

func TestDashboardsService_CreateRaw(t *testing.T) {
	raw := []byte(`{"uid":"db_test","dashboard":{"name":"Raw Dashboard"}}`)

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, mediaTypeJSON, r.Header.Get(headerContentType))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, string(raw), string(body))

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err = fmt.Fprint(w, `{
			"status": "created",
			"dashboard": {
				"uid": "db_test",
				"id": "dash_123",
				"version": 1,
				"dashboard": {"name": "Raw Dashboard"},
				"createdAt": "2026-01-01T00:00:00Z",
				"updatedAt": "2026-01-01T00:00:00Z",
				"createdBy": "usr_1",
				"updatedBy": "usr_1"
			}
		}`)
		assert.NoError(t, err)
	}

	client := setup(t, "POST /v2/dashboards", hf)

	res, err := client.Dashboards.CreateRaw(t.Context(), raw)
	require.NoError(t, err)
	assert.Contains(t, string(res), `"status": "created"`)
	assert.Contains(t, string(res), `"name": "Raw Dashboard"`)
}

func TestDashboardsService_UpdateRaw(t *testing.T) {
	raw := []byte(`{"dashboard":{"name":"Updated Raw Dashboard"},"overwrite":true}`)

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, mediaTypeJSON, r.Header.Get(headerContentType))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, string(raw), string(body))

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err = fmt.Fprint(w, `{
			"status": "updated",
			"dashboard": {
				"uid": "db_test",
				"id": "dash_123",
				"version": 2,
				"dashboard": {"name": "Updated Raw Dashboard"},
				"createdAt": "2026-01-01T00:00:00Z",
				"updatedAt": "2026-01-02T00:00:00Z",
				"createdBy": "usr_1",
				"updatedBy": "usr_2"
			}
		}`)
		assert.NoError(t, err)
	}

	client := setup(t, "PUT /v2/dashboards/uid/db_test", hf)

	res, err := client.Dashboards.UpdateRaw(t.Context(), "db_test", raw)
	require.NoError(t, err)
	assert.Contains(t, string(res), `"status": "updated"`)
	assert.Contains(t, string(res), `"name": "Updated Raw Dashboard"`)
}

func TestDashboardsService_Delete(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	}

	client := setup(t, "DELETE /v2/dashboards/uid/db_test", hf)

	err := client.Dashboards.Delete(t.Context(), "db_test")
	require.NoError(t, err)
}
