package query

import "time"

// Options specifies the parameters for the Query method of the Datasets
// service.
type Options struct {
	// StreamingDuration of a query.
	StreamingDuration time.Duration `url:"streaming-duration,omitempty"`
	// NoCache omits the query cache.
	NoCache bool `url:"no-cache,omitempty"`
}
