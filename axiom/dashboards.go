package axiom

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

const dashboardAgainstTimestampFormat = "02 Jan 2006, 15:04"

// Dashboard represents a dashboard.
type Dashboard struct {
	// ID is the unique ID of the dashboard.
	ID string `json:"id"`
	// Name of the dashboard.
	Name string `json:"name"`
	// Description of the dashboard.
	Description string `json:"description"`
	// Owner is the team or user ID of the dashboards owner.
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
	// dashboard.
	TimeWindowStart string `json:"timeWindowStart"`
	// TimeWindowEnd is the end of the time window displayed by the dashboard.
	TimeWindowEnd string `json:"timeWindowEnd"`
	// Against specifies the time offset to compare the dashboards time window,
	// as specified by `TimeWindowStart` and `TimeWindowEnd`, against. This
	// field and `AgainstTimestamp` mutual exclude each other.
	Against time.Duration `json:"against"`
	// AgainstTimestamp is a timestamp that specifies the time offset to compare
	// the dashboards time window, as specified by `TimeWindowStart` and
	// `TimeWindowEnd`, against.  This field and `Against` mutual exclude each
	// other.
	AgainstTimestamp time.Time `json:"againstTimestamp"`
	// Version of the dashboard.
	Version string `json:"version"`
}

// MarshalJSON implements `json.Marshaler`. It is in place to marshal some
// fields to different representations for transport because that's what the
// server expects.
func (d Dashboard) MarshalJSON() ([]byte, error) {
	type LocalDash Dashboard
	localDash := struct {
		LocalDash

		Against          string `json:"against"`
		AgainstTimestamp string `json:"againstTimestamp"`
	}{
		LocalDash: LocalDash(d),
	}

	// Set to the value in seconds.
	localDash.RefreshTime = time.Duration(d.RefreshTime.Seconds())

	if d.Against != 0 {
		localDash.Against = d.Against.String()
	}

	// Format using the custom time format.
	if !d.AgainstTimestamp.IsZero() {
		localDash.AgainstTimestamp = d.AgainstTimestamp.Format(dashboardAgainstTimestampFormat)
	}

	return json.Marshal(localDash)
}

// UnmarshalJSON implements `json.Unmarshaler`. It is in place to unmarshal some
// fields from their transport representation into proper Go types because the
// server returns them differently.
func (d *Dashboard) UnmarshalJSON(b []byte) error {
	type LocalDash Dashboard
	localDash := struct {
		*LocalDash

		Against          string `json:"against"`
		AgainstTimestamp string `json:"againstTimestamp"`
	}{
		LocalDash: (*LocalDash)(d),
	}

	if err := json.Unmarshal(b, &localDash); err != nil {
		return err
	}

	// Set to a proper time.Duration value by interpreting the server response
	// value in seconds.
	d.RefreshTime = d.RefreshTime * time.Second

	var err error
	if localDash.Against != "" {
		if d.Against, err = time.ParseDuration(localDash.Against); err != nil {
			return err
		}
	}

	// Set to a proper time.Duration value by parsing the server response value
	// as a custom time format.
	if ts := localDash.AgainstTimestamp; ts != "" {
		if d.AgainstTimestamp, err = time.Parse(dashboardAgainstTimestampFormat, ts); err != nil {
			return err
		}
	}

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
	return s.client.call(ctx, http.MethodDelete, s.basePath+"/"+id, nil, nil)
}
