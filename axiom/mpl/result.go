package mpl

import "time"

// Result is the result of an MPL query.
type Result struct {
	// Metadata of the result.
	Metadata Metadata
	// Series in the result.
	Series []Series
	// TraceID is the ID of the trace that the server generated for the query
	// request.
	TraceID string
}

// Metadata of an MPL query [Result].
type Metadata struct {
	// GroupKeys are the tags that the query groups the series by.
	GroupKeys []string `json:"group_keys"`
	// Warnings that the server emitted for the query.
	Warnings []string `json:"warnings"`
	// Unit of the values. Empty if unknown.
	Unit string `json:"unit"`
	// CustomUnit is the unit that the query sets.
	CustomUnit string `json:"custom_unit"`
	// IgnoreUnit is true if Unit does not apply to the values.
	IgnoreUnit bool `json:"ignore_unit"`
}

// Series is a time series in an MPL query [Result].
type Series struct {
	// Metric is the name of the metric.
	Metric string
	// Tags identify the series. Numeric tag values are float64.
	Tags map[string]any
	// Start is the time of the first value.
	Start time.Time
	// Resolution is the time between two values.
	Resolution time.Duration
	// Data holds the values. The value at index i is at Start + i*Resolution.
	// A nil value is a gap or a value that is not finite.
	Data []*float64
	// Summary is the last value of the series, or nil.
	Summary *float64
}
