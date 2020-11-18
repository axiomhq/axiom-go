package axiom

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Dashboard represents a dashboard.
type Dashboard struct {
	// ID is the unique ID of the dashboard.
	ID string `json:"id"`
	// Name of the dashboard.
	Name string `json:"name"`
	// Description of the dashboard.
	Description string `json:"description"`
	// Owner is the ID of the dashboard owner.
	Owner string `json:"owner"`
	// Charts contains the raw data composing the dashboards charts.
	Charts []interface{} `json:"charts"`
	// Layout contains the raw data composing the dashboards layout.
	Layout []interface{} `json:"layout"`
	// RefreshTime is the duration after which the dashboards data is updated.
	RefreshTime time.Duration `json:"refreshTime"`
	// SchemaVersion auto increments with ever change made to the dashboard.
	SchemaVersion int `json:"schemaVersion"`
	// TimeWindowStart is the start of the time window displayed by the
	// dashboard. The format is special: It has the prefix "qr-now", followed
	// by a string duration. If the dashboard has a time range of "last 30
	// minutes", this will be: "qr-now-30m".
	TimeWindowStart string `json:"timeWindowStart"`
	// TimeWindowEnd is the end of the time window displayed by the dashboard.
	// The format is special: It has the prefix "qr-now", followed
	// by a string duration. If the dashboard has a time range of "last 30
	// minutes of yesterday", this will be: "qr-now-1d". But in most cases it
	// will "qr-now".
	TimeWindowEnd string `json:"timeWindowEnd"`
	// Version of the dashboard.
	Version string `json:"version"`
}

// MarshalJSON implements json.Marshaler. It is in place to marshal the
// RefreshTime to seconds because that's what the server expects.
func (d Dashboard) MarshalJSON() ([]byte, error) {
	type localDash Dashboard

	// Set to the value in seconds.
	d.RefreshTime = time.Duration(d.RefreshTime.Seconds())

	return json.Marshal(localDash(d))
}

// UnmarshalJSON implements json.Unmarshaler. It is in place to unmarshal the
// RefreshTime into a proper time.Duration value because the server returns it
// in seconds.
func (d *Dashboard) UnmarshalJSON(b []byte) error {
	type localDash *Dashboard

	if err := json.Unmarshal(b, localDash(d)); err != nil {
		return err
	}

	// Set to a proper time.Duration value interpreting the server response
	// value in seconds.
	d.RefreshTime = d.RefreshTime * time.Second

	return nil
}

// DashboardsService handles communication with the dashboard related operations
// of the Axiom API.
//
// Axiom API Reference: /api/v1/dashboards
type DashboardsService service

// List all available dashboards.
func (s *DashboardsService) List(ctx context.Context, opts ListOptions) ([]*Dashboard, error) {
	path, err := addOptions(s.basePath, opts)
	if err != nil {
		return nil, err
	}

	var res []*Dashboard
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
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
// When updating, the Version is mandantory and must be set to the current
// version of the dashboard as returned by a Get() call.
func (s *DashboardsService) Update(ctx context.Context, id string, req Dashboard) (*Dashboard, error) {
	path := s.basePath + "/" + id

	var res Dashboard
	if err := s.client.call(ctx, http.MethodPut, path, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Delete the dashboard identified by the given id.
func (s *DashboardsService) Delete(ctx context.Context, id string) error {
	path := s.basePath + "/" + id

	if err := s.client.call(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return err
	}

	return nil
}
