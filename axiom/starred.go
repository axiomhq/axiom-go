package axiom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/axiomhq/axiom-go/axiom/query"
)

//go:generate ../bin/stringer -type=OwnerKind,QueryKind -linecomment -output=starred_string.go

// OwnerKind represents the kind of a starred queries owner.
type OwnerKind uint8

// All available query kinds.
const (
	OwnedByUser OwnerKind = iota
	OwnedByTeam           // team
)

// EncodeValues implements query.Encoder. It is in place to encode the OwnerKind
// into a string URL value because that's what the server expects.
func (ok OwnerKind) EncodeValues(key string, v *url.Values) error {
	v.Set(key, ok.String())
	return nil
}

// QueryKind represents the role of a query.
type QueryKind uint8

// All available query kinds.
const (
	Analytics QueryKind = iota + 1 // analytics
	Stream                         // stream
)

// MarshalJSON implements json.Marshaler. It is in place to marshal the
// QueryKind to its string representation because that's what the server
// expects.
func (qk QueryKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(qk.String())
}

// UnmarshalJSON implements json.Unmarshaler. It is in place to unmarshal the
// QueryKind from the string representation the server returns.
func (qk *QueryKind) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch s {
	case Analytics.String():
		*qk = Analytics
	case Stream.String():
		*qk = Stream
	default:
		return fmt.Errorf("unknown query kind %q", s)
	}

	return nil
}

// EncodeValues implements query.Encoder. It is in place to encode the QueryKind
// into a string URL value because that's what the server expects.
func (qk QueryKind) EncodeValues(key string, v *url.Values) error {
	v.Set(key, qk.String())
	return nil
}

// StarredQuery represents a starred query of a dataset.
type StarredQuery struct {
	// ID is the unique id of the starred query.
	ID string `json:"id"`
	// Kind of the starred query.
	Kind QueryKind `json:"kind"`
	// Dataset the starred query belongs to.
	Dataset string `json:"dataset"`
	// Owner is the ID of the starred queries owner. Can be a user or team ID.
	Owner string `json:"who"`
	// Name is the display name of the starred query.
	Name string `json:"name"`
	// Query is the actual query.
	Query query.Query `json:"query"`
	// Metadata associated with the query.
	Metadata map[string]string `json:"metadata"`
	// Created is the time the starred query was created at.
	Created time.Time `json:"created"`
}

// StarredQueriesListOptions specifies the parameters for the List method of the
// StarredQuerys service.
type StarredQueriesListOptions struct {
	// Kind of queries to list. Required.
	Kind QueryKind `url:"kind"`
	// Dataset to list starred queries for.
	Dataset string `url:"dataset,omitempty"`
	// Owner will list the starred queries of a users teams instead of personal
	// ones, if set to true.
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
	path := s.basePath + "/" + id

	if err := s.client.call(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return err
	}

	return nil
}
