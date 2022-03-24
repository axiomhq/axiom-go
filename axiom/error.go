package axiom

import (
	"errors"
	"fmt"
	"net/http"
	"time"
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

	// ErrUnauthorized is raised when the user or token misses permissions to
	// perform the requested operation.
	ErrUnauthorized = errors.New("insufficient permissions")

	// ErrUnauthenticated is raised when the authentication on the request is
	// not valid.
	ErrUnauthenticated = errors.New("invalid authentication credentials")

	// ErrUnprivilegedToken is raised when a client tries to call an ingest or
	// query endpoint with an API token configured.
	ErrUnprivilegedToken = errors.New("using API token for non-ingest or non-query operation")

	// ErrNotFound is returned when the requested resource is not found.
	ErrNotFound = errors.New("not found")

	// ErrExists is returned when the requested resource already exists.
	ErrExists = errors.New("entity exists")
)

// Error is the generic error response returned on non 2xx HTTP status codes.
// Either one of the two fields is populated. However, calling the Error()
// method is preferred.
type Error struct {
	Status  int    `json:"-"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("API error %d: %s", e.Status, e.Message)
}

// Is returns whether the provided error equals this error.
func (e *Error) Is(target error) bool {
	v, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Status == v.Status && e.Message == v.Message
}

// LimitError occurs when http status code 429 (TooManyRequests) is encountered.
type LimitError struct {
	Limit   Limit  `json:"-"`
	Message string `json:"message"`

	response *http.Response
}

// Error returns the string representation of the error.
//
// It implements the `error` interface.
func (e *LimitError) Error() string {
	return fmt.Sprintf("%s: try again in %s",
		e.Message, time.Until(e.Limit.Reset).Truncate(time.Second))
}

// Is returns whether the provided error equals this error.
func (e *LimitError) Is(target error) bool {
	v, ok := target.(*LimitError)
	if !ok {
		return false
	}
	return e.Limit == v.Limit && e.Message == v.Message
}
