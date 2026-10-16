package axiom

import (
	"bytes"
	"context"
	"encoding/json"
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

// MetricInfo describes a metric of a metrics dataset.
type MetricInfo struct {
	// Type of the metric, for example "Gauge" or "Histogram".
	Type string `json:"type"`
	// Temporality of the metric, for example "Cumulative". Empty for a gauge or
	// if unknown.
	Temporality string `json:"temporality"`
	// Unit of the metric. Empty if not defined.
	Unit string `json:"unit"`
}

type findMetricsRequest struct {
	Value any `json:"value"`
}

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

// List returns the metrics of the dataset with data between start and end,
// mapped to their type, temporality and unit.
func (s *MetricsService) List(ctx context.Context, dataset string, start, end time.Time) (map[string]MetricInfo, error) {
	ctx, span := s.client.trace(ctx, "Metrics.List", trace.WithAttributes(
		attribute.String("axiom.dataset_id", dataset),
	))
	defer span.End()

	// The edge sends the metadata only for its vendor media type, but labels a
	// successful response as JSON.
	var res map[string]MetricInfo
	if err := s.info(ctx, http.MethodGet, dataset+"/metrics", mediaTypeMetricsInfoV2+", "+mediaTypeJSON, start, end, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return res, nil
}

// Tags returns the tags of the metric with data between start and end.
func (s *MetricsService) Tags(ctx context.Context, dataset, metric string, start, end time.Time) ([]string, error) {
	ctx, span := s.client.trace(ctx, "Metrics.Tags", trace.WithAttributes(
		attribute.String("axiom.dataset_id", dataset),
	))
	defer span.End()

	var res []string
	if err := s.info(ctx, http.MethodGet, dataset+"/metrics/"+metric+"/tags", "", start, end, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return res, nil
}

// TagValues returns the values of the tag of the metric with data between start
// and end. Numeric values are [json.Number], which keeps the form that
// [MetricsService.Find] matches. Recently ingested data can be missing.
func (s *MetricsService) TagValues(ctx context.Context, dataset, metric, tag string, start, end time.Time) ([]any, error) {
	ctx, span := s.client.trace(ctx, "Metrics.TagValues", trace.WithAttributes(
		attribute.String("axiom.dataset_id", dataset),
	))
	defer span.End()

	var raw json.RawMessage
	if err := s.info(ctx, http.MethodGet, dataset+"/metrics/"+metric+"/tags/"+tag+"/values", "", start, end, nil, &raw); err != nil {
		return nil, spanError(span, err)
	}

	res, err := decodeTagValues(raw)
	if err != nil {
		return nil, spanError(span, err)
	}

	return res, nil
}

// DatasetTags returns the tags of all metrics of the dataset with data between
// start and end.
func (s *MetricsService) DatasetTags(ctx context.Context, dataset string, start, end time.Time) ([]string, error) {
	ctx, span := s.client.trace(ctx, "Metrics.DatasetTags", trace.WithAttributes(
		attribute.String("axiom.dataset_id", dataset),
	))
	defer span.End()

	var res []string
	if err := s.info(ctx, http.MethodGet, dataset+"/tags", "", start, end, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return res, nil
}

// DatasetTagValues returns the values of the tag across all metrics of the
// dataset with data between start and end. Numeric values are [json.Number],
// which keeps the form that [MetricsService.Find] matches. Recently ingested
// data can be missing.
func (s *MetricsService) DatasetTagValues(ctx context.Context, dataset, tag string, start, end time.Time) ([]any, error) {
	ctx, span := s.client.trace(ctx, "Metrics.DatasetTagValues", trace.WithAttributes(
		attribute.String("axiom.dataset_id", dataset),
	))
	defer span.End()

	var raw json.RawMessage
	if err := s.info(ctx, http.MethodGet, dataset+"/tags/"+tag+"/values", "", start, end, nil, &raw); err != nil {
		return nil, spanError(span, err)
	}

	res, err := decodeTagValues(raw)
	if err != nil {
		return nil, spanError(span, err)
	}

	return res, nil
}

// Find returns the metrics of the dataset with data between start and end that
// have a tag with the value, mapped to the names of these tags. Recently
// ingested data can be missing.
//
// The value matches only a tag value with the same JSON form. A float64 of 25.0
// encodes as 25 and does not match the tag value 25.0. Pass a value from
// [MetricsService.TagValues] as is, or a [json.Number] such as "25.0".
func (s *MetricsService) Find(ctx context.Context, dataset string, value any, start, end time.Time) (map[string][]string, error) {
	ctx, span := s.client.trace(ctx, "Metrics.Find", trace.WithAttributes(
		attribute.String("axiom.dataset_id", dataset),
	))
	defer span.End()

	var res map[string][]string
	if err := s.info(ctx, http.MethodPost, dataset+"/metrics", "", start, end, findMetricsRequest{Value: value}, &res); err != nil {
		return nil, spanError(span, err)
	}

	return res, nil
}

// decodeTagValues decodes numeric tag values as [json.Number].
func decodeTagValues(raw json.RawMessage) ([]any, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()

	var values []any
	if err := dec.Decode(&values); err != nil {
		return nil, err
	}
	return values, nil
}

// info calls the metrics info endpoint at path, which starts with the dataset.
// A non-empty accept replaces the default Accept header.
func (s *MetricsService) info(ctx context.Context, method, path, accept string, start, end time.Time, body, v any) error {
	edgeURL := s.client.config.EdgeEndpoint(s.basePath + "/metrics/info/datasets/" + path)
	if edgeURL == nil {
		return ErrMissingEdge
	}

	u, err := AddURLOptions(edgeURL.String(), struct {
		Start time.Time `url:"start" layout:"2006-01-02T15:04:05.999999999Z07:00"`
		End   time.Time `url:"end" layout:"2006-01-02T15:04:05.999999999Z07:00"`
	}{
		Start: start,
		End:   end,
	})
	if err != nil {
		return err
	}

	req, err := s.client.NewRequest(ctx, method, u, body)
	if err != nil {
		return err
	}
	if accept != "" {
		req.Header.Set(headerAccept, accept)
	}

	_, err = s.client.Do(req, v)
	return err
}
