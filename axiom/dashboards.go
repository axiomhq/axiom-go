package axiom

import (
	"bytes"
	"context"
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
func (s *DashboardsService) ListRaw(ctx context.Context, opts *DashboardsListOptions) ([]byte, error) {
	ctx, span := s.client.trace(ctx, "Dashboards.ListRaw")
	defer span.End()

	path, err := AddURLOptions(s.basePath, opts)
	if err != nil {
		return nil, spanError(span, err)
	}

	res, err := s.rawCall(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, spanError(span, err)
	}

	return res, nil
}

// GetRaw returns a dashboard as raw JSON.
func (s *DashboardsService) GetRaw(ctx context.Context, uid string) ([]byte, error) {
	ctx, span := s.client.trace(ctx, "Dashboards.GetRaw", trace.WithAttributes(
		attribute.String("axiom.dashboard_uid", uid),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, "uid", uid)
	if err != nil {
		return nil, spanError(span, err)
	}

	res, err := s.rawCall(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, spanError(span, err)
	}

	return res, nil
}

// CreateRaw creates or updates a dashboard from a raw JSON payload.
//
// The payload is sent as-is, which is useful when callers already have the
// request document in JSON form.
func (s *DashboardsService) CreateRaw(ctx context.Context, payload []byte) ([]byte, error) {
	ctx, span := s.client.trace(ctx, "Dashboards.CreateRaw")
	defer span.End()

	res, err := s.rawCall(ctx, http.MethodPost, s.basePath, payload)
	if err != nil {
		return nil, spanError(span, err)
	}

	return res, nil
}

// UpdateRaw updates a dashboard from a raw JSON payload.
//
// The payload is sent as-is, which is useful when callers already have the
// request document in JSON form.
func (s *DashboardsService) UpdateRaw(ctx context.Context, uid string, payload []byte) ([]byte, error) {
	ctx, span := s.client.trace(ctx, "Dashboards.UpdateRaw", trace.WithAttributes(
		attribute.String("axiom.dashboard_uid", uid),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, "uid", uid)
	if err != nil {
		return nil, spanError(span, err)
	}

	res, err := s.rawCall(ctx, http.MethodPut, path, payload)
	if err != nil {
		return nil, spanError(span, err)
	}

	return res, nil
}

func (s *DashboardsService) rawCall(ctx context.Context, method, path string, payload []byte) ([]byte, error) {
	var body any
	if payload != nil {
		body = bytes.NewReader(payload)
	}

	req, err := s.client.NewRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		req.Header.Set(headerContentType, mediaTypeJSON)
	}

	var buf bytes.Buffer
	if _, err := s.client.Do(req, &buf); err != nil {
		return nil, err
	}

	return append([]byte(nil), buf.Bytes()...), nil
}

// Delete the dashboard identified by uid.
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
