package query

import (
	"encoding/json"
	"time"
)

// Query represents a query that gets executed on a dataset.
type Query struct {
	// StartTime of the query. Required.
	StartTime time.Time `json:"startTime"`
	// EndTime of the query. Required.
	EndTime time.Time `json:"endTime"`
	// Resolution of the queries graph. Valid values are the queries time
	// range / 100 at maximum and / 1000 at minimum. Use zero value for
	// serve-side auto-detection.
	Resolution time.Duration `json:"resolution"`
	// Aggregations performed as part of the query.
	Aggregations []Aggregation `json:"aggregations"`
	// Filter applied on the queried results.
	Filter Filter `json:"filter"`
	// GroupBy specifies a list of field names to group the query result by.
	GroupBy []string `json:"groupBy"`
	// Order specifies a list of order rules that specify the order of the query
	// result.
	Order []Order `json:"order"`
	// Limit the amount of results returned from the query.
	Limit uint32 `json:"limit"`
	// VirtualFields specifies a list of virtual fields that can be referenced
	// by aggregations, filters and orders.
	VirtualFields []VirtualField `json:"virtualFields"`
	// Cursor is the query cursor. It should be set to the Cursor returned with
	// a previous query result if it was partial.
	Cursor string `json:"cursor"`
}

// MarshalJSON implements json.Marshaler. It is in place to marshal the
// Resolutions zero value to its proper string representation because that's
// what the server expects.
func (q Query) MarshalJSON() ([]byte, error) {
	type LocalQuery Query
	localQuery := struct {
		LocalQuery

		Resolution string `json:"resolution"`
	}{
		LocalQuery: LocalQuery(q),

		Resolution: q.Resolution.String(),
	}

	// If the resolution is not specified, it is set to auto for resolution
	// auto-detection on the server side.
	if q.Resolution == 0 {
		localQuery.Resolution = "auto"
	}

	return json.Marshal(localQuery)
}

// UnmarshalJSON implements json.Unmarshaler. It is in place to unmarshal the
// Resolutionstring value to a proper time.Duration because that's what the
// server returns.
func (q *Query) UnmarshalJSON(b []byte) error {
	type LocalQuery Query
	localQuery := struct {
		*LocalQuery

		Resolution string `json:"resolution"`
	}{
		LocalQuery: (*LocalQuery)(q),
	}

	if err := json.Unmarshal(b, &localQuery); err != nil {
		return err
	}

	// If the resolution is not specified, parsing it is omitted.
	var err error
	if s := localQuery.Resolution; s != "" && s != "auto" {
		q.Resolution, err = time.ParseDuration(s)
	}

	return err
}

// Order specifies the order a queries result will be in.
type Order struct {
	// Field to order on.
	Field string `json:"field"`
	// Desc specifies if the field is ordered ascending or descending.
	Desc bool `json:"desc"`
}

// A VirtualField is not part of a dataset and its value is derived from an
// expression. Aggregations, filters and orders can reference this field like
// any other field.
type VirtualField struct {
	// Alias the virtual field is referenced by.
	Alias string `json:"alias"`
	// Expression which specifies the virtual fields value.
	Expression string `json:"expr"`
}
