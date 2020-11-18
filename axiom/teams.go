package axiom

import (
	"context"
	"net/http"
)

// Team represents a team
type Team struct {
	// ID is the unique id of the team.
	ID string `json:"id"`
	// Name of the team.
	Name string `json:"name"`
	// Members are the IDs of the teams members.
	Members []string `json:"members"`
	// Datasets are the IDs of the teams assigned datasets.
	Datasets []string `json:"datasets"`
}

// TeamCreateRequest is a request used to create a team.
type TeamCreateRequest struct {
	// Name of the team.
	Name string `json:"name"`
	// Datasets of the team.
	Datasets []string `json:"datasets"`
}

// TeamsService handles communication with the team related operations of the
// Axiom API.
//
// Axiom API Reference: /api/v1/teams
type TeamsService service

// List all available teams.
func (s *TeamsService) List(ctx context.Context) ([]*Team, error) {
	var res []*Team
	if err := s.client.call(ctx, http.MethodGet, s.basePath, nil, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// Get a team by id.
func (s *TeamsService) Get(ctx context.Context, id string) (*Team, error) {
	path := s.basePath + "/" + id

	var res Team
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Create a team with the given properties.
func (s *TeamsService) Create(ctx context.Context, req TeamCreateRequest) (*Team, error) {
	var res Team
	if err := s.client.call(ctx, http.MethodPost, s.basePath, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Update the team identified by the given id with the given properties.
func (s *TeamsService) Update(ctx context.Context, id string, req Team) (*Team, error) {
	path := s.basePath + "/" + id

	var res Team
	if err := s.client.call(ctx, http.MethodPut, path, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Delete the team identified by the given id.
func (s *TeamsService) Delete(ctx context.Context, id string) error {
	path := s.basePath + "/" + id

	if err := s.client.call(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return err
	}

	return nil
}
