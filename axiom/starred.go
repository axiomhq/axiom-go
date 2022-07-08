package axiom

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/axiomhq/axiom-go/axiom/apl"
	"github.com/axiomhq/axiom-go/axiom/query"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=OwnerKind -linecomment -output=starred_string.go

// OwnerKind represents the kind of a starred queries owner.
type OwnerKind uint8

// All available query kinds.
const (
	OwnedByUser OwnerKind = iota // user
	OwnedByTeam                  // team
)

// EncodeValues implements `query.Encoder`. It is in place to encode the
// OwnerKind into a string URL value because that's what the server expects.
func (ok OwnerKind) EncodeValues(key string, v *url.Values) error {
	v.Set(key, ok.String())
	return nil
}

// StarredQuery represents a starred query of a dataset.
type StarredQuery struct {
	// ID is the unique ID of the starred query.
	ID string `json:"id"`
	// Kind of the starred query. For create or update operations the field is
	// set to `APL` by the client if the query is an APL query.
	Kind query.Kind `json:"kind"`
	// Dataset the starred query belongs to.
	Dataset string `json:"dataset"`
	// Owner is the team or user ID of the starred queries owner.
	Owner string `json:"who"`
	// Name is the display name of the starred query.
	Name string `json:"name"`
	// Query is the actual query.
	Query Query `json:"query"`
	// Metadata associated with the query.
	Metadata map[string]string `json:"metadata"`
	// CreatedAt is the time the starred query was created.
	CreatedAt time.Time `json:"created"`
}

// MarshalJSON implements `json.Marshaler`. It is in place to set the
// appropriate query kind.
func (s StarredQuery) MarshalJSON() ([]byte, error) {
	type localStarredQuery StarredQuery

	// Make sure the `Kind` field matches if the query is an APL query.
	if _, ok := s.Query.(apl.Query); ok {
		s.Kind = query.APL
	}

	return json.Marshal(localStarredQuery(s))
}

// UnmarshalJSON implements `json.Unmarshaler`. It is in place to unmarshal the
// query in to its appropriate type.
func (s *StarredQuery) UnmarshalJSON(b []byte) error {
	type LocalStarredQuery StarredQuery
	localStarredQuery := struct {
		*LocalStarredQuery

		Query json.RawMessage `json:"query"`
	}{
		LocalStarredQuery: (*LocalStarredQuery)(s),
	}

	if err := json.Unmarshal(b, &localStarredQuery); err != nil {
		return err
	}

	// Figure out if the query is an APL query or not and unmarshal into the
	// appropriate type, should there be data to unmarshal.
	if b = localStarredQuery.Query; len(b) > 0 {
		var err error
		if s.Kind == query.APL {
			var q apl.Query
			err = json.Unmarshal(b, &q)
			s.Query = q
		} else {
			var q query.Query
			err = json.Unmarshal(b, &q)
			s.Query = q
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// StarredQueriesListOptions specifies the parameters for the List method of the
// StarredQuerys service.
type StarredQueriesListOptions struct {
	// Kind of queries to list. Required.
	Kind query.Kind `url:"kind"`
	// Dataset to list starred queries for.
	Dataset string `url:"dataset,omitempty"`
	// Owner specifies if the starred queries of a users teams or personal ones
	// are listed.
	Owner OwnerKind `url:"who,omitempty"`

	ListOptions
}

// StarredQueriesService handles communication with the starred query related
// operations of the Axiom API.
//
// Axiom API Reference: /api/v1/starred
type StarredQueriesService service

// List all available starred queries.
func (s *StarredQueriesService) List(ctx context.Context, opts StarredQueriesListOptions) ([]*StarredQuery, error) {
	path, err := addOptions(s.basePath, opts)
	if err != nil {
		return nil, err
	}

	var res []*StarredQuery
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// Get a starred query by id.
func (s *StarredQueriesService) Get(ctx context.Context, id string) (*StarredQuery, error) {
	path := s.basePath + "/" + id

	var res StarredQuery
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Create a starred query with the given properties.
func (s *StarredQueriesService) Create(ctx context.Context, req StarredQuery) (*StarredQuery, error) {
	var res StarredQuery
	if err := s.client.call(ctx, http.MethodPost, s.basePath, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Update the starred query identified by the given id with the given properties.
func (s *StarredQueriesService) Update(ctx context.Context, id string, req StarredQuery) (*StarredQuery, error) {
	path := s.basePath + "/" + id

	var res StarredQuery
	if err := s.client.call(ctx, http.MethodPut, path, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Delete the starred query identified by the given id.
func (s *StarredQueriesService) Delete(ctx context.Context, id string) error {
	return s.client.call(ctx, http.MethodDelete, s.basePath+"/"+id, nil, nil)
}
