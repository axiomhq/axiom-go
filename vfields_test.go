package axiom

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVirtualFieldsService_List(t *testing.T) {
	exp := []*VirtualField{
		{
			ID:          "PiGheBIFBc4Khn4dBZ",
			Dataset:     "nginx-logs",
			Name:        "Successful Requests",
			Description: "Statuses <= x < 400",
			Alias:       "status_success",
			Expression:  "response <= 200 && response < 400",
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		assert.Equal(t, "nginx-logs", r.URL.Query().Get("dataset"))
		assert.Equal(t, "1", r.URL.Query().Get("limit"))
		assert.Equal(t, "1", r.URL.Query().Get("offset"))

		_, err := fmt.Fprint(w, `[
			{
				"dataset": "nginx-logs",
				"name": "Successful Requests",
				"description": "Statuses <= x < 400",
				"alias": "status_success",
				"expression": "response <= 200 && response < 400",
				"id": "PiGheBIFBc4Khn4dBZ"
			}
		]`)
		require.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/vfields", hf)
	defer teardown()

	res, err := client.VirtualFields.List(context.Background(), VirtualFieldListOptions{
		Dataset: "nginx-logs",
		ListOptions: ListOptions{
			Limit:  1,
			Offset: 1,
		},
	})
	require.NoError(t, err)

	assert.EqualValues(t, exp, res)
}

func TestVirtualFieldsService_Get(t *testing.T) {
	exp := &VirtualField{
		ID:          "PiGheBIFBc4Khn4dBZ",
		Dataset:     "nginx-logs",
		Name:        "Successful Requests",
		Description: "Statuses <= x < 400",
		Alias:       "status_success",
		Expression:  "response <= 200 && response < 400",
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `{
			"dataset": "nginx-logs",
			"name": "Successful Requests",
			"description": "Statuses <= x < 400",
			"alias": "status_success",
			"expression": "response <= 200 && response < 400",
			"id": "PiGheBIFBc4Khn4dBZ"
		}`)
		require.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/vfields/PiGheBIFBc4Khn4dBZ", hf)
	defer teardown()

	res, err := client.VirtualFields.Get(context.Background(), "PiGheBIFBc4Khn4dBZ")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestVirtualFieldsService_Create(t *testing.T) {
	exp := &VirtualField{
		ID:          "FmgciXxL3njoNgzWVR",
		Dataset:     "nginx-logs",
		Name:        "Failed Requests",
		Description: "Statuses >= 400",
		Alias:       "status_failed",
		Expression:  "response >= 400",
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		_, err := fmt.Fprint(w, `{
			"dataset": "nginx-logs",
			"name": "Failed Requests",
			"description": "Statuses >= 400",
			"alias": "status_failed",
			"expression": "response >= 400",
			"id": "FmgciXxL3njoNgzWVR"
		}`)
		require.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/vfields", hf)
	defer teardown()

	res, err := client.VirtualFields.Create(context.Background(), VirtualField{
		Dataset:     "nginx-logs",
		Name:        "Failed Requests",
		Description: "Statuses >= 400",
		Alias:       "status_failed",
		Expression:  "response >= 400",
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestVirtualFieldsService_Update(t *testing.T) {
	exp := &VirtualField{
		ID:          "FmgciXxL3njoNgzWVR",
		Dataset:     "nginx-logs",
		Name:        "Failed Requests",
		Description: "Statuses > 399",
		Alias:       "status_failed",
		Expression:  "response > 399",
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)

		_, err := fmt.Fprint(w, `{
			"dataset": "nginx-logs",
			"name": "Failed Requests",
			"description": "Statuses > 399",
			"alias": "status_failed",
			"expression": "response > 399",
			"id": "FmgciXxL3njoNgzWVR"
		}`)
		require.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/vfields/FmgciXxL3njoNgzWVR", hf)
	defer teardown()

	res, err := client.VirtualFields.Update(context.Background(), "FmgciXxL3njoNgzWVR", VirtualField{
		Dataset:     "nginx-logs",
		Name:        "Failed Requests",
		Description: "Statuses > 399",
		Alias:       "status_failed",
		Expression:  "response > 399",
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestVirtualFieldsService_Delete(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}

	client, teardown := setup(t, "/api/v1/vfields/FmgciXxL3njoNgzWVR", hf)
	defer teardown()

	err := client.VirtualFields.Delete(context.Background(), "FmgciXxL3njoNgzWVR")
	require.NoError(t, err)
}
