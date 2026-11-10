package axiom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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
	MonitorTypeNewValues                           // NewValues
)

func typeFromString(s string) (c MonitorType) {
	switch s {
	case MonitorTypeMatchEvent.String():
		return MonitorTypeMatchEvent
	case MonitorTypeAnomalyDetection.String():
		return MonitorTypeAnomalyDetection
	case MonitorTypeNewValues.String():
		return MonitorTypeNewValues
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

// MonitorSeverity represents the severity of a [MonitorThreshold] level.
type MonitorSeverity string

// All available [MonitorSeverity]s.
const (
	MonitorSeverityInfo     MonitorSeverity = "info"
	MonitorSeverityWarning  MonitorSeverity = "warning"
	MonitorSeverityError    MonitorSeverity = "error"
	MonitorSeverityCritical MonitorSeverity = "critical"
)

// MonitorThreshold is one named level of a [MonitorTypeThreshold] or
// [MonitorTypeAnomalyDetection] monitor. All levels share the monitor's query
// and [Operator].
type MonitorThreshold struct {
	// Name of the level, shown in notifications.
	Name string `json:"name,omitempty"`
	// Severity of the level, sent to PagerDuty and Opsgenie.
	Severity MonitorSeverity `json:"severity,omitempty"`
	// Value the query result must cross, by the monitor's operator, to reach
	// the level. On an anomaly detection monitor, it is a tolerance like
	// [Monitor.Tolerance]: the number of standard deviations from the
	// expected value, above 0 and at most 100.
	Value float64 `json:"value"`
	// Recovery is the value the query result must cross back to leave the
	// level. Defaults to Value when nil. On an anomaly detection monitor, it
	// is a tolerance from 0 to Value.
	Recovery *float64 `json:"recovery,omitempty"`
	// NotifierIDs are the notifiers of the level. Empty means the monitor's
	// [Monitor.NotifierIDs]. An alert notifies the notifiers of its level and
	// of every less severe one. A notifier gets a resolve when the alert drops
	// below the least severe level that lists it, unless
	// [Monitor.SkipResolved] is set. A level with notifiers is a named level,
	// and setting them must be enabled for the organization. An update that
	// sends the stored levels back unchanged, in the order the server returned
	// them, needs no flag.
	NotifierIDs []string `json:"notifierIds,omitempty"`
}

type Monitor struct {
	// ID is the unique ID of the monitor.
	ID string `json:"id,omitempty"`
	// CreatedAt is the time when the monitor was created.
	CreatedAt time.Time `json:"createdAt"`
	// CreatedBy is the ID of the user who created the monitor.
	CreatedBy string `json:"createdBy"`
	// UpdatedAt is the time when the monitor was last updated.
	UpdatedAt time.Time `json:"updatedAt"`
	// Name is the name of the monitor.
	Name string `json:"name"`
	// Type sets the type of monitor. Defaults to [Threshold]
	Type MonitorType `json:"type"`
	// Description of the monitor.
	Description string `json:"description,omitempty"`

	// APLQuery is the APL query to use for the monitor.
	APLQuery string `json:"aplQuery"`
	// MPLQuery is the MPL query to use for the monitor.
	MPLQuery string `json:"mplQuery,omitempty"`
	// Operator is the operator to use for the monitor.
	Operator Operator `json:"operator"`
	// Threshold the query result is compared against, which evaluates if the
	// monitor acts or not. On a threshold monitor with Thresholds, the server
	// sets it to the value of the most severe level. An update without
	// Thresholds on a monitor with more than one level fails with a 400
	// response; send the levels back, or remove them with an empty Thresholds.
	Threshold float64 `json:"threshold"`
	// Thresholds are the levels of a [MonitorTypeThreshold] or
	// [MonitorTypeAnomalyDetection] monitor, from least to most severe. On a
	// threshold monitor, level Values take the place of Threshold. On an
	// anomaly detection monitor, they take the place of Tolerance and must
	// increase with severity whatever the [Operator]. The rest of this comment
	// calls that field the monitor's value.
	//
	// The server returns a monitor without levels as a single level with
	// only a Value, equal to the monitor's value. While that level has no
	// Severity, Recovery or NotifierIDs, the server ignores its Value; change
	// the monitor's value instead. Setting any of them makes it a named
	// level, which needs a Name and a Severity. More than one level must be
	// enabled for the organization.
	//
	// A nil Thresholds is not sent. On update, that moves a single stored
	// level to the monitor's value and fails on more than one. A non-nil
	// empty Thresholds, such as []MonitorThreshold{}, is sent as an empty
	// list and removes all levels, leaving the monitor's value as the only
	// one, so set it with the empty list.
	Thresholds []MonitorThreshold `json:"thresholds,omitzero"`
	// AlertOnNoData indicates whether to alert on no data.
	AlertOnNoData bool `json:"alertOnNoData"`
	// NotifyByGroup tracks each non-time group independently.
	NotifyByGroup bool `json:"notifyByGroup"`
	// Resolvable determines whether the events triggered by the monitor are
	// resolvable. This has no effect on threshold monitors.
	Resolvable bool `json:"resolvable"`
	// NotifierIDs attached to the monitor. When any of Thresholds sets its
	// own NotifierIDs, these only notify for levels without them.
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
	// IncludeChart indicates whether to include a query chart in triggered
	// notifications. This only has an effect on threshold monitors.
	IncludeChart bool `json:"includeChart,omitempty"`
	// Tolerance is the acceptable tolerance for an anomaly detection monitor:
	// the number of standard deviations from the expected value that triggers
	// the monitor, or its most severe level. On a monitor with Thresholds, the
	// server sets it to the value of the most severe level. Updates treat it as
	// they treat Threshold on a threshold monitor; see Thresholds.
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
	// TrackedValue is the column of the query result whose values a new values
	// monitor checks for ones it has not seen. Required for new values
	// monitors only.
	TrackedValue string `json:"trackedValue,omitempty"`
	// LookbackDays is the number of days a new values monitor remembers a
	// value: a value counts as new when it was not seen in this many days.
	// Required for new values monitors only.
	LookbackDays int64 `json:"lookbackDays,omitempty"`
	// BackfillDays is the number of days of values a new values monitor
	// records as seen before its first run, so they do not alert. It must not
	// exceed LookbackDays, and 0 disables the backfill. New values monitors
	// only.
	BackfillDays int64 `json:"backfillDays,omitempty"`
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

func validateMonitorQueries(m Monitor) error {
	hasAPL := strings.TrimSpace(m.APLQuery) != ""
	hasMPL := strings.TrimSpace(m.MPLQuery) != ""

	if hasAPL == hasMPL {
		if hasAPL {
			return fmt.Errorf("aplQuery and mplQuery are mutually exclusive, provide only one")
		}

		return fmt.Errorf("one of aplQuery or mplQuery is required")
	}

	return nil
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

	if err := validateMonitorQueries(req.Monitor); err != nil {
		return nil, spanError(span, err)
	}

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

	if err := validateMonitorQueries(req.Monitor); err != nil {
		return nil, spanError(span, err)
	}

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
