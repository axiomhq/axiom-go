package query

import "encoding/json"

// Query represents an Axiom Processing Language (APL) query.
type Query string

type aplQueryRequest struct {
	Query string `json:"apl"`
}

// Query implements `axiom.Query`.
func (q Query) Query() {}

// MarshalJSON implements `json.Marshaler`. It is in place to marshal the
// query string wrapped in a JSON object.
func (q Query) MarshalJSON() ([]byte, error) {
	return json.Marshal(aplQueryRequest{
		Query: string(q),
	})
}

// UnmarshalJSON implements `json.Unmarshaler`. It is in place to unmarshal the
// query string value that is wrapped in a JSON object.
func (q *Query) UnmarshalJSON(b []byte) error {
	var req aplQueryRequest
	err := json.Unmarshal(b, &req)
	*q = Query(req.Query)
	return err
}
