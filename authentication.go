package axiom

import (
	"context"
	"net/http"
)

// AuthenticationService bundles all the Axiom API authentication operations.
type AuthenticationService interface {
	// Valid returns true if the access token is valid.
	Valid(context.Context) (bool, error)
}

var _ AuthenticationService = (*authenticationService)(nil)

type authenticationService struct {
	client *Client
}

// Valid returns nil if the authentication is valid.
// TODO(lukasmalkmus): Don't abuse the /version endpoint.
func (s *authenticationService) Valid(ctx context.Context) (bool, error) {
	path := "/api/v1/version"

	resp, err := s.client.call(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return false, err
	}

	return resp.StatusCode != 401 && resp.StatusCode != 403, nil
}
