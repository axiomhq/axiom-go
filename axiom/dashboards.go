package axiom

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DashboardsService handles communication with dashboard related operations of
// the Axiom API.
//
// Axiom API Reference: /v2/dashboards
type DashboardsService service

// DashboardsListOptions configures optional pagination query parameters for
// [DashboardsService.ListRaw].
type DashboardsListOptions struct {
	Limit  int `url:"limit,omitempty"`
	Offset int `url:"offset,omitempty"`
}

// ListRaw returns dashboards as raw JSON.
func (s *DashboardsService) ListRaw(ctx context.Context, opts *DashboardsListOptions) (json.RawMessage, error) {
	ctx, span := s.client.trace(ctx, "Dashboards.ListRaw")
	defer span.End()

	path, err := AddURLOptions(s.basePath, opts)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res json.RawMessage
	if err := s.client.Call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return res, nil
}

// GetRaw returns a dashboard as raw JSON.
func (s *DashboardsService) GetRaw(ctx context.Context, uid string) (json.RawMessage, error) {
	ctx, span := s.client.trace(ctx, "Dashboards.GetRaw", trace.WithAttributes(
		attribute.String("axiom.dashboard_uid", uid),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, "uid", uid)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res json.RawMessage
	if err := s.client.Call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return res, nil
}

// CreateRaw creates or updates a dashboard from a raw JSON payload.
//
// The payload is sent as-is, which is useful when callers already have the
// request document in JSON form.
func (s *DashboardsService) CreateRaw(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	ctx, span := s.client.trace(ctx, "Dashboards.CreateRaw")
	defer span.End()

	var res json.RawMessage
	if err := s.client.Call(ctx, http.MethodPost, s.basePath, payload, &res); err != nil {
		return nil, spanError(span, err)
	}

	return res, nil
}

// UpdateRaw updates a dashboard from a raw JSON payload.
//
// The payload is sent as-is, which is useful when callers already have the
// request document in JSON form.
func (s *DashboardsService) UpdateRaw(ctx context.Context, uid string, payload json.RawMessage) (json.RawMessage, error) {
	ctx, span := s.client.trace(ctx, "Dashboards.UpdateRaw", trace.WithAttributes(
		attribute.String("axiom.dashboard_uid", uid),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, "uid", uid)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res json.RawMessage
	if err := s.client.Call(ctx, http.MethodPut, path, payload, &res); err != nil {
		return nil, spanError(span, err)
	}

	return res, nil
}

// Delete the dashboard identified by the given uid.
func (s *DashboardsService) Delete(ctx context.Context, uid string) error {
	ctx, span := s.client.trace(ctx, "Dashboards.Delete", trace.WithAttributes(
		attribute.String("axiom.dashboard_uid", uid),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, "uid", uid)
	if err != nil {
		return spanError(span, err)
	}

	if err := s.client.Call(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return spanError(span, err)
	}

	return nil
}
