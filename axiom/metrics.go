package axiom

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/axiomhq/axiom-go/axiom/mpl"
)

// ErrMissingEdge is raised by [MetricsService] methods when no edge endpoint is
// configured.
var ErrMissingEdge = errors.New("missing edge endpoint, set one with SetEdge, SetEdgeURL, AXIOM_EDGE or AXIOM_EDGE_URL")

type mplQueryRequest struct {
	MPL       string            `json:"mpl"`
	StartTime time.Time         `json:"startTime"`
	EndTime   time.Time         `json:"endTime"`
	Params    map[string]string `json:"params,omitempty"`
}

// mplQueryResponse is the metrics-v2 response. It has no custom unmarshaler, so
// strict decoding still rejects unknown fields.
type mplQueryResponse struct {
	Metadata mpl.Metadata `json:"metadata"`
	Series   []struct {
		Metric     string         `json:"metric"`
		Tags       map[string]any `json:"tags"`
		Start      int64          `json:"start"`      // Unix seconds.
		Resolution int64          `json:"resolution"` // Seconds.
		Data       []*float64     `json:"data"`
		Summary    *float64       `json:"summary"`
	} `json:"series"`
}

// MetricsService handles communication with the metrics related operations of
// the Axiom API.
//
// Its methods only run on an edge endpoint. Configure one with [SetEdge] or
// [SetEdgeURL]. [Dataset.EdgeDeploymentURL] is the edge URL of a dataset.
// Without an edge endpoint, the methods return [ErrMissingEdge].
//
// Axiom API Reference: /v1/query
type MetricsService service

// Query executes the given query specified using the Metrics Processing
// Language (MPL) for the time range from start to end.
//
// To learn more about MPL, please refer to [our documentation].
//
// [our documentation]: https://www.axiom.co/docs/mpl/introduction
func (s *MetricsService) Query(ctx context.Context, q string, start, end time.Time, options ...mpl.Option) (*mpl.Result, error) {
	// Apply supplied options.
	var opts mpl.Options
	for _, option := range options {
		if option != nil {
			option(&opts)
		}
	}

	ctx, span := s.client.trace(ctx, "Metrics.Query", trace.WithAttributes(
		attribute.String("axiom.param.mpl", q),
		attribute.String("axiom.param.start_time", start.Format(time.RFC3339Nano)),
		attribute.String("axiom.param.end_time", end.Format(time.RFC3339Nano)),
	))
	defer span.End()

	edgeURL := s.client.config.EdgeEndpoint(s.basePath + "/_mpl")
	if edgeURL == nil {
		return nil, spanError(span, ErrMissingEdge)
	}

	body := mplQueryRequest{
		MPL:       q,
		StartTime: start,
		EndTime:   end,
	}
	if len(opts.Params) > 0 {
		body.Params = make(map[string]string, len(opts.Params))
		for name, value := range opts.Params {
			body.Params["param__"+name] = value
		}
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, edgeURL.String(), body)
	if err != nil {
		return nil, spanError(span, err)
	}
	// Without the JSON fallback, the edge sends errors without a content type.
	req.Header.Set(headerAccept, mediaTypeMetricsV2+", "+mediaTypeJSON)

	var (
		res  mplQueryResponse
		resp *Response
	)
	if resp, err = s.client.Do(req, &res); err != nil {
		return nil, spanError(span, err)
	}

	result := mpl.Result{
		Metadata: res.Metadata,
		Series:   make([]mpl.Series, len(res.Series)),
		TraceID:  resp.TraceID(),
	}
	for i, series := range res.Series {
		result.Series[i] = mpl.Series{
			Metric:     series.Metric,
			Tags:       series.Tags,
			Start:      time.Unix(series.Start, 0).UTC(),
			Resolution: time.Duration(series.Resolution) * time.Second,
			Data:       series.Data,
			Summary:    series.Summary,
		}
	}

	span.SetAttributes(attribute.String("axiom.trace_id", result.TraceID))

	return &result, nil
}
