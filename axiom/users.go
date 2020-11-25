package axiom

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

//go:generate ../bin/stringer -type=UserRole -linecomment -output=users_string.go

// UserRole represents the role of a user.
type UserRole uint8

// All available user roles.
const (
	RoleReadOnly UserRole = iota + 1 // read-only
	RoleUser                         // user
	RoleAdmin                        // admin
	RoleOwner                        // owner
)

// MarshalJSON implements json.Marshaler. It is in place to marshal the
// UserRole to its string representation because that's what the server expects.
func (ur UserRole) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("%q", ur)
	return []byte(s), nil
}

// UnmarshalJSON implements json.Unmarshaler. It is in place to unmarshal the
// UserRole from the string representation the server returns.
func (ur *UserRole) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)

	switch s {
	case RoleReadOnly.String():
		*ur = RoleReadOnly
	case RoleUser.String():
		*ur = RoleUser
	case RoleAdmin.String():
		*ur = RoleAdmin
	case RoleOwner.String():
		*ur = RoleOwner
	default:
		return fmt.Errorf("unknown user role %q", s)
	}

	return nil
}

// User represents an user of the deployment.
type User struct {
	// ID is the unique id of the user.
	ID string `json:"id"`
	// Name of the user.
	Name string `json:"name"`
	// Email is the primary email of the user.
	Email string `json:"email"`
	// Role of the user.
	Role UserRole `json:"role"`
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
	// Role of the user.
	Role UserRole `json:"role"`
	// TeamIDs are the unique IDs of the teams the user will be part of.
	TeamIDs []string `json:"teamIds"`
}

// UpdateUserRequest is a request used to update an user.
type UpdateUserRequest struct {
	// Name of the user.
	Name string `json:"name"`
}

type updateUserRoleRequest struct {
	// Role is the new role of the user.
	Role UserRole `json:"role"`
}

// UsersService handles communication with the user related operations of the
// Axiom API.
//
// Axiom API Reference: /api/v1/users
type UsersService service

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
func (s *UsersService) List(ctx context.Context) ([]*User, error) {
	var res []*User
	if err := s.client.call(ctx, http.MethodGet, s.basePath, nil, &res); err != nil {
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

// UpdateRole updates the role of the user identified by the given id with the
// given properties.
func (s *UsersService) UpdateRole(ctx context.Context, id string, role UserRole) (*User, error) {
	path := s.basePath + "/" + id + "/role"

	var res User
	if err := s.client.call(ctx, http.MethodPut, path, updateUserRoleRequest{role}, &res); err != nil {
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
