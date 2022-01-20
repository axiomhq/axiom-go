package axiom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

//go:generate go run -mod=mod golang.org/x/tools/cmd/stringer -type=Permission -linecomment -output=tokens_string.go

// Permission of an API token.
type Permission uint8

// All available permissions.
const (
	emptyPermission Permission = iota //

	// CanIngest is the permission to write to a dataset.
	CanIngest // CanIngest
	// CanQuery is the permission to read from a dataset.
	CanQuery // CanQuery
)

func permissionFromString(s string) (permission Permission, err error) {
	switch s {
	// case emptyPermission.String():
	// 	permission = emptyPermission
	case CanIngest.String():
		permission = CanIngest
	case CanQuery.String():
		permission = CanQuery
	default:
		err = fmt.Errorf("unknown permission %q", s)
	}

	return permission, err
}

// MarshalJSON implements json.Marshaler. It is in place to marshal the
// Permission to its string representation because that's what the server
// expects.
func (p Permission) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// UnmarshalJSON implements json.Unmarshaler. It is in place to unmarshal the
// Permission from the string representation the server returns.
func (p *Permission) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}

	*p, err = permissionFromString(s)

	return err
}

// Token represents an access token. Tokens can either be API tokens, valid
// across the whole Axiom API and granting access to the specified resources or
// personal tokens, granting access to the whole Axiom API only limited by the
// users role.
type Token struct {
	// ID is the unique ID of the token.
	ID string `json:"id"`
	// Name of the token.
	Name string `json:"name"`
	// Description of the token.
	Description string `json:"description"`
	// Scopes of the token. Only used by API tokens. Usually the name of the
	// dataset to grant access to. `*` is the wildcard that grants access to all
	// datasets.
	Scopes []string `json:"scopes"`
	// Permissions of the token. Only used by API tokens.
	Permissions []Permission `json:"permissions"`
}

// RawToken represents a raw token secret and its attached scopes and
// permissions.
type RawToken struct {
	// Token is the actual secret value of the token.
	Token string `json:"token"`
	// Scopes of the token. Only used by API tokens. Usually the name
	// of the dataset to grant access to. `*` is the wildcard that grants access
	// to all datasets.
	Scopes []string `json:"scopes"`
	// Permissions of the token. Only used by API tokens.
	Permissions []Permission `json:"permissions"`
}

// TokenCreateUpdateRequest is a request used to create a token.
type TokenCreateUpdateRequest struct {
	// Name of the token.
	Name string `json:"name"`
	// Description of the token.
	Description string `json:"description"`
	// Scopes of the token. Only used by API tokens. Usually the name of the
	// dataset to grant access to. If left empty, will default to `*` which
	// grants access to all datasets.
	Scopes []string `json:"scopes,omitempty"`
	// Permissions of the token. Only used by API tokens.
	Permissions []Permission `json:"permissions,omitempty"`
}

// tokensService implements the methods sharred between the api, ingest and
// personal token services.
type tokensService service

// List all available tokens.
func (s *tokensService) List(ctx context.Context) ([]*Token, error) {
	var res []*Token
	if err := s.client.call(ctx, http.MethodGet, s.basePath, nil, &res); err != nil {
		return nil, err
	}

	for _, t := range res {
		cleanupTokenResponse(s.basePath, t)
	}

	return res, nil
}

// Get a token by id.
func (s *tokensService) Get(ctx context.Context, id string) (*Token, error) {
	path := s.basePath + "/" + id

	var res Token
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	cleanupTokenResponse(s.basePath, &res)

	return &res, nil
}

// View a raw token secret by id.
func (s *tokensService) View(ctx context.Context, id string) (*RawToken, error) {
	path := s.basePath + "/" + id + "/token"

	var res RawToken
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	cleanupRawTokenResponse(s.basePath, &res)

	return &res, nil
}

// Create a token with the given properties.
func (s *tokensService) Create(ctx context.Context, req TokenCreateUpdateRequest) (*Token, error) {
	prepareTokenCreateUpdateRequest(s.basePath, &req)

	var res Token
	if err := s.client.call(ctx, http.MethodPost, s.basePath, req, &res); err != nil {
		return nil, err
	}

	cleanupTokenResponse(s.basePath, &res)

	return &res, nil
}

// Update the token identified by the given id with the given properties.
func (s *tokensService) Update(ctx context.Context, id string, req TokenCreateUpdateRequest) (*Token, error) {
	prepareTokenCreateUpdateRequest(s.basePath, &req)

	path := s.basePath + "/" + id

	var res Token
	if err := s.client.call(ctx, http.MethodPut, path, req, &res); err != nil {
		return nil, err
	}

	cleanupTokenResponse(s.basePath, &res)

	return &res, nil
}

// Delete the token identified by the given id.
func (s *tokensService) Delete(ctx context.Context, id string) error {
	return s.client.call(ctx, http.MethodDelete, s.basePath+"/"+id, nil, nil)
}

const personalTokenStr = "personal"

func prepareTokenCreateUpdateRequest(basePath string, req *TokenCreateUpdateRequest) {
	// Nor scopes nor permissions are allowed on personal tokens.
	pathParts := strings.Split(basePath, "/")
	tokenType := pathParts[len(pathParts)-1]
	if tokenType == personalTokenStr {
		req.Scopes = nil
		req.Permissions = nil
	}
}

func cleanupTokenResponse(basePath string, t *Token) {
	// Nor scopes nor permissions are allowed on personal tokens.
	pathParts := strings.Split(basePath, "/")
	tokenType := pathParts[len(pathParts)-1]
	if tokenType == personalTokenStr {
		t.Scopes = nil
		t.Permissions = nil
	}
}

func cleanupRawTokenResponse(basePath string, t *RawToken) {
	// Nor scopes nor permissions are allowed on personal tokens.
	pathParts := strings.Split(basePath, "/")
	tokenType := pathParts[len(pathParts)-1]
	if tokenType == personalTokenStr {
		t.Scopes = nil
		t.Permissions = nil
	}
}

// APITokensService handles communication with the API token related operations
// of the Axiom API. API tokens allow access to datasets for either querying
// data, ingesting data, or both. They cannot be used to access other endpoints
// and can be configured to have access to all or specific datasets.
//
// Axiom API Reference: /api/v1/tokens/api
type APITokensService struct {
	tokensService
}

// PersonalTokensService handles communication with the personal token related
// operations of the Axiom API. Personal tokens act on behalf of the user and
// interactions are restricted by the users role and permissions.
//
// Axiom API Reference: /api/v1/tokens/personal
type PersonalTokensService struct {
	tokensService
}
