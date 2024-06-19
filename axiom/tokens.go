package axiom

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"net/url"
	"time"
)

type Actions []string

type APIToken struct {
	ID                  string                         `json:"id"`
	Name                string                         `json:"name"`
	Description         string                         `json:"description"`
	ExpiresAt           time.Time                      `json:"expiresAt"`
	DatasetCapabilities map[string]DatasetCapabilities `json:"datasetCapabilities"`
	OrgCapabilities     OrgCapabilities                `json:"orgCapabilities"`
}

type DatasetCapabilities struct {
	Ingest         []string `json:"ingest"`
	Query          []string `json:"query"`
	StarredQueries []string `json:"starredQueries"`
	VirtualFields  []string `json:"virtualFields"`
}

type OrgCapabilities struct {
	Annotations      []string `json:"annotations,omitempty"`
	APITokens        []string `json:"apiTokens,omitempty"`
	Billing          []string `json:"billing,omitempty"`
	Dashboards       []string `json:"dashboards,omitempty"`
	Datasets         []string `json:"datasets,omitempty"`
	Endpoints        []string `json:"endpoints,omitempty"`
	Flows            []string `json:"flows,omitempty"`
	Integrations     []string `json:"integrations,omitempty"`
	Monitors         []string `json:"monitors,omitempty"`
	Notifiers        []string `json:"notifiers,omitempty"`
	Rbac             []string `json:"rbac,omitempty"`
	SharedAccessKeys []string `json:"sharedAccessKeys,omitempty"`
	Users            []string `json:"users,omitempty"`
}

type CreateTokenRequest struct {
	Name                string                         `json:"name"`
	Description         string                         `json:"description"`
	ExpiresAt           time.Time                      `json:"expiresAt"`
	DatasetCapabilities map[string]DatasetCapabilities `json:"datasetCapabilities"`
	OrgCapabilities     OrgCapabilities                `json:"orgCapabilities"`
}

type CreateTokenResponse struct {
	APIToken
	Token string `json:"token"`
}

type RegenerateTokenRequest struct {
	ExistingTokenExpiresAt time.Time `json:"existingTokenExpiresAt"`
	NewTokenExpiresAt      time.Time `json:"newTokenExpiresAt"`
}

// Axiom API Reference: /v2/tokens
type TokensService service

// List all available tokens.
func (s *TokensService) List(ctx context.Context) ([]*APIToken, error) {
	ctx, span := s.client.trace(ctx, "Tokens.List")
	defer span.End()

	var res []*APIToken
	if err := s.client.Call(ctx, http.MethodGet, s.basePath, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return res, nil
}

// Get a token by id.
func (s *TokensService) Get(ctx context.Context, id string) (*APIToken, error) {
	ctx, span := s.client.trace(ctx, "Tokens.Get", trace.WithAttributes(
		attribute.String("axiom.token_id", id),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, id)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res APIToken
	if err := s.client.Call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// Create a token with the given properties.
func (s *TokensService) Create(ctx context.Context, req CreateTokenRequest) (*CreateTokenResponse, error) {
	ctx, span := s.client.trace(ctx, "Tokens.Create", trace.WithAttributes(
		attribute.String("axiom.param.name", req.Name),
	))
	defer span.End()

	var res CreateTokenResponse
	if err := s.client.Call(ctx, http.MethodPost, s.basePath, req, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// Regenerate the token identified by the given id.
func (s *TokensService) Regenerate(ctx context.Context, id string, req RegenerateTokenRequest) (*CreateTokenResponse, error) {
	ctx, span := s.client.trace(ctx, "Tokens.Regenerate", trace.WithAttributes(
		attribute.String("axiom.token_id", id),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, id, "/regenerate")
	if err != nil {
		return nil, spanError(span, err)
	}

	var res CreateTokenResponse
	if err := s.client.Call(ctx, http.MethodPost, path, req, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// Delete the token identified by the given id.
func (s *TokensService) Delete(ctx context.Context, id string) error {
	ctx, span := s.client.trace(ctx, "Tokens.Delete", trace.WithAttributes(
		attribute.String("axiom.token_id", id),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, "/", id)
	if err != nil {
		return spanError(span, err)
	}

	if err := s.client.Call(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return spanError(span, err)
	}

	return nil
}

func (t *CreateTokenResponse) AsAPIToken() *APIToken {
	return &t.APIToken
}
