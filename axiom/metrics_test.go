package axiom

import (
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

func TestMetricsService_Query_MissingEdge(t *testing.T) {
	hf := func(http.ResponseWriter, *http.Request) {
		t.Error("request sent without an edge endpoint")
	}

	client := setup(t, "/", hf)

	_, err := client.Metrics.Query(t.Context(), "test:metric", time.Now().Add(-time.Hour), time.Now())
	require.ErrorIs(t, err, ErrMissingEdge)
}
