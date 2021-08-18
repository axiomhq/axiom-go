package query

import (
	"encoding/json"
	"fmt"
	"time"
)

//go:generate ../../bin/stringer -type=MessageCode,MessagePriority -linecomment -output=result_string.go

// MessageCode represents the code of a message associated with a query.
type MessageCode uint8

// All available message codes.
const (
	VirtualFieldFinalizeError   MessageCode = iota + 1 // virtual_field_finalize_error
	MissingColumn                                      // missing_column
	LicenseLimitForQueryWarning                        // license_limit_for_query_warning
	DefaultLimitWarning                                // default_limit_warning
)

// UnmarshalJSON implements json.Unmarshaler. It is in place to unmarshal the
// MessageCode from the string representation the server returns.
func (mc *MessageCode) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch s {
	case VirtualFieldFinalizeError.String():
		*mc = VirtualFieldFinalizeError
	case MissingColumn.String():
		*mc = MissingColumn
	default:
		return fmt.Errorf("unknown message code %q", s)
	}

	return nil
}

// MessagePriority represents the priority of a message associated with a query.
type MessagePriority uint8

// All available message priorities.
const (
	Trace MessagePriority = iota + 1 // trace
	Debug                            // debug
	Info                             // info
	Warn                             // warn
	Error                            // error
	Fatal                            // fatal
)

// UnmarshalJSON implements json.Unmarshaler. It is in place to unmarshal the
// MessagePriority from the string representation the server returns.
func (mp *MessagePriority) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch s {
	case Trace.String():
		*mp = Trace
	case Debug.String():
		*mp = Debug
	case Info.String():
		*mp = Info
	case Warn.String():
		*mp = Warn
	case Error.String():
		*mp = Error
	case Fatal.String():
		*mp = Fatal
	default:
		return fmt.Errorf("unknown message priority %q", s)
	}

	return nil
}

// Result is the result of a query.
type Result struct {
	// Status of the query result.
	Status Status `json:"status"`
	// Matches are the events that matched the query.
	Matches []Entry `json:"matches"`
	// Buckets are the time series buckets.
	Buckets Timeseries `json:"buckets"`
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
}

// MarshalJSON implements json.Marshaler. It is in place to marshal the
// ElapsedTime into its microsecond representation because that's what the
// server expects.
func (s Status) MarshalJSON() ([]byte, error) {
	type localStatus Status

	// Set to the value in microseconds.
	s.ElapsedTime = time.Duration(s.ElapsedTime.Microseconds())

	return json.Marshal(localStatus(s))
}

// UnmarshalJSON implements json.Unmarshaler. It is in place to unmarshal the
// ElapsedTime into a proper time.Duration value because the server returns it
// in microseconds.
func (s *Status) UnmarshalJSON(b []byte) error {
	type localStatus *Status

	if err := json.Unmarshal(b, localStatus(s)); err != nil {
		return err
	}

	// Set to a proper time.Duration value interpreting the server response
	// value in microseconds.
	s.ElapsedTime = s.ElapsedTime * time.Microsecond

	return nil
}

// Message is a message associated with a query result.
type Message struct {
	// Priority of the message.
	Priority MessagePriority `json:"priority"`
	// Count describes how often a message of this type was raised by the query.
	Count uint `json:"count"`
	// Code of the message.
	Code MessageCode `json:"code"`
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
	Data map[string]interface{} `json:"data"`
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
	Group map[string]interface{} `json:"group"`
	// Aggregations of the group.
	Aggregations []EntryGroupAgg `json:"aggregations"`
}

// EntryGroupAgg is an aggregation which is part of a group of queried events.
type EntryGroupAgg struct {
	// Op is the aggregations operation.
	Op AggregationOp `json:"op"`
	// Value is the result value of the aggregation.
	Value interface{} `json:"value"`
}
