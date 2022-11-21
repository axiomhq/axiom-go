package query

import (
	"encoding/json"
	"fmt"
	"time"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=MessageCode,MessagePriority -linecomment -output=result_string.go

// MessageCode represents the code of a message associated with a query.
type MessageCode uint8

// All available message codes.
const (
	emptyMessageCode MessageCode = iota //

	VirtualFieldFinalizeError   // virtual_field_finalize_error
	MissingColumn               // missing_column
	LicenseLimitForQueryWarning // license_limit_for_query_warning
	DefaultLimitWarning         // default_limit_warning
)

func messageCodeFromString(s string) (mc MessageCode, err error) {
	switch s {
	case emptyMessageCode.String():
		mc = emptyMessageCode
	case VirtualFieldFinalizeError.String():
		mc = VirtualFieldFinalizeError
	case MissingColumn.String():
		mc = MissingColumn
	case LicenseLimitForQueryWarning.String():
		mc = LicenseLimitForQueryWarning
	case DefaultLimitWarning.String():
		mc = DefaultLimitWarning
	default:
		err = fmt.Errorf("unknown message code %q", s)
	}

	return mc, err
}

// MarshalJSON implements [json.Marshaler]. It is in place to marshal the
// message code to its string representation because that's what the server
// expects.
func (mc MessageCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(mc.String())
}

// UnmarshalJSON implements [json.Unmarshaler]. It is in place to unmarshal the
// message code from the string representation the server returns.
func (mc *MessageCode) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}

	*mc, err = messageCodeFromString(s)

	return err
}

// MessagePriority represents the priority of a message associated with a query.
type MessagePriority uint8

// All available message priorities.
const (
	emptyMessagePriority MessagePriority = iota //

	Trace // trace
	Debug // debug
	Info  // info
	Warn  // warn
	Error // error
	Fatal // fatal
)

func messagePriorityFromString(s string) (mp MessagePriority, err error) {
	switch s {
	case emptyMessagePriority.String():
		mp = emptyMessagePriority
	case Trace.String():
		mp = Trace
	case Debug.String():
		mp = Debug
	case Info.String():
		mp = Info
	case Warn.String():
		mp = Warn
	case Error.String():
		mp = Error
	case Fatal.String():
		mp = Fatal
	default:
		err = fmt.Errorf("unknown message priority %q", s)
	}

	return mp, err
}

// MarshalJSON implements [json.Marshaler]. It is in place to marshal the
// message priority to its string representation because that's what the server
// expects.
func (mp MessagePriority) MarshalJSON() ([]byte, error) {
	return json.Marshal(mp.String())
}

// UnmarshalJSON implements [json.Unmarshaler]. It is in place to unmarshal the
// message priority from the string representation the server returns.
func (mp *MessagePriority) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}

	*mp, err = messagePriorityFromString(s)

	return err
}

// Result is the result of an APL query.
type Result struct {
	// The datasets that were queried in order to create the result.
	Datasets []string `json:"datasetNames"`
	// Status of the query result.
	Status Status `json:"status"`
	// Matches are the events that matched the query.
	Matches []Entry `json:"matches"`
	// Buckets are the time series buckets.
	Buckets Timeseries `json:"buckets"`
	// GroupBy is a list of field names to group the query result by. Only valid
	// when at least one aggregation is specified.
	GroupBy []string `json:"-"`
	// SavedQueryID is the ID of the query that generated this result when it
	// was saved on the server. This is only set when the query was sent with
	// the `SaveKind` option specified.
	SavedQueryID string `json:"-"`
}

