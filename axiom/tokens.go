package axiom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

//go:generate go tool stringer -type=Action -linecomment -output=tokens_string.go

// Action represents an action that can be performed on an Axiom resource.
type Action uint8

// All available [Action]s.
const (
	emptyAction Action = iota //

	ActionCreate // create
	ActionRead   // read
	ActionUpdate // update
	ActionDelete // delete
)

func actionFromString(s string) (a Action, err error) {
	switch s {
	case emptyAction.String():
		a = emptyAction
	case ActionCreate.String():
		a = ActionCreate
	case ActionRead.String():
		a = ActionRead
	case ActionUpdate.String():
		a = ActionUpdate
	case ActionDelete.String():
		a = ActionDelete
	default:
		err = fmt.Errorf("unknown action %q", s)
	}

	return a, err
}

// MarshalJSON implements [json.Marshaler]. It is in place to marshal the
// Action to its string representation because that's what the server expects.
func (a Action) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

// UnmarshalJSON implements [json.Unmarshaler]. It is in place to unmarshal the
// Action from the string representation the server returns.
func (a *Action) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}

	*a, err = actionFromString(s)

	return err
}

// APIToken represents an API token returned from the Axiom API.
type APIToken struct {
	// ID is the unique ID of the token.
	ID string `json:"id"`
	// Name is the name of the token.
	Name string `json:"name"`
	// Description is the description of the token.
	Description string `json:"description"`
	// ExpiresAt is the time when the token expires.
	ExpiresAt time.Time `json:"expiresAt"`
	// DatasetCapabilities is a map of dataset names to the capabilities
	// available to that dataset for the token.
	DatasetCapabilities map[string]DatasetCapabilities `json:"datasetCapabilities"`
	// OrganisationCapabilities is the organisation capabilities available to
	// the token.
	OrganisationCapabilities OrganisationCapabilities `json:"orgCapabilities"`
	// SAMLAuthenticated is a flag that determines whether the token can access
	// a SAML authenticated org
	SAMLAuthenticated bool `json:"samlAuthenticated"`
}

// DatasetCapabilities represents the capabilities available to a token for a
// dataset.
type DatasetCapabilities struct {
	// Ingest is the ingest capability and the actions that can be performed on
	// them.
	Ingest []Action `json:"ingest"`
	// Query is the query capability and the actions that can be performed on
	// them.
	Query []Action `json:"query"`
	// StarredQueries is the starred queries capability and the actions that can
	// be performed on them.
	StarredQueries []Action `json:"starredQueries"`
	// VirtualFields is the VirtualFields capability and the actions that can be
	// performed on them.
	VirtualFields []Action `json:"virtualFields"`
	// Trim is the trim capability and the actions that can be performed on
	// them.
	Trim []Action `json:"trim"`
	// Vacuum is the vacuum capability and the actions that can be performed on
	// them.
	Vacuum []Action `json:"vacuum"`
	// Data is the data capability and the actions that can be performed on
	// them.
	Data []Action `json:"data"`
	// Share is the share capability and the actions that can be performed on
	// them.
	Share []Action `json:"share"`
}

// OrganisationCapabilities represents the capabilities available to a token for
// an organisation.
type OrganisationCapabilities struct {
	// Annotations is the Annotations capability and the actions that can be
	// performed on them.
	Annotations []Action `json:"annotations,omitempty"`
	// APITokens is the APITokens capability and the actions that can be
	// performed on them.
	APITokens []Action `json:"apiTokens,omitempty"`
	// AuditLog is the AuditLog capability and the actions that can be
	// performed on it.
	AuditLog []Action `json:"auditLog,omitempty"`
	// Billing is the Billing capability and the actions that can be performed
	// on them.
	Billing []Action `json:"billing,omitempty"`
	// Dashboards is the Dashboards capability and the actions that can be
	// performed on them.
	Dashboards []Action `json:"dashboards,omitempty"`
	// Datasets is the Datasets capability and the actions that can be performed
	// on them.
	Datasets []Action `json:"datasets,omitempty"`
	// Endpoints is the Endpoints capability and the actions that can be
	// performed on them.
	Endpoints []Action `json:"endpoints,omitempty"`
	// Flows is the Flows capability and the actions that can be performed on
	// them.
	Flows []Action `json:"flows,omitempty"`
	// Integrations is the Integrations capability and the actions that can be
	// performed on them.
	Integrations []Action `json:"integrations,omitempty"`
	// Monitors is the Monitors capability and the actions that can be performed
	// on them.
	Monitors []Action `json:"monitors,omitempty"`
	// Notifiers is the Notifiers capability and the actions that can be
	// performed on them.
	Notifiers []Action `json:"notifiers,omitempty"`
	// RBAC is the RBAC capability and the actions that can be performed on
	// them.
	RBAC []Action `json:"rbac,omitempty"`
	// SharedAccessKeys is the SharedAccessKeys capability and the actions that
	// can be performed on them.
	SharedAccessKeys []Action `json:"sharedAccessKeys,omitempty"`
	// Users is the Users capability and the actions that can be performed on
	// them.
	Users []Action `json:"users,omitempty"`
	// Views is the view capability and the actions that can be performed on
	// them.
	Views []Action `json:"views,omitempty"`
}

// CreateTokenRequest is the request payload for creating a new token with the
// Axiom API.
type CreateTokenRequest struct {
	// Name is the name of the token.
	Name string `json:"name"`
	// Description is the description of the token.
	Description string `json:"description"`
	// ExpiresAt is the time when the token expires.
	ExpiresAt time.Time `json:"expiresAt"`
	// DatasetCapabilities is a map of dataset names to the capabilities
	// available to that dataset for the token.
	DatasetCapabilities map[string]DatasetCapabilities `json:"datasetCapabilities"`
	// OrganisationCapabilities is the organisation capabilities available to
	// the token.
	OrganisationCapabilities OrganisationCapabilities `json:"orgCapabilities"`
}

// CreateTokenResponse is the response payload for creating a new token with the
// Axiom API.
type CreateTokenResponse struct {
	APIToken
	// Token is the token value to be used in api calls
	Token string `json:"token"`
}

// RegenerateTokenRequest is the request payload for regenerating a token with
// the Axiom API.
type RegenerateTokenRequest struct {
	// ExistingTokenExpiresAt is the time when the existing token will expire.
	ExistingTokenExpiresAt time.Time `json:"existingTokenExpiresAt"`
	// NewTokenExpiresAt is the time when the new token will expire.
	NewTokenExpiresAt time.Time `json:"newTokenExpiresAt"`
}

// TokensService handles communication with the api token related operations
// of the Axiom API.
//
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

	path, err := url.JoinPath(s.basePath, id, "regenerate")
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

	path, err := url.JoinPath(s.basePath, id)
	if err != nil {
		return spanError(span, err)
	}

	if err := s.client.Call(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return spanError(span, err)
	}

	return nil
}
