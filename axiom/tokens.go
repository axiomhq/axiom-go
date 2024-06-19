package axiom

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"net/url"
	"time"
)

type APIToken struct {
	// ID is the unique ID of the token.
	ID string `json:"id"`
	// Name is the name of the token.
	Name string `json:"name"`
	// Description is the description of the token.
	Description string `json:"description"`
	// ExpiresAt is the time when the token expires.
	ExpiresAt time.Time `json:"expiresAt"`
	// DatasetCapabilities is a map of dataset names to the capabilities available to that dataset for the token.
	DatasetCapabilities map[string]DatasetCapabilities `json:"datasetCapabilities"`
	// OrgCapabilities is the organisation capabilities available to the token.
	OrgCapabilities OrgCapabilities `json:"orgCapabilities"`
}

type DatasetCapabilities struct {
	// Ingest is the ingest capability and the actions that can be performed on them.
	Ingest []string `json:"ingest"`
	// Query is the query capability and the actions that can be performed on them.
	Query []string `json:"query"`
	// StarredQueries is the starred queries capability and the actions that can be performed on them.
	StarredQueries []string `json:"starredQueries"`
	// VirtualFields is the VirtualFields capability and the actions that can be performed on them.
	VirtualFields []string `json:"virtualFields"`
}

type OrgCapabilities struct {
	// Annotations is the Annotations capability and the actions that can be performed on them.
	Annotations []string `json:"annotations,omitempty"`
	// APITokens is the APITokens capability and the actions that can be performed on them.
	APITokens []string `json:"apiTokens,omitempty"`
	// Billing is the Billing capability and the actions that can be performed on them.
	Billing []string `json:"billing,omitempty"`
	// Dashboards is the Dashboards capability and the actions that can be performed on them.
	Dashboards []string `json:"dashboards,omitempty"`
	// Datasets is the Datasets capability and the actions that can be performed on them.
	Datasets []string `json:"datasets,omitempty"`
	// Endpoints is the Endpoints capability and the actions that can be performed on them.
	Endpoints []string `json:"endpoints,omitempty"`
	// Flows is the Flows capability and the actions that can be performed on them.
	Flows []string `json:"flows,omitempty"`
	// Integrations is the Integrations capability and the actions that can be performed on them.
	Integrations []string `json:"integrations,omitempty"`
	// Monitors is the Monitors capability and the actions that can be performed on them.
	Monitors []string `json:"monitors,omitempty"`
	// Notifiers is the Notifiers capability and the actions that can be performed on them.
	Notifiers []string `json:"notifiers,omitempty"`
	// Rbac is the Rbac capability and the actions that can be performed on them.
	Rbac []string `json:"rbac,omitempty"`
	// SharedAccessKeys is the SharedAccessKeys capability and the actions that can be performed on them.
	SharedAccessKeys []string `json:"sharedAccessKeys,omitempty"`
	// Users is the Users capability and the actions that can be performed on them.
	Users []string `json:"users,omitempty"`
}

type CreateTokenRequest struct {
	// Name is the name of the token.
	Name string `json:"name"`
	// Description is the description of the token.
	Description string `json:"description"`
	// ExpiresAt is the time when the token expires.
	ExpiresAt time.Time `json:"expiresAt"`

	// DatasetCapabilities is a map of dataset names to the capabilities available to that dataset for the token.
	DatasetCapabilities map[string]DatasetCapabilities `json:"datasetCapabilities"`
	// OrgCapabilities is the organisation capabilities available to the token.
	OrgCapabilities OrgCapabilities `json:"orgCapabilities"`
}

type CreateTokenResponse struct {
	APIToken
	// Token is the token value to be used in api calls
	Token string `json:"token"`
}

type RegenerateTokenRequest struct {
	// ExistingTokenExpiresAt is the time when the existing token will expire.
	ExistingTokenExpiresAt time.Time `json:"existingTokenExpiresAt"`
	// NewTokenExpiresAt is the time when the new token will expire.
	NewTokenExpiresAt time.Time `json:"newTokenExpiresAt"`
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
