package axiom

import (
	"context"
	"net/http"
)

// Token represents an access token. Tokens can either be api tokens, valid
// across the whole Axiom API and granting access to the specified resources,
// ingest tokens, valid for ingestion into one or more datasets or personal
// tokens, granting access to the whole Axiom API only limited by the users
// role.
type Token struct {
	// ID is the unique ID of the token.
	ID string `json:"id"`
	// Name of the token.
	Name string `json:"name"`
	// Description of the token.
	Description string `json:"description"`
	// Scopes of the token. Only used by api and ingest tokens.
	Scopes []string `json:"scopes"`
	// Permissions of the token. Only used by api and ingest tokens.
	Permissions []string `json:"permissions"`
}

// RawToken represents a raw token secret and its attached scopes.
type RawToken struct {
	// Token is the actual secret value of the token.
	Token string `json:"token"`
	// Scopes of the token. Only used by api and ingest tokens.
	Scopes []string `json:"scopes"`
	// Permissions of the token. Only used by api and ingest tokens.
	Permissions []string `json:"permissions"`
}

// TokenCreateUpdateRequest is a request used to create a token.
type TokenCreateUpdateRequest struct {
	// Name of the token.
	Name string `json:"name"`
	// Description of the token.
	Description string `json:"description"`
	// Scopes of the token. Only used by api and ingest tokens.
	Scopes []string `json:"scopes"`
	// Permissions of the token. Only used by api and ingest tokens.
	Permissions []string `json:"permissions"`
}

// tokensService implements the methods sharred between the ingest and personal
// token services.
type tokensService service

// List all available tokens.
func (s *tokensService) List(ctx context.Context) ([]*Token, error) {
	var res []*Token
	if err := s.client.call(ctx, http.MethodGet, s.basePath, nil, &res); err != nil {
		return nil, err
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

	return &res, nil
}

// View a raw token secret by id.
func (s *tokensService) View(ctx context.Context, id string) (*RawToken, error) {
	path := s.basePath + "/" + id + "/token"

	var res RawToken
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Create a token with the given properties.
func (s *tokensService) Create(ctx context.Context, req TokenCreateUpdateRequest) (*Token, error) {
	var res Token
	if err := s.client.call(ctx, http.MethodPost, s.basePath, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Update the token identified by the given id with the given properties.
func (s *tokensService) Update(ctx context.Context, id string, req TokenCreateUpdateRequest) (*Token, error) {
	path := s.basePath + "/" + id

	var res Token
	if err := s.client.call(ctx, http.MethodPut, path, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Delete the token identified by the given id.
func (s *tokensService) Delete(ctx context.Context, id string) error {
	return s.client.call(ctx, http.MethodDelete, s.basePath+"/"+id, nil, nil)
}

// APITokensService handles communication with the api token related operations
// of the Axiom API.
//
// Axiom API Reference: /api/v1/tokens/api
type APITokensService struct {
	tokensService
}

// IngestTokensService handles communication with the ingest token related
// operations of the Axiom API.
//
// Axiom API Reference: /api/v1/tokens/ingest
type IngestTokensService struct {
	tokensService
}

// Validate the token that is used for authentication.
func (s *IngestTokensService) Validate(ctx context.Context) error {
	return s.client.call(ctx, http.MethodGet, s.basePath+"/validate", nil, nil)
}

// PersonalTokensService handles communication with the personal token related
// operations of the Axiom API.
//
// Axiom API Reference: /api/v1/tokens/personal
type PersonalTokensService struct {
	tokensService
}
