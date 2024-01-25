package axiom

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"net/url"
)

type Monitor struct {
	AlertOnNoData   bool     `json:"alertOnNoData"`
	AplQuery        string   `json:"aplQuery"`
	Description     string   `json:"description,omitempty"`
	Disabled        bool     `json:"disabled"`
	ID              string   `json:"id,omitempty"`
	IntervalMinutes int64    `json:"intervalMinutes"`
	MatchEveryN     int64    `json:"matchEveryN,omitempty"`
	MatchValue      string   `json:"matchValue,omitempty"`
	Name            string   `json:"name"`
	NotifierIds     []string `json:"notifierIds"`
	Operator        string   `json:"operator"`
	RangeMinutes    int64    `json:"rangeMinutes"`
	Threshold       float64  `json:"threshold"`
}

// Axiom API Reference: /v2/monitors
type MonitorsService service

// List all available monitors.
func (s *MonitorsService) List(ctx context.Context) ([]*Monitor, error) {
	ctx, span := s.client.trace(ctx, "Monitors.List")
	defer span.End()

	var res []*Monitor
	if err := s.client.Call(ctx, http.MethodGet, s.basePath, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return res, nil
}

// Get a monitor by id.
func (s *MonitorsService) Get(ctx context.Context, id string) (*Monitor, error) {
	ctx, span := s.client.trace(ctx, "Monitors.Get", trace.WithAttributes(
		attribute.String("axiom.Monitor_id", id),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, id)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res Monitor
	if err := s.client.Call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// Create a monitor with the given properties.
func (s *MonitorsService) Create(ctx context.Context, req Monitor) (*Monitor, error) {
	ctx, span := s.client.trace(ctx, "Monitors.Create", trace.WithAttributes(
		attribute.String("axiom.param.name", req.Name),
		attribute.String("axiom.param.description", req.Description),
	))
	defer span.End()

	var res Monitor
	if err := s.client.Call(ctx, http.MethodPost, s.basePath, req, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// Update the monitor identified by the given id with the given properties.
func (s *MonitorsService) Update(ctx context.Context, id string, req Monitor) (*Monitor, error) {
	ctx, span := s.client.trace(ctx, "Monitors.Update", trace.WithAttributes(
		attribute.String("axiom.monitor_id", id),
		attribute.String("axiom.param.description", req.Description),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, id)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res Monitor
	if err := s.client.Call(ctx, http.MethodPut, path, req, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// Delete the monitor identified by the given id.
func (s *MonitorsService) Delete(ctx context.Context, id string) error {
	ctx, span := s.client.trace(ctx, "Monitors.Delete", trace.WithAttributes(
		attribute.String("axiom.monitor_id", id),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, "/", id)
	if err != nil {
		return spanError(span, err)
	}

	if err := s.client.Call(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return spanError(span, err)
	}

	return nil
}
