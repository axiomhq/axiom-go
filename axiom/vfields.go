package axiom

import (
	"context"
	"net/http"
)

// VirtualField represents a virtual field of a dataset.
type VirtualField struct {
	// ID is the unique id of the virtual field.
	ID string `json:"id"`
	// Dataset the virtual field belongs to.
	Dataset string `json:"dataset"`
	// Name is the display name of the virtual field.
	Name string `json:"name"`
	// Description of the virtual field.
	Description string `json:"description"`
	// Alias the virtual field is referenced by.
	Alias string `json:"alias"`
	// Expression that evaluates the virtual fields value.
	Expression string `json:"expression"`
}

// VirtualFieldListOptions specifies the parameters for the List method of the
// VirtualFields service.
type VirtualFieldListOptions struct {
	// Dataset to list virtual fields for. Required.
	Dataset string `url:"dataset"`

	ListOptions
}

// VirtualFieldsService handles communication with the virtual field related operations of the
// Axiom API.
//
// Axiom API Reference: /api/v1/vfields
type VirtualFieldsService service

// List all available virtual fields.
func (s *VirtualFieldsService) List(ctx context.Context, opts VirtualFieldListOptions) ([]*VirtualField, error) {
	path, err := addOptions(s.basePath, opts)
	if err != nil {
		return nil, err
	}

	var res []*VirtualField
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// Get a virtual field by id.
func (s *VirtualFieldsService) Get(ctx context.Context, id string) (*VirtualField, error) {
	path := s.basePath + "/" + id

	var res VirtualField
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Create a virtual field with the given properties.
func (s *VirtualFieldsService) Create(ctx context.Context, req VirtualField) (*VirtualField, error) {
	var res VirtualField
	if err := s.client.call(ctx, http.MethodPost, s.basePath, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Update the virtual field identified by the given id with the given properties.
func (s *VirtualFieldsService) Update(ctx context.Context, id string, req VirtualField) (*VirtualField, error) {
	path := s.basePath + "/" + id

	var res VirtualField
	if err := s.client.call(ctx, http.MethodPut, path, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Delete the virtual field identified by the given id.
func (s *VirtualFieldsService) Delete(ctx context.Context, id string) error {
	path := s.basePath + "/" + id

	if err := s.client.call(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return err
	}

	return nil
}
