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
	exp := []*VirtualFieldWithID{
		{
			ID: "vfield1",
			VirtualField: VirtualField{
				Dataset:    "dataset1",
				Name:       "field1",
				Expression: "a + b",
				Type:       "number",
			},
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "dataset1", r.URL.Query().Get("dataset"))

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `[{
			"id": "vfield1",
			"dataset": "dataset1",
			"name": "field1",
			"expression": "a + b",
			"type": "number"
		}]`)
		assert.NoError(t, err)
	}
	client := setup(t, "GET /v2/vfields", hf)

	res, err := client.VirtualFields.List(context.Background(), "dataset1")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestVirtualFieldsService_Get(t *testing.T) {
	exp := &VirtualFieldWithID{
		ID: "vfield1",
		VirtualField: VirtualField{
			Dataset:    "dataset1",
			Name:       "field1",
			Expression: "a + b",
			Type:       "number",
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
			"id": "vfield1",
			"dataset": "dataset1",
			"name": "field1",
			"expression": "a + b",
			"type": "number"
		}`)
		assert.NoError(t, err)
	}
	client := setup(t, "GET /v2/vfields/vfield1", hf)

	res, err := client.VirtualFields.Get(context.Background(), "vfield1")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestVirtualFieldsService_Create(t *testing.T) {
	exp := &VirtualFieldWithID{
		ID: "vfield1",
		VirtualField: VirtualField{
			Dataset:    "dataset1",
			Name:       "field1",
			Expression: "a + b",
			Type:       "number",
		},
	}
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, mediaTypeJSON, r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
			"id": "vfield1",
			"dataset": "dataset1",
			"name": "field1",
			"expression": "a + b",
			"type": "number"
		}`)
		assert.NoError(t, err)
	}
	client := setup(t, "POST /v2/vfields", hf)

	res, err := client.VirtualFields.Create(context.Background(), VirtualField{
		Dataset:    "dataset1",
		Name:       "field1",
		Expression: "a + b",
		Type:       "number",
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestVirtualFieldsService_Update(t *testing.T) {
	exp := &VirtualFieldWithID{
		ID: "vfield1",
		VirtualField: VirtualField{
			Dataset:    "dataset1",
			Name:       "field1_updated",
			Expression: "a - b",
			Type:       "number",
		},
	}
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, mediaTypeJSON, r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
			"id": "vfield1",
			"dataset": "dataset1",
			"name": "field1_updated",
			"expression": "a - b",
			"type": "number"
		}`)
		assert.NoError(t, err)
	}
	client := setup(t, "PUT /v2/vfields/vfield1", hf)

	res, err := client.VirtualFields.Update(context.Background(), "vfield1", VirtualField{
		Dataset:    "dataset1",
		Name:       "field1_updated",
		Expression: "a - b",
		Type:       "number",
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestVirtualFieldsService_Delete(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}

	client := setup(t, "DELETE /v2/vfields/vfield1", hf)

	err := client.VirtualFields.Delete(context.Background(), "vfield1")
	require.NoError(t, err)
}
