package axiom

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

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

	res, err := client.Annotations.Create(t.Context(), exp)
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

	res, err := client.Annotations.Get(t.Context(), "ann_test")
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

	start := time.Date(2024, 3, 18, 8, 39, 28, 382_000_001, time.FixedZone("IST", 5*60*60+30*60))
	end := start.Add(time.Hour)

	tests := []struct {
		name   string
		filter *AnnotationsFilter
		want   url.Values
	}{
		{
			name: "nil filter",
			want: url.Values{},
		},
		{
			name:   "empty filter",
			filter: &AnnotationsFilter{},
			want:   url.Values{},
		},
		{
			name:   "datasets",
			filter: &AnnotationsFilter{Datasets: []string{"a", "b"}},
			want:   url.Values{"datasets": {"a,b"}},
		},
		{
			name:   "time range",
			filter: &AnnotationsFilter{Start: start, End: end},
			want: url.Values{
				"start": {"2024-03-18T08:39:28.382000001+05:30"},
				"end":   {"2024-03-18T09:39:28.382000001+05:30"},
			},
		},
		{
			name: "utc time range",
			filter: &AnnotationsFilter{
				Start: time.Date(2024, 3, 18, 8, 39, 28, 0, time.UTC),
				End:   time.Date(2024, 3, 18, 9, 39, 28, 0, time.UTC),
			},
			want: url.Values{
				"start": {"2024-03-18T08:39:28Z"},
				"end":   {"2024-03-18T09:39:28Z"},
			},
		},
		{
			name:   "start only",
			filter: &AnnotationsFilter{Start: start},
			want:   url.Values{"start": {"2024-03-18T08:39:28.382000001+05:30"}},
		},
		{
			name:   "end only",
			filter: &AnnotationsFilter{End: end},
			want:   url.Values{"end": {"2024-03-18T09:39:28.382000001+05:30"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hf := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, tt.want, r.URL.Query())

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

			res, err := client.Annotations.List(t.Context(), tt.filter)
			require.NoError(t, err)

			assert.Equal(t, exp, res)
		})
	}
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

	res, err := client.Annotations.Update(t.Context(), "ann_test", exp)
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

	err := client.Annotations.Delete(t.Context(), "ann_test")
	require.NoError(t, err)
}
