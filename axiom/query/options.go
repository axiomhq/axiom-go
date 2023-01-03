package query

import "time"

// Options specifies the optional parameters for a query.
type Options struct {
	// StartTime for the interval to query. If sorting with newest first the
	// value is inclusive, otherwise exclusive.
	StartTime time.Time `json:"startTime,omitempty"`
	// EndTime of the interval to query. If sorting with newest first the value
	// is exclusive, otherwise inclusive.
	EndTime time.Time `json:"endTime,omitempty"`
	// Cursor to use for pagination.
	Cursor string `json:"cursor,omitempty"`
	// IncludeCursor specifies whether the event that matches the cursor should
	// be included in the result.
	IncludeCursor bool `json:"includeCursor,omitempty"`
}

// An Option applies an optional parameter to a query.
type Option func(*Options)

// SetStartTime specifies the query intervals start time. If sorting with newest
// first the value is inclusive, otherwise exclusive.
func SetStartTime(startTime time.Time) Option {
	return func(o *Options) { o.StartTime = startTime }
}

// SetEndTime specifies the query intervals end time. If sorting with newest
// first the value is exclusive, otherwise inclusive.
func SetEndTime(endTime time.Time) Option {
	return func(o *Options) { o.EndTime = endTime }
}

// SetCursor specifies the queries cursor. If include is set to true the event
// that matches the cursor will be included in the result. Be advised that the
// exlusive nature of the cursor depends on the sort order, see [SetStartTime]
// and [SetEndTime].
func SetCursor(cursor string, include bool) Option {
	return func(o *Options) { o.Cursor = cursor; o.IncludeCursor = include }
}
