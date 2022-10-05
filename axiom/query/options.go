package query

import (
	"time"
)

// Options specifies the optional parameters to APL query methods.
type Options struct {
	// StartTime of the query.
	StartTime time.Time `url:"-"`
	// EndTime of the query.
	EndTime time.Time `url:"-"`

	// NoCache omits the query cache.
	NoCache bool `url:"nocache,omitempty"`
	// Save the query on the server, if set to `true`. The ID of the saved query
	// is returned with the query result as part of the response.
	// HINT(lukasmalkmus): The server automatically sets the query kind to "apl"
	// for queries going to the "/_apl" query endpoint. This allows us to set
	// any value for the `saveAsKind` query param. For user experience, we use a
	// boolean here instead of forcing the user to set a concrete value.
	Save bool `url:"saveAsKind,omitempty"`
	// Format specifies the format of the APL query. Defaults to Legacy.
	Format Format `url:"format"`
}
