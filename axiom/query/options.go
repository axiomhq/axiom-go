package query

import "time"

// Options specifies the optional parameters for a query.
type Options struct {
	// StartTime for the interval to query.
	StartTime time.Time `json:"startTime,omitempty"`
	// EndTime of the interval to query.
	EndTime time.Time `json:"endTime,omitempty"`
	// Cursor to use for pagination.
	Cursor string `json:"cursor,omitempty"`
	// IncludeCursor specifies whether the event that matches the cursor should
	// be included in the result.
	IncludeCursor bool `json:"includeCursor,omitempty"`
	// Variables is an optional set of additional variables that are inserted into the APL
	Variables map[string]any `json:"variables,omitempty"`
}

// An Option applies an optional parameter to a query.
type Option func(*Options)

// SetStartTime specifies the start time of the query interval.
func SetStartTime(startTime time.Time) Option {
	return func(o *Options) { o.StartTime = startTime }
}

// SetEndTime specifies the end time of the query interval.
func SetEndTime(endTime time.Time) Option {
	return func(o *Options) { o.EndTime = endTime }
}

// SetCursor specifies the cursor of the query. If include is set to true the
// event that matches the cursor will be included in the result.
func SetCursor(cursor string, include bool) Option {
	return func(o *Options) { o.Cursor = cursor; o.IncludeCursor = include }
}

// SetVariables specifies variables which can be referenced by the APL query.
func SetVariables(variables map[string]any) Option {
	return func(o *Options) { o.Variables = variables }
}

// SetVariable sets a single variable which can be referenced by the APL
// query.
func SetVariable(key string, value any) Option {
	return func(o *Options) {
		if o.Variables == nil {
			o.Variables = make(map[string]any, 1)
		}
		o.Variables[key] = value
	}
}
