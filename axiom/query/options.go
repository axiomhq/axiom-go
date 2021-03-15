package query

import "time"

// Options specifies the parameters for the Query method of the Datasets
// service.
type Options struct {
	// ContinuationToken used to get more results of a previous query.
	ContinuationToken string `url:"continuation-token,omitempty"`
	// StreamingDuration of a query.
	StreamingDuration time.Duration `url:"streaming-duration,omitempty"`
	// NoCache omits the query cache.
	NoCache bool `url:"no-cache,omitempty"`
}
