package config

import "errors"

var (
	// ErrMissingAccessToken is raised when an access token is not provided. Set
	// it manually or export `AXIOM_TOKEN`.
	ErrMissingAccessToken = errors.New("missing access token")

	// ErrMissingOrganizationID is raised when an organization ID is not
	// provided. Set it manually using the SetOrgID option or export
	// `AXIOM_ORG_ID`.
	ErrMissingOrganizationID = errors.New("missing organization id")

	// ErrInvalidToken is returned when the access token is invalid.
	ErrInvalidToken = errors.New("invalid access token")
)
