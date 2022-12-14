package query

import "time"

// Options specifies the optional parameters for a query.
type Options struct {
	// StartTime of the query.
	StartTime time.Time `url:"-" json:"startTime,omitempty"`
	// EndTime of the query.
	EndTime time.Time `url:"-" json:"endTime,omitempty"`
	// Cursor is the cursor to use for pagination.
	Cursor string `url:"-" json:"cursor,omitempty"`
	// IncludeCursor specifies whether the cursor should be included in the
	// request.
	IncludeCursor bool `url:"-" json:"includeCursor,omitempty"`
}

// An Option applies an optional parameter to a query.
type Option func(*Options)

// SetStartTime specifies the queries start time.
func SetStartTime(startTime time.Time) Option {
	return func(o *Options) { o.StartTime = startTime }
}

// SetEndTime specifies the queries end time.
func SetEndTime(endTime time.Time) Option {
	return func(o *Options) { o.EndTime = endTime }
}

// SetCursor specifies the queries cursor.
func SetCursor(cursor string) Option {
	return func(o *Options) { o.Cursor = cursor; o.IncludeCursor = true }
}
