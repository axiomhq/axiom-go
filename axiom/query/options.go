package query

import "time"

// Options specifies the parameters for the Query method of the Datasets
// service.
type Options struct {
	// StreamingDuration of a query.
	StreamingDuration time.Duration `url:"streaming-duration,omitempty"`
	// NoCache omits the query cache.
	NoCache bool `url:"no-cache,omitempty"`
	// SaveKind saves the query on the server with the given query kind. The ID
	// of the saved query is returned with the query result inside the
	SaveKind Kind `url:"saveAsKind,omitempty"`
}
