package axiom

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnnotationService_Create(t *testing.T) {
	exp := &AnnotationCreateRequest{
		Type:     "test",
		Datasets: []string{"test"},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
				"id": "ann_test",
				"type": "test",
				"datasets": ["test"]
			}`)
		assert.NoError(t, err)
	}

	client := setup(t, "POST /v2/annotations", hf)

	res, err := client.Annotations.Create(context.Background(), exp)
	require.NoError(t, err)

	assert.Equal(t, exp.Type, res.Type)
	assert.Equal(t, exp.Datasets, res.Datasets)
}

func TestAnnotationService_Get(t *testing.T) {
	exp := &AnnotationCreateRequest{
		Type:     "test",
		Datasets: []string{"test"},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
				"id": "ann_test",
				"type": "test",
				"datasets": ["test"]
			}`)
		assert.NoError(t, err)
	}

	client := setup(t, "GET /v2/annotations/ann_test", hf)

	res, err := client.Annotations.Get(context.Background(), "ann_test")
	require.NoError(t, err)

	assert.Equal(t, exp.Type, res.Type)
	assert.Equal(t, exp.Datasets, res.Datasets)
}

func TestAnnotationService_List(t *testing.T) {
	exp := []*Annotation{
		{
			ID:       "ann_test",
			Type:     "test",
			Datasets: []string{"test"},
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `[
			{
				"id": "ann_test",
				"type": "test",
				"datasets": ["test"]
			}
		]`)
		assert.NoError(t, err)
	}

	client := setup(t, "GET /v2/annotations", hf)

	res, err := client.Annotations.List(context.Background(), nil)
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestAnnotationService_Update(t *testing.T) {
	exp := &AnnotationUpdateRequest{
		Type:     "test-2",
		Datasets: []string{"test"},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
				"id": "ann_test",
				"type": "test-2",
				"datasets": ["test"]
			}`)
		assert.NoError(t, err)
	}

	client := setup(t, "PUT /v2/annotations/ann_test", hf)

	res, err := client.Annotations.Update(context.Background(), "ann_test", exp)
	require.NoError(t, err)

	assert.Equal(t, exp.Type, res.Type)
	assert.Equal(t, exp.Datasets, res.Datasets)
}

func TestAnnotationService_Delete(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
	}

	client := setup(t, "DELETE /v2/annotations/ann_test", hf)

	err := client.Annotations.Delete(context.Background(), "ann_test")
	require.NoError(t, err)
}
