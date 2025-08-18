package axiom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

//go:generate go tool stringer -type=Operator,MonitorType -linecomment -output=monitors_string.go

// Operator represents a comparison operation for a monitor. A [Monitor] acts on
// the result of comparing a query result with a threshold.
type Operator uint8

// All available [Monitor] comparison [Operator]s.
const (
	emptyOperator Operator = iota //

	Below        // Below
	BelowOrEqual // BelowOrEqual
	Above        // Above
	AboveOrEqual // AboveOrEqual
	AboveOrBelow // AboveOrBelow
)

func operatorFromString(s string) (c Operator, err error) {
	switch s {
	case emptyOperator.String():
		c = emptyOperator
	case Below.String():
		c = Below
	case BelowOrEqual.String():
		c = BelowOrEqual
	case Above.String():
		c = Above
	case AboveOrEqual.String():
		c = AboveOrEqual
	case AboveOrBelow.String():
		c = AboveOrBelow
	default:
		err = fmt.Errorf("unknown operator %q", s)
	}

	return c, err
}

// MarshalJSON implements [json.Marshaler]. It is in place to marshal the
// Operator to its string representation because that's what the server
// expects.
func (c Operator) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

// UnmarshalJSON implements [json.Unmarshaler]. It is in place to unmarshal the
// Operator from the string representation the server returns.
func (c *Operator) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}

	*c, err = operatorFromString(s)

	return err
}

// MonitorType represents the type of the monitor.
type MonitorType uint8

// All available [MonitorTypes]s.
const (
	MonitorTypeThreshold        MonitorType = iota // Threshold
	MonitorTypeMatchEvent                          // MatchEvent
	MonitorTypeAnomalyDetection                    // AnomalyDetection
)

func typeFromString(s string) (c MonitorType) {
	switch s {
	case MonitorTypeMatchEvent.String():
		return MonitorTypeMatchEvent
	case MonitorTypeAnomalyDetection.String():
		return MonitorTypeAnomalyDetection
	default:
		return MonitorTypeThreshold
	}
}

// MarshalJSON implements [json.Marshaler]. It is in place to marshal the
// MonitorType to its string representation because that's what the server
// expects.
func (c MonitorType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

// UnmarshalJSON implements [json.Unmarshaler]. It is in place to unmarshal the
// MonitorType from the string representation the server returns.
func (c *MonitorType) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}

	*c = typeFromString(s)

	return nil
}

type Monitor struct {
	// ID is the unique ID of the monitor.
	ID string `json:"id,omitempty"`
	// CreatedAt is the time when the monitor was created.
	CreatedAt time.Time `json:"createdAt"`
	// CreatedBy is the ID of the user who created the monitor.
	CreatedBy string `json:"createdBy"`
	// Name is the name of the monitor.
	Name string `json:"name"`
	// Type sets the type of monitor. Defaults to [Threshold]
	Type MonitorType `json:"type"`
	// Description of the monitor.
	Description string `json:"description,omitempty"`

	// APLQuery is the APL query to use for the monitor.
	APLQuery string `json:"aplQuery"`
	// Operator is the operator to use for the monitor.
	Operator Operator `json:"operator"`
	// Threshold the query result is compared against, which evaluates if the
	// monitor acts or not.
	Threshold float64 `json:"threshold"`
	// AlertOnNoData indicates whether to alert on no data.
	AlertOnNoData bool `json:"alertOnNoData"`
	// NotifyByGroup tracks each non-time group independently.
	NotifyByGroup bool `json:"notifyByGroup"`
	// Resolvable determines whether the events triggered by the monitor are
	// resolvable. This has no effect on threshold monitors.
	Resolvable bool `json:"resolvable"`
	// NotifierIDs attached to the monitor.
	NotifierIDs []string `json:"notifierIDs"`

	// Interval is the interval in minutes in which the monitor will run. The
	// lowest resolution is 1 minute.
	Interval time.Duration `json:"intervalMinutes"`
	// Range the monitor goes back in time and looks at the data it acts on. The
	// lowest resolution is 1 minute.
	Range time.Duration `json:"rangeMinutes"`
	// Disabled sets whether the monitor is disabled or not.
	Disabled bool `json:"disabled"`
	// DisabledUntil is the time that the monitor will be disabled until.
	DisabledUntil time.Time `json:"disabledUntil"`
	// Delay after the end time that the monitor runs. The lowest resolution is
	// 1 second.
	Delay time.Duration `json:"secondDelay,omitempty"`
	// NotifyEveryRun indicates whether to notify on every trigger.
	NotifyEveryRun bool `json:"notifyEveryRun,omitempty"`
	// SkipResolved indicates whether to skip resolved alerts.
	SkipResolved bool `json:"skipResolved,omitempty"`
	// Tolerance is the acceptable tolerance for an anomaly detection monitor.
	Tolerance float64 `json:"tolerance,omitempty"`
	// TriggerFromNRuns indicates the number of runs to consider when
	// determining whether the current run should notify.
	TriggerFromNRuns int64 `json:"triggerFromNRuns,omitempty"`
	// TriggerAfterNPositiveResults indicates the number of positive results
	// required to trigger a notification (can be from a larger set of TriggerFromNRuns)
	TriggerAfterNPositiveResults int64 `json:"triggerAfterNPositiveResults,omitempty"`
	// CompareDays indicates the number of days to compare for anomaly detection
	// only.
	CompareDays int64 `json:"compareDays,omitempty"`
}

// MarshalJSON implements [json.Marshaler]. It is in place to marshal the
// Range and Interval to minutes because that's what the
// server expects as well as setting the appropriate query type.
func (m Monitor) MarshalJSON() ([]byte, error) {
	type localMonitor Monitor

	// Set to the value in minutes.
	m.Range = time.Duration(m.Range.Minutes())
	m.Interval = time.Duration(m.Interval.Minutes())
	m.Delay = time.Duration(m.Delay.Seconds())

	return json.Marshal(localMonitor(m))
}

// UnmarshalJSON implements [json.Unmarshaler]. It is in place to convert the
// Range and Interval field values into proper
// time.Duration values because the server returns them in seconds as well as
// unmarshalling the query in to its appropriate type.
func (m *Monitor) UnmarshalJSON(b []byte) error {
	type LocalMonitor Monitor
	localMonitor := struct {
		*LocalMonitor
	}{
		LocalMonitor: (*LocalMonitor)(m),
	}
	if err := json.Unmarshal(b, &localMonitor); err != nil {
		return err
	}

	// Set to a proper time.Duration value by interpreting the server response
	// value in seconds.
	m.Range *= time.Minute
	m.Interval *= time.Minute
	m.Delay *= time.Second

	return nil
}

type MonitorCreateRequest struct {
	Monitor
}

type MonitorUpdateRequest struct {
	Monitor
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
		attribute.String("axiom.monitor_id", id),
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
func (s *MonitorsService) Create(ctx context.Context, req MonitorCreateRequest) (*Monitor, error) {
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
func (s *MonitorsService) Update(ctx context.Context, id string, req MonitorUpdateRequest) (*Monitor, error) {
	ctx, span := s.client.trace(ctx, "Monitors.Update", trace.WithAttributes(
		attribute.String("axiom.monitor_id", id),
		attribute.String("axiom.param.name", req.Name),
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

	path, err := url.JoinPath(s.basePath, id)
	if err != nil {
		return spanError(span, err)
	}

	if err := s.client.Call(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return spanError(span, err)
	}

	return nil
}
