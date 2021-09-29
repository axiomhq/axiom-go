package axiom

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidToken is returned when the access token is invalid.
	ErrInvalidToken = errors.New("invalid access token")

	// ErrMissingAccessToken is raised when an access token is not provided. Set
	// it manually using the SetAccessToken option or export `AXIOM_TOKEN`.
	ErrMissingAccessToken = errors.New("missing access token")

	// ErrMissingOrganizationID is raised when an organization ID is not
	// provided. Set it manually using the SetOrgID option or export
	// `AXIOM_ORG_ID`.
	ErrMissingOrganizationID = errors.New("missing organization id")

	// ErrUnauthenticated is raised when the access token is not valid.
	ErrUnauthenticated = errors.New("invalid authentication credentials")

	// ErrUnprivilegedToken is raised when a client tries to call a non-ingest
	// endpoint with an ingest-only token configured.
	ErrUnprivilegedToken = errors.New("using ingest token for non-ingest operation")

	// ErrNotFound is returned when the requested resource is not found.
	ErrNotFound = errors.New("not found")

	// ErrExists is returned when the requested resource already exists.
	ErrExists = errors.New("entity exists")
)

// Error is the generic error response returned on non 2xx HTTP status codes.
// Either one of the two fields is populated. However, calling the Error()
// method is preferred.
type Error struct {
	Status       int    `json:"-"`
	ErrorMessage string `json:"error"`
	Message      string `json:"message"`
}

// Error implements the error interface.
func (e Error) Error() string {
	if e.ErrorMessage != "" {
		return fmt.Sprintf("API error %d: %s", e.Status, e.ErrorMessage)
	}
	return fmt.Sprintf("API error %d: %s", e.Status, e.Message)
}
