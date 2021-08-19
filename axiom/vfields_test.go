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
			ID:          "test.status_success",
			Dataset:     "test",
			Name:        "status_success",
			Description: "Successful Requests",
			Expression:  "response < 400",
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		assert.Equal(t, "test", r.URL.Query().Get("dataset"))
		assert.Equal(t, "1", r.URL.Query().Get("limit"))
		assert.Equal(t, "1", r.URL.Query().Get("offset"))

		_, err := fmt.Fprint(w, `[
			{
				"dataset": "test",
				"description": "Successful Requests",
				"name": "status_success",
				"expression": "response < 400",
				"id": "test.status_success"
			}
		]`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/vfields", hf)
	defer teardown()

	res, err := client.VirtualFields.List(context.Background(), VirtualFieldListOptions{
		Dataset: "test",
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
		ID:          "test.status_success",
		Dataset:     "test",
		Name:        "status_success",
		Description: "Successful Requests",
		Expression:  "response < 400",
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `{
			"dataset": "test",
			"description": "Successful Requests",
			"name": "status_success",
			"expression": "response < 400",
			"id": "test.status_success"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/vfields/test.status_success", hf)
	defer teardown()

	res, err := client.VirtualFields.Get(context.Background(), "test.status_success")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestVirtualFieldsService_Create(t *testing.T) {
	exp := &VirtualField{
		ID:          "status_failed",
		Dataset:     "test",
		Name:        "status_failed",
		Description: "Failed Requests",
		Expression:  "response > 399",
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		_, err := fmt.Fprint(w, `{
			"dataset": "test",
			"description": "Failed Requests",
			"name": "status_failed",
			"expression": "response > 399",
			"id": "status_failed"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/vfields", hf)
	defer teardown()

	res, err := client.VirtualFields.Create(context.Background(), VirtualField{
		Dataset:     "test",
		Name:        "status_failed",
		Description: "Failed Requests",
		Expression:  "response > 399",
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestVirtualFieldsService_Update(t *testing.T) {
	exp := &VirtualField{
		ID:          "status_failed",
		Dataset:     "test",
		Name:        "status_failed",
		Description: "Failed Requests",
		Expression:  "response > 399",
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		_, err := fmt.Fprint(w, `{
			"dataset": "test",
			"description": "Failed Requests",
			"name": "status_failed",
			"expression": "response > 399",
			"id": "status_failed"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/vfields/status_failed", hf)
	defer teardown()

	res, err := client.VirtualFields.Update(context.Background(), "status_failed", VirtualField{
		Dataset:     "test",
		Name:        "status_failed",
		Description: "Failed Requests",
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

	client, teardown := setup(t, "/api/v1/vfields/status_failed", hf)
	defer teardown()

	err := client.VirtualFields.Delete(context.Background(), "status_failed")
	require.NoError(t, err)
}
