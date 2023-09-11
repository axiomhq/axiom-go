package axiom

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// ErrUnauthorized is raised when the user or authentication token misses
// permissions to perform the requested operation.
var ErrUnauthorized = newHTTPError(http.StatusForbidden)

// ErrUnauthenticated is raised when the authentication on the request is not
// valid.
var ErrUnauthenticated = newHTTPError(http.StatusUnauthorized)

// ErrNotFound is returned when the requested resource is not found.
var ErrNotFound = newHTTPError(http.StatusNotFound)

// ErrExists is returned when the resource that was attempted to create already
// exists.
var ErrExists = newHTTPError(http.StatusConflict)

// ErrUnprivilegedToken is raised when a [Client] tries to call an ingest or
// query endpoint with an API token configured.
var ErrUnprivilegedToken = errors.New("using API token for non-ingest or non-query operation")

// HTTPError is the generic error response returned on non 2xx HTTP status
// codes.
type HTTPError struct {
	Status  int    `json:"-"`
	Message string `json:"message"`
	TraceID string `json:"-"`
}

func newHTTPError(code int) HTTPError {
	return HTTPError{
		Status:  code,
		Message: http.StatusText(code),
	}
}

// Error implements error.
func (e HTTPError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.Status, e.Message)
}

// Is returns whether the provided error equals this error.
func (e HTTPError) Is(target error) bool {
	v, ok := target.(HTTPError)
	if !ok {
		return false
	}
	return e.Status == v.Status
}

// LimitError occurs when http status codes 429 (TooManyRequests) or 430
// (Axiom-sepcific when ingest or query limit are reached) are encountered.
type LimitError struct {
	HTTPError

	Limit Limit
}

// Error returns the string representation of the limit error.
//
// It implements error.
func (e LimitError) Error() string {
	return fmt.Sprintf("%s limit exceeded: try again in %s",
		e.Limit.limitType, time.Until(e.Limit.Reset).Truncate(time.Second))
}

// Is returns whether the provided error equals this error.
func (e LimitError) Is(target error) bool {
	v, ok := target.(LimitError)
	if !ok {
		return false
	}
	return e.Limit == v.Limit && e.HTTPError.Is(v.HTTPError)
}
