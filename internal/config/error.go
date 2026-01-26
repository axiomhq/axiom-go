package config

import "errors"

// ErrMissingToken is raised when a token is not provided. Set it manually using
// the [SetToken] [Option] or export "AXIOM_TOKEN".
var ErrMissingToken = errors.New("missing token")

// ErrMissingOrganizationID is raised when an organization ID is not provided.
// Set it manually using the [SetOrganizationID] [Option] or export
// "AXIOM_ORG_ID".
var ErrMissingOrganizationID = errors.New("missing organization id")

// ErrInvalidToken is returned when the token is invalid.
var ErrInvalidToken = errors.New("invalid token")

// ErrPersonalTokenNotSupportedForEdge is returned when a personal token is
// used for edge operations. Edge endpoints only support API tokens.
var ErrPersonalTokenNotSupportedForEdge = errors.New("personal tokens are not supported for edge operations, use an API token (xaat-)")
