package axiom

import (
	"context"
	"net/http"
)

// Dashboard represents an Axiom dashboard.
type Dashboard struct {
	// ID is the unique ID of the dashboard.
	ID string `json:"id"`
	// Name of the dashboard.
	Name string `json:"name"`
	// Description of the dashboard.
	Description string `json:"description"`
	// Owner of the dashboard.
	Owner string `json:"owner"`
	// Charts contains the raw data composing the dashboards charts.
	Charts []interface{} `json:"charts"`
	// Layout contains the raw data composing the dashboards layout.
	Layout []interface{} `json:"layout"`
	// RefreshTime is the duration in seconds after which the dashboard is
	// updated.
	RefreshTime     int    `json:"refreshTime"`
	SchemaVersion   int    `json:"schemaVersion"`
	TimeWindowStart string `json:"timeWindowStart"`
	TimeWindowEnd   string `json:"timeWindowEnd"`
	// Version of the dashboard.
	Version string `json:"version"`
}

// DashboardsService handles communication with the dashboard related operations
// of the Axiom API.
//
// Axiom API Reference: /api/v1/dashboards
type DashboardsService service

// List all available dashboards.
func (s *DashboardsService) List(ctx context.Context) ([]*Dashboard, error) {
	var res []*Dashboard
	if err := s.client.call(ctx, http.MethodGet, s.basePath, nil, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// Get a dashboard by id.
func (s *DashboardsService) Get(ctx context.Context, id string) (*Dashboard, error) {
	path := s.basePath + "/" + id

	var res Dashboard
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Create a dashboard with the given properties. The ID and Version fields of
// the request payload are ignored.
func (s *DashboardsService) Create(ctx context.Context, req Dashboard) (*Dashboard, error) {
	var res Dashboard
	if err := s.client.call(ctx, http.MethodPost, s.basePath, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Update the dashboard identified by the given id with the given properties.
// func (s *DashboardsService) Update(ctx context.Context, id string, req Dashboard) (*Dashboard, error) {
// 	path := s.basePath + "/" + id

// 	var res Dashboard
// 	if err := s.client.call(ctx, http.MethodPut, path, req, &res); err != nil {
// 		return nil, err
// 	}

// 	return &res, nil
// }

// Delete the dashboard identified by the given id.
func (s *DashboardsService) Delete(ctx context.Context, id string) error {
	path := s.basePath + "/" + id

	if err := s.client.call(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return err
	}

	return nil
}
