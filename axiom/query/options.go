package query

import "time"

// Options specifies the optional parameters to various query methods.
type Options struct {
	// StreamingDuration of a query.
	StreamingDuration time.Duration `url:"streaming-duration,omitempty"`
	// NoCache omits the query cache.
	NoCache bool `url:"nocache,omitempty"`
	// SaveKind saves the query on the server with the given query kind. The ID
	// of the saved query is returned with the query result as part of the
	// response.
	SaveKind Kind `url:"saveAsKind,omitempty"`
}
