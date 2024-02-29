package query

import "time"

// Options specifies the optional parameters for a query.
type Options struct {
	// StartTime for the interval to query.
	StartTime string `json:"startTime,omitempty"`
	// EndTime of the interval to query.
	EndTime string `json:"endTime,omitempty"`
	// Cursor to use for pagination. When used, don't specify new start and end
	// times but rather use the start and end times of the query that returned
	// the cursor that will be used.
	Cursor string `json:"cursor,omitempty"`
	// IncludeCursor specifies whether the event that matches the cursor should
	// be included in the result.
	IncludeCursor bool `json:"includeCursor,omitempty"`
	// Variables is an optional set of additional variables can be referenced by
	// the APL query. Defining variables in APL using the "let" keyword takes
	// precedence over variables provided via the query options.
	Variables map[string]any `json:"variables,omitempty"`
}

// An Option applies an optional parameter to a query.
type Option func(*Options)

// SetStartTime specifies the start time of the query interval. When also using
// [SetCursor], please make sure to use the start time of the query that
// returned the cursor that will be used.
func SetStartTime[T time.Time | string](startTime T) Option {
	return func(o *Options) { o.StartTime = timeOrStringToString(startTime) }
}

// SetEndTime specifies the end time of the query interval. When also using
// [SetCursor], please make sure to use the end time of the query that returned
// the cursor that will be used.
func SetEndTime[T time.Time | string](endTime T) Option {
	return func(o *Options) { o.EndTime = timeOrStringToString(endTime) }
}

// SetCursor specifies the cursor of the query. If include is set to true the
// event that matches the cursor will be included in the result. When using this
// option, please make sure to use the initial query's start and end times.
func SetCursor(cursor string, include bool) Option {
	return func(o *Options) { o.Cursor = cursor; o.IncludeCursor = include }
}

// SetVariable adds a variable that can be referenced by the APL query. This
// option can be called multiple times to add multiple variables. If a variable
// with the same name already exists, it will be overwritten. Defining variables
// in APL using the "let" keyword takes precedence over variables provided via
// the query options.
func SetVariable(name string, value any) Option {
	return func(o *Options) {
		if o.Variables == nil {
			o.Variables = make(map[string]any, 1)
		}
		o.Variables[name] = value
	}
}

// SetVariables sets the variables that can be referenced by the APL query. It
// will overwrite any existing variables. Defining variables in APL using the
// "let" keyword takes precedence over variables provided via the query options.
func SetVariables(variables map[string]any) Option {
	return func(o *Options) { o.Variables = variables }
}

func timeOrStringToString[T time.Time | string](t T) string {
	switch t := any(t).(type) {
	case time.Time:
		return t.Format(time.RFC3339Nano)
	case string:
		return t
	}
	panic("time is neither time.Time nor string")
}