// Status is the status of a query result.
type Status struct {
	// ElapsedTime is the duration it took the query to execute.
	ElapsedTime time.Duration `json:"elapsedTime"`
	// BlocksExamined is the amount of blocks that have been examined by the
	// query.
	BlocksExamined uint64 `json:"blocksExamined"`
	// RowsExamined is the amount of rows that have been examined by the query.
	RowsExamined uint64 `json:"rowsExamined"`
	// RowsMatched is the amount of rows that matched the query.
	RowsMatched uint64 `json:"rowsMatched"`
	// NumGroups is the amount of groups returned by the query.
	NumGroups uint32 `json:"numGroups"`
	// IsPartial describes if the query result is a partial result.
	IsPartial bool `json:"isPartial"`
	// ContinuationToken is populated when IsPartial is true and must be passed
	// to the next query request to retrieve the next result set.
	ContinuationToken string `json:"continuationToken"`
	// IsEstimate describes if the query result is estimated.
	IsEstimate bool `json:"isEstimate"`
	// MinBlockTime is the timestamp of the oldest block examined.
	MinBlockTime time.Time `json:"minBlockTime"`
	// MaxBlockTime is the timestamp of the newest block examined.
	MaxBlockTime time.Time `json:"maxBlockTime"`
	// Messages associated with the query.
	Messages []Message `json:"messages"`
	// MinCursor is the id of the oldest row, as seen server side. May be lower
	// than what the results include if the server scanned more data than
	// included in the results. Can be used to efficiently resume time-sorted
	// non-aggregating queries (i.e. filtering only).
	MinCursor string `json:"minCursor"`
	// MaxCursor is the id of the newest row, as seen server side. May be higher
	// than what the results include if the server scanned more data than
	// included in the results. Can be used to efficiently resume time-sorted
	// non-aggregating queries (i.e. filtering only).
	MaxCursor string `json:"maxCursor"`
}

// MarshalJSON implements [json.Marshaler]. It is in place to marshal the
// elapsed time into its microsecond representation because that's what the
// server expects.
func (s Status) MarshalJSON() ([]byte, error) {
	type localStatus Status

	// Set to the value in microseconds.
	s.ElapsedTime = time.Duration(s.ElapsedTime.Microseconds())

	return json.Marshal(localStatus(s))
}

// UnmarshalJSON implements [json.Unmarshaler]. It is in place to unmarshal the
// elapsed time into a proper time.Duration value because the server returns it
// in microseconds.
func (s *Status) UnmarshalJSON(b []byte) error {
	type localStatus *Status

	if err := json.Unmarshal(b, localStatus(s)); err != nil {
		return err
	}

	// Set to a proper time.Duration value by interpreting the server response
	// value in microseconds.
	s.ElapsedTime = s.ElapsedTime * time.Microsecond

	return nil
}

// Message is a message associated with a query result.
type Message struct {
	// Priority of the message.
	Priority MessagePriority `json:"priority"`
	// Code of the message.
	Code MessageCode `json:"code"`
	// Count describes how often a message of this type was raised by the query.
	Count uint `json:"count"`
	// Text is a human readable text representation of the message.
	Text string `json:"msg"`
}

// Entry is an event that matched a query and is thus part of the result set.
type Entry struct {
	// Time is the time the event occurred. Matches SysTime if not specified
	// during ingestion.
	Time time.Time `json:"_time"`
	// SysTime is the time the event was recorded on the server.
	SysTime time.Time `json:"_sysTime"`
	// RowID is the unique ID of the event row.
	RowID string `json:"_rowId"`
	// Data contains the raw data of the event (with filters and aggregations
	// applied).
	Data map[string]any `json:"data"`
}

// Timeseries are queried time series.
type Timeseries struct {
	// Series are the intervals that build a time series.
	Series []Interval `json:"series"`
	// Totals of the time series.
	Totals []EntryGroup `json:"totals"`
}

// Interval is the interval of queried time series.
type Interval struct {
	// StartTime of the interval.
	StartTime time.Time `json:"startTime"`
	// EndTime of the interval.
	EndTime time.Time `json:"endTime"`
	// Groups of the interval.
	Groups []EntryGroup `json:"groups"`
}

// EntryGroup is a group of queried event.
type EntryGroup struct {
	// ID is the unique the group.
	ID uint64 `json:"id"`
	// Group maps the fieldnames to the unique values for the entry.
	Group map[string]any `json:"group"`
	// Aggregations of the group.
	Aggregations []EntryGroupAgg `json:"aggregations"`
}

// EntryGroupAgg is an aggregation which is part of a group of queried events.
type EntryGroupAgg struct {
	// Alias is the aggregations alias. If it wasn't specified at query time, it
	// is the uppercased string representation of the aggregation operation.
	Alias string `json:"op"`
	// Value is the result value of the aggregation.
	Value any `json:"value"`
}
