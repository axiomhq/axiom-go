package config

import "errors"

var (
	// ErrMissingToken is raised when a token is not provided. Set it manually
	// using the `axiom.SetToken` option when constructing a client or export
	// `AXIOM_TOKEN`.
	ErrMissingToken = errors.New("missing token")

	// ErrMissingOrganizationID is raised when an organization ID is not
	// provided. Set it manually using the `axiom.SetOrganizationID` option when
	// constructing a client or export `AXIOM_ORG_ID`.
	ErrMissingOrganizationID = errors.New("missing organization id")

	// ErrInvalidToken is returned when the token is invalid.
	ErrInvalidToken = errors.New("invalid token")
)
