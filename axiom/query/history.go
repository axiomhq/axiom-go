package query

import "time"

// History represents a query stored inside the query history.
type History struct {
	// ID is the unique ID of the history query.
	ID string `json:"id"`
	// Kind of the history query.
	Kind Kind `json:"kind"`
	// Dataset the history query belongs to.
	Dataset string `json:"dataset"`
	// Owner is the team or user ID of the history queries owner.
	Owner string `json:"who"`
	// Query is the actual query.
	Query Query `json:"query"`
	// CreatedAt is the time the history query was created.
	CreatedAt time.Time `json:"created"`
}
