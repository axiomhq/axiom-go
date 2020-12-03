package axiom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/axiomhq/axiom-go/axiom/query"
)

//go:generate ../bin/stringer -type=Comparison -output=monitors_string.go

// Comparison represents a comparison operation for a monitor. A monitor acts on
// the result of comparing a query result with a threshold.
type Comparison uint8

// All available monitor comparison modes.
const (
	Below Comparison = iota + 1
	BelowOrEqual
	Above
	AboveOrEqual
)

// MarshalJSON implements json.Marshaler. It is in place to marshal the
// Comparison to its string representation because that's what the server
// expects.
func (c Comparison) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

// UnmarshalJSON implements json.Unmarshaler. It is in place to unmarshal the
// Comparison from the string representation the server returns.
func (c *Comparison) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch s {
	case Below.String():
		*c = Below
	case BelowOrEqual.String():
		*c = BelowOrEqual
	case Above.String():
		*c = Above
	case AboveOrEqual.String():
		*c = AboveOrEqual
	default:
		return fmt.Errorf("unknown comparison %q", s)
	}

	return nil
}

// A Monitor continuesly runs a query on a dataset and evaluates its result
// against a configured threshold. Upon reaching those it will invoke the
// configured notifiers.
type Monitor struct {
	// ID is the unique id of the monitor.
	ID string `json:"id"`
	// Dataset the monitor belongs to.
	Dataset string `json:"dataset"`
	// Name is the display name of the monitor.
	Name string `json:"name"`
	// Description of the monitor.
	Description string `json:"description"`
	// DisabledUntil is the time until the monitor is being executed again.
	DisabledUntil time.Time `json:"disabledUntil"`
	// Query is executed by the monitor and the result is compared using the
	// monitors configured comparison operator against the configured threshold.
	Query query.Query `json:"query"`
	// Threshold the query result is compared against, which evalutes if the
	// monitor acts or not.
	Threshold float64 `json:"threshold"`
	// Comparison operator to use for comparing the query result and threshold
	// value.
	Comparison Comparison `json:"comparison"`
	// NoDataCloseWait specifies after which amount of laking a query result,
	// the monitor is closed.
	NoDataCloseWait time.Duration `json:"noDataCloseWaitMinutes"`
	// Frequency the monitor is executed by.
	Frequency time.Duration `json:"frequencyMinutes"`
	// Duration the monitor goes back in time and looks at the data it acts on.
	Duration time.Duration `json:"durationMinutes"`
	// Notifiers attached to the monitor.
	Notifiers []string `json:"notifiers"`
	// LastCheckTime specifies the last time the monitor executed the query and
	// compared its result against the threshold.
	LastCheckTime time.Time `json:"lastCheckTime"`
	// LastCheckState is the state associated with the last monitor execution.
	LastCheckState map[string]string `json:"lastCheckState"`
}

// MarshalJSON implements json.Marshaler. It is in place to marshal the
// NoDataCloseWait, Frequency and Duration to minutes because that's what the
// server expects.
func (m Monitor) MarshalJSON() ([]byte, error) {
	type localMonitor Monitor

	// Set to the value in minutes.
	m.NoDataCloseWait = time.Duration(m.NoDataCloseWait.Minutes())
	m.Frequency = time.Duration(m.Frequency.Minutes())
	m.Duration = time.Duration(m.Duration.Minutes())

	return json.Marshal(localMonitor(m))
}

// UnmarshalJSON implements json.Unmarshaler. It is in place to unmarshal the
// RefreshTime into a proper time.Duration value because the server returns it
// in seconds.
func (m *Monitor) UnmarshalJSON(b []byte) error {
	type localMonitor *Monitor

	if err := json.Unmarshal(b, localMonitor(m)); err != nil {
		return err
	}

	// Set to a proper time.Duration value interpreting the server response
	// value in seconds.
	m.NoDataCloseWait = m.NoDataCloseWait * time.Minute
	m.Frequency = m.Frequency * time.Minute
	m.Duration = m.Duration * time.Minute

	return nil
}

// MonitorsService handles communication with the monitor related operations of
// the Axiom API.
//
// Axiom API Reference: /api/v1/monitors
type MonitorsService service

// List all available monitors.
func (s *MonitorsService) List(ctx context.Context) ([]*Monitor, error) {
	var res []*Monitor
	if err := s.client.call(ctx, http.MethodGet, s.basePath, nil, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// Get a monitor by id.
func (s *MonitorsService) Get(ctx context.Context, id string) (*Monitor, error) {
	path := s.basePath + "/" + id

	var res Monitor
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Create a monitor with the given properties.
func (s *MonitorsService) Create(ctx context.Context, req Monitor) (*Monitor, error) {
	var res Monitor
	if err := s.client.call(ctx, http.MethodPost, s.basePath, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Update the monitor identified by the given id with the given properties.
func (s *MonitorsService) Update(ctx context.Context, id string, req Monitor) (*Monitor, error) {
	path := s.basePath + "/" + id

	var res Monitor
	if err := s.client.call(ctx, http.MethodPut, path, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Delete the monitor identified by the given id.
func (s *MonitorsService) Delete(ctx context.Context, id string) error {
	path := s.basePath + "/" + id

	if err := s.client.call(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return err
	}

	return nil
}
