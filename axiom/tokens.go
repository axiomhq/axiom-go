package axiom

import (
	"context"
	"net/http"
)

// Token represents an access token. Tokens can either be ingest tokens, valid
// for ingestion into one or more datasets or personal tokens for access to the
// whole Axiom API.
type Token struct {
	// ID is the unique id of the token.
	ID string `json:"id"`
	// Name of the token.
	Name string `json:"name"`
	// Description of the token.
	Description string `json:"description"`
	// Scopes of the token. Only used by ingest tokens.
	Scopes []string `json:"scopes"`
}

// RawToken represents a raw token secret and its attached scopes.
type RawToken struct {
	// Token is the actual secret value of the token.
	Token string `json:"token"`
	// Scopes of the token. Only used by ingest tokens.
	Scopes []string `json:"scopes"`
}

// TokenCreateRequest is a request used to create a token.
type TokenCreateRequest struct {
	// Name of the token.
	Name string `json:"name"`
	// Description of the token.
	Description string `json:"description"`
	// Scopes of the token. Only used by ingest tokens. Can be set to a list of
	// dataset IDs to allow ingestion into the specified ones or "*" for all,
	// which is the default when the field is unset.
	Scopes []string `json:"scopes"`
}

// TokensService handles communication with the token related operations of the
// Axiom API.
//
// Axiom API Reference: /api/v1/tokens/{ingest,personal}
type TokensService service

// List all available tokens.
func (s *TokensService) List(ctx context.Context) ([]*Token, error) {
	var res []*Token
	if err := s.client.call(ctx, http.MethodGet, s.basePath, nil, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// Get a token by id.
func (s *TokensService) Get(ctx context.Context, id string) (*Token, error) {
	path := s.basePath + "/" + id

	var res Token
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// View a raw token secret by id.
func (s *TokensService) View(ctx context.Context, id string) (*RawToken, error) {
	path := s.basePath + "/" + id + "/token"

	var res RawToken
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Create a token with the given properties.
func (s *TokensService) Create(ctx context.Context, req TokenCreateRequest) (*Token, error) {
	var res Token
	if err := s.client.call(ctx, http.MethodPost, s.basePath, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Update the token identified by the given id with the given properties.
func (s *TokensService) Update(ctx context.Context, id string, req Token) (*Token, error) {
	path := s.basePath + "/" + id

	var res Token
	if err := s.client.call(ctx, http.MethodPut, path, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Delete the token identified by the given id.
func (s *TokensService) Delete(ctx context.Context, id string) error {
	path := s.basePath + "/" + id

	if err := s.client.call(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return err
	}

	return nil
}
