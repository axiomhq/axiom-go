package axiom

import (
	"context"
	"net/http"
)

// UsersService handles communication with the user related operations of the
// Axiom API.
//
// Axiom API Reference: /api/v1/users
type UsersService service

// User represents an user of the deployment.
type User struct {
	// ID is the unique id of the user.
	ID string `json:"id"`
	// Name of the user.
	Name string `json:"name"`
	// Email is the primary email of the user.
	Email string `json:"email"`
	// Role of the user. Can be "owner", "admin", "user" or "read-only".
	Role string `json:"role"`
	// Permissions of the user.
	Permissions []string `json:"permissions"`
}

// AuthenticatedUser represents an authenticated Axiom user.
type AuthenticatedUser struct {
	// ID is the unique id of the user.
	ID string `json:"id"`
	// Name of the user.
	Name string `json:"name"`
	// Emails are the email addresses of the user.
	Emails []string `json:"emails"`
}

// CreateUserRequest is a request used to create an user.
type CreateUserRequest struct {
	// Name of the user.
	Name string `json:"name"`
	// Email is the primary email address of the user.
	Email string `json:"email"`
	// Role of the user. Must be one of "owner", "admin", "user" or "read-only".
	Role string `json:"role"`
	// TeamIDs are the unique IDs of the teams the user will be part of.
	TeamIDs []string `json:"teamIds"`
}

// UpdateUserRequest is a request used to update an user.
type UpdateUserRequest struct {
	// Name of the user.
	Name string `json:"name"`
}

// Current retrieves the authenticated user.
func (s *UsersService) Current(ctx context.Context) (*AuthenticatedUser, error) {
	path := "/api/v1/user"

	var res AuthenticatedUser
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// List all available users.
func (s *UsersService) List(ctx context.Context, opts ListOptions) ([]*User, error) {
	path, err := addOptions(s.basePath, opts)
	if err != nil {
		return nil, err
	}

	var res []*User
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// Get a user by id.
func (s *UsersService) Get(ctx context.Context, id string) (*User, error) {
	path := s.basePath + "/" + id

	var res User
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Create a user with the given properties.
func (s *UsersService) Create(ctx context.Context, req CreateUserRequest) (*User, error) {
	var res User
	if err := s.client.call(ctx, http.MethodPost, s.basePath, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Update the user identified by the given id with the given properties.
func (s *UsersService) Update(ctx context.Context, id string, req UpdateUserRequest) (*User, error) {
	path := s.basePath + "/" + id

	var res User
	if err := s.client.call(ctx, http.MethodPut, path, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Delete the user identified by the given id.
func (s *UsersService) Delete(ctx context.Context, id string) error {
	path := s.basePath + "/" + id

	if err := s.client.call(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return err
	}

	return nil
}
