package axiom

import (
	"context"
	"net/http"
)

type version struct {
	// CurrentVersion of the deployment.
	CurrentVersion string `json:"currentVersion"`
}

// VersionService handles communication with the version related operations of
// the Axiom API.
//
// Axiom API Reference: /api/v1/version
type VersionService service

// Get the version of a deployment.
func (s *VersionService) Get(ctx context.Context) (string, error) {
	var res version
	if err := s.client.call(ctx, http.MethodGet, s.basePath, nil, &res); err != nil {
		return "", err
	}

	return res.CurrentVersion, nil
}
