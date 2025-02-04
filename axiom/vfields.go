package axiom

import (
	"context"
	"net/http"
	"net/url"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type VirtualField struct {
	// Dataset is the dataset to which the virtual field belongs.
	Dataset string `json:"dataset"`
	// Name is the name of the virtual field.
	Name string `json:"name"`
	// Expression defines the virtual field's APL.
	Expression string `json:"expression"`
	// Description is an optional description of the virtual field.
	Description string `json:"description,omitempty"`
	// Type is the type of the virtual field. E.g. string | number
	Type string `json:"type,omitempty"`
	// Unit is the unit for the type of data returned by the virtual field.
	Unit string `json:"unit,omitempty"`
}

type VirtualFieldWithID struct {
	VirtualField
	// ID is the unique identifier of the virtual field.
	ID string `json:"id"`
}

// Axiom API Reference: /v2/vfields
type VirtualFieldsService service

// List all virtual fields for a given dataset.
func (s *VirtualFieldsService) List(ctx context.Context, dataset string) ([]*VirtualFieldWithID, error) {
	ctx, span := s.client.trace(ctx, "VirtualFields.List", trace.WithAttributes(
		attribute.String("axiom.param.dataset", dataset),
	))
	defer span.End()

	params := url.Values{}
	params.Set("dataset", dataset)

	var res []*VirtualFieldWithID
	if err := s.client.Call(ctx, http.MethodGet, s.basePath+"?"+params.Encode(), nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return res, nil
}

// Get a virtual field by id.
func (s *VirtualFieldsService) Get(ctx context.Context, id string) (*VirtualFieldWithID, error) {
	ctx, span := s.client.trace(ctx, "VirtualFields.Get", trace.WithAttributes(
		attribute.String("axiom.virtual_field_id", id),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, id)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res VirtualFieldWithID
	if err := s.client.Call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// Create a virtual field with the given properties.
func (s *VirtualFieldsService) Create(ctx context.Context, req VirtualField) (*VirtualFieldWithID, error) {
	ctx, span := s.client.trace(ctx, "VirtualFields.Create", trace.WithAttributes(
		attribute.String("axiom.param.dataset", req.Dataset),
		attribute.String("axiom.param.name", req.Name),
	))
	defer span.End()

	var res VirtualFieldWithID
	if err := s.client.Call(ctx, http.MethodPost, s.basePath, req, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// Update the virtual field identified by the given id with the given properties.
func (s *VirtualFieldsService) Update(ctx context.Context, id string, req VirtualField) (*VirtualFieldWithID, error) {
	ctx, span := s.client.trace(ctx, "VirtualFields.Update", trace.WithAttributes(
		attribute.String("axiom.virtual_field_id", id),
		attribute.String("axiom.param.dataset", req.Dataset),
		attribute.String("axiom.param.name", req.Name),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, id)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res VirtualFieldWithID
	if err := s.client.Call(ctx, http.MethodPut, path, req, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// Delete the virtual field identified by the given id.
func (s *VirtualFieldsService) Delete(ctx context.Context, id string) error {
	ctx, span := s.client.trace(ctx, "VirtualFields.Delete", trace.WithAttributes(
		attribute.String("axiom.virtual_field_id", id),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, id)
	if err != nil {
		return spanError(span, err)
	}

	if err := s.client.Call(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return spanError(span, err)
	}

	return nil
}
