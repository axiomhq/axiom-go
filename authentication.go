package axiom

import (
	"context"
	"errors"
	"net/http"
)

// ErrUnauthenticated is raised when the access token used by the client isn't
// valid.
var ErrUnauthenticated = errors.New("invalid authentication credentials")

// AuthenticationService bundles all the Axiom API authentication operations.
type AuthenticationService interface {
	// Valid returns nil if the authentication is valid.
	Valid(context.Context) error
}

var _ AuthenticationService = (*authenticationService)(nil)

type authenticationService struct {
	client *Client
}

// Valid returns nil if the authentication is valid.
// TODO(lukasmalkmus): Don't abuse the /version endpoint.
func (s *authenticationService) Valid(ctx context.Context) error {
	path := "/api/v1/version"

	resp, err := s.client.call(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			return ErrUnauthenticated
		}
		return err
	}

	return nil
}
