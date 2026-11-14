package axiom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/axiom/mpl"
)

func TestMetricsService_Query(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.RawQuery)

		body, err := io.ReadAll(r.Body)
		if assert.NoError(t, err) {
			assert.JSONEq(t, `{
				"mpl": "test:metric | align to 5m using avg",
				"startTime": "2026-09-17T08:00:00Z",
				"endTime": "2026-09-17T09:00:00Z",
				"params": {"param__name": "\"frontend\""}
			}`, string(body))
		}

		w.Header().Set("Content-Type", "application/vnd.metrics.v2+json")
		w.Header().Set("X-Axiom-Trace-Id", "abc")
		_, err = fmt.Fprint(w, `{
			"metadata": {"group_keys": ["name"], "unit": "s"},
			"series": [
				{
					"metric": "metric",
					"tags": {"name": "frontend"},
					"start": 1789632900,
					"resolution": 300,
					"data": [0.5, null, 1.5],
					"summary": 1.5
				}
			]
		}`)
		assert.NoError(t, err)
	}

	client := setupAccept(t, "POST /v1/query/_mpl", mediaTypeMetricsV2+", "+mediaTypeJSON, hf)
	require.NoError(t, client.Options(SetEdgeURL(client.config.BaseURL().String())))

	start := time.Date(2026, 9, 17, 8, 0, 0, 0, time.UTC)
	res, err := client.Metrics.Query(t.Context(),
		"test:metric | align to 5m using avg",
		start, start.Add(time.Hour),
		mpl.SetParam("name", `"frontend"`),
	)
	require.NoError(t, err)

	assert.Equal(t, &mpl.Result{
		Metadata: mpl.Metadata{GroupKeys: []string{"name"}, Unit: "s"},
		Series: []mpl.Series{{
			Metric:     "metric",
			Tags:       map[string]any{"name": "frontend"},
			Start:      time.Date(2026, 9, 17, 8, 15, 0, 0, time.UTC),
			Resolution: 5 * time.Minute,
			Data:       []*float64{new(0.5), nil, new(1.5)},
			Summary:    new(1.5),
		}},
		TraceID: "abc",
	}, res)
}

func TestMetricsService_Info(t *testing.T) {
	var (
		start = time.Date(2026, 9, 17, 8, 0, 0, 500_000_000, time.UTC)
		end   = start.Add(time.Hour)
	)

	tests := []struct {
		name    string
		pattern string
		accept  string
		request string
		resp    string
		call    func(context.Context, *MetricsService) (any, error)
		want    any
	}{
		{
			name:    "List",
			pattern: "GET /v1/query/metrics/info/datasets/test/metrics",
			accept:  mediaTypeMetricsInfoV2 + ", " + mediaTypeJSON,
			resp:    `{"metric": {"type": "Gauge", "temporality": null, "unit": ""}}`,
			call: func(ctx context.Context, s *MetricsService) (any, error) {
				return s.List(ctx, "test", start, end)
			},
			want: map[string]MetricInfo{"metric": {Type: "Gauge"}},
		},
		{
			name:    "Tags",
			pattern: "GET /v1/query/metrics/info/datasets/test/metrics/metric/tags",
			accept:  mediaTypeJSON,
			resp:    `["code", "test"]`,
			call: func(ctx context.Context, s *MetricsService) (any, error) {
				return s.Tags(ctx, "test", "metric", start, end)
			},
			want: []string{"code", "test"},
		},
		{
			name:    "TagValues",
			pattern: "GET /v1/query/metrics/info/datasets/test/metrics/metric/tags/le/values",
			accept:  mediaTypeJSON,
			resp:    `[25.0, "+Inf"]`,
			call: func(ctx context.Context, s *MetricsService) (any, error) {
				return s.TagValues(ctx, "test", "metric", "le", start, end)
			},
			want: []any{json.Number("25.0"), "+Inf"},
		},
		{
			name:    "DatasetTags",
			pattern: "GET /v1/query/metrics/info/datasets/test/tags",
			accept:  mediaTypeJSON,
			resp:    `["code"]`,
			call: func(ctx context.Context, s *MetricsService) (any, error) {
				return s.DatasetTags(ctx, "test", start, end)
			},
			want: []string{"code"},
		},
		{
			name:    "DatasetTagValues",
			pattern: "GET /v1/query/metrics/info/datasets/test/tags/code/values",
			accept:  mediaTypeJSON,
			resp:    `[200]`,
			call: func(ctx context.Context, s *MetricsService) (any, error) {
				return s.DatasetTagValues(ctx, "test", "code", start, end)
			},
			want: []any{json.Number("200")},
		},
		{
			name:    "Find",
			pattern: "POST /v1/query/metrics/info/datasets/test/metrics",
			accept:  mediaTypeJSON,
			request: `{"value":25.0}` + "\n",
			resp:    `{"metric": ["le"]}`,
			call: func(ctx context.Context, s *MetricsService) (any, error) {
				return s.Find(ctx, "test", json.Number("25.0"), start, end)
			},
			want: map[string][]string{"metric": {"le"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hf := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "2026-09-17T08:00:00.5Z", r.URL.Query().Get("start"))
				assert.Equal(t, "2026-09-17T09:00:00.5Z", r.URL.Query().Get("end"))

				body, err := io.ReadAll(r.Body)
				if assert.NoError(t, err) {
					assert.Equal(t, tt.request, string(body))
				}

				w.Header().Set("Content-Type", mediaTypeJSON)
				_, err = fmt.Fprint(w, tt.resp)
				assert.NoError(t, err)
			}

			client := setupAccept(t, tt.pattern, tt.accept, hf)
			require.NoError(t, client.Options(SetEdgeURL(client.config.BaseURL().String())))

			res, err := tt.call(t.Context(), client.Metrics)
			require.NoError(t, err)

			assert.Equal(t, tt.want, res)
		})
	}
}

func TestMetricsService_MissingEdge(t *testing.T) {
	hf := func(http.ResponseWriter, *http.Request) {
		t.Error("request sent without an edge endpoint")
	}

	client := setup(t, "/", hf)

	var (
		ctx   = t.Context()
		s     = client.Metrics
		start = time.Now().Add(-time.Hour)
		end   = time.Now()
	)

	calls := map[string]func() error{
		"Query": func() error {
			_, err := s.Query(ctx, "test:metric", start, end)
			return err
		},
		"List": func() error {
			_, err := s.List(ctx, "test", start, end)
			return err
		},
		"Tags": func() error {
			_, err := s.Tags(ctx, "test", "metric", start, end)
			return err
		},
		"TagValues": func() error {
			_, err := s.TagValues(ctx, "test", "metric", "tag", start, end)
			return err
		},
		"DatasetTags": func() error {
			_, err := s.DatasetTags(ctx, "test", start, end)
			return err
		},
		"DatasetTagValues": func() error {
			_, err := s.DatasetTagValues(ctx, "test", "tag", start, end)
			return err
		},
		"Find": func() error {
			_, err := s.Find(ctx, "test", "value", start, end)
			return err
		},
	}
	for name, call := range calls {
		assert.ErrorIs(t, call(), ErrMissingEdge, name)
	}
}
