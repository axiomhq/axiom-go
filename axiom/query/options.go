package query

import "time"

// Options specifies the optional parameters for a query.
type Options struct {
	// StartTime of the query.
	StartTime time.Time `url:"-"`
	// EndTime of the query.
	EndTime time.Time `url:"-"`
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
