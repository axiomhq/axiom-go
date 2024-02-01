package axiom

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=UserRole -linecomment -output=users_string.go

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type UpdateUserRequest struct {
	Name string `json:"name"`
}

type UpdateUserRoleRequest struct {
	Role string `json:"role"`
}

// UserRole represents the role of an [User].
type UserRole uint8

// All available [User] roles.
const (
	RoleCustom   UserRole = iota // custom
	RoleNone                     // none
	RoleReadOnly                 // read-only
	RoleUser                     // user
	RoleAdmin                    // admin
	RoleOwner                    // owner
)

func userRoleFromString(s string) (ur UserRole) {
	switch s {
	case RoleNone.String():
		ur = RoleNone
	case RoleReadOnly.String():
		ur = RoleReadOnly
	case RoleUser.String():
		ur = RoleUser
	case RoleAdmin.String():
		ur = RoleAdmin
	case RoleOwner.String():
		ur = RoleOwner
	default:
		ur = RoleCustom
	}

	return ur
}

// MarshalJSON implements [json.Marshaler]. It is in place to marshal the
// UserRole to its string representation because that's what the server expects.
func (ur UserRole) MarshalJSON() ([]byte, error) {
	return json.Marshal(ur.String())
}

// UnmarshalJSON implements [json.Unmarshaler]. It is in place to unmarshal the
// UserRole from the string representation the server returns.
func (ur *UserRole) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}

	*ur = userRoleFromString(s)

	return
}

// User represents an user.
type User struct {
	// ID is the unique ID of the user.
	ID string `json:"id"`
	// Name of the user.
	Name string `json:"name"`
	// Emails are the email addresses of the user.
	Emails []string `json:"emails"`
}

// UsersService handles communication with the user related operations of the
// Axiom API.
//
// Axiom API Reference: /v1/users
type UsersService service

// Current retrieves the authenticated user.
func (s *UsersService) Current(ctx context.Context) (*User, error) {
	ctx, span := s.client.trace(ctx, "Users.Current")
	defer span.End()

	var res User
	if err := s.client.Call(ctx, http.MethodGet, "/v2/user", nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// List all users.
func (s *UsersService) List(ctx context.Context) ([]*User, error) {
	ctx, span := s.client.trace(ctx, "Users.List")
	defer span.End()

	var res []*User
	if err := s.client.Call(ctx, http.MethodGet, s.basePath, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return res, nil
}

// Get a user by id.
func (s *UsersService) Get(ctx context.Context, id string) (*User, error) {
	ctx, span := s.client.trace(ctx, "Users.Get")
	defer span.End()

	path, err := url.JoinPath(s.basePath, "/", id)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res User
	if err := s.client.Call(ctx, http.MethodPost, path, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// Create will create and invite a user to the organisation
func (s *UsersService) Create(ctx context.Context, req CreateUserRequest) (*User, error) {
	ctx, span := s.client.trace(ctx, "Users.Create")
	defer span.End()

	var res User
	if err := s.client.Call(ctx, http.MethodPost, s.basePath, req, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// Update will update a user.
func (s *UsersService) Update(ctx context.Context, id string, req UpdateUserRequest) (*User, error) {
	ctx, span := s.client.trace(ctx, "Users.Update")
	defer span.End()

	path, err := url.JoinPath(s.basePath, "/", id)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res User
	if err := s.client.Call(ctx, http.MethodPost, path, req, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// UpdateUsersRole will update a user role.
func (s *UsersService) UpdateUsersRole(ctx context.Context, id string, req UpdateUserRoleRequest) (*User, error) {
	ctx, span := s.client.trace(ctx, "Users.UpdateUsersRole")
	defer span.End()

	path, err := url.JoinPath(s.basePath, "/", id)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res User
	if err := s.client.Call(ctx, http.MethodPost, path, req, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// Delete will remove a user from the organization.
func (s *UsersService) Delete(ctx context.Context, id string) error {
	ctx, span := s.client.trace(ctx, "Users.Delete")
	defer span.End()

	path, err := url.JoinPath(s.basePath, "/", id)
	if err != nil {
		return spanError(span, err)
	}

	if err := s.client.Call(ctx, http.MethodDelete, path, nil, &nil); err != nil {
		return spanError(span, err)
	}

	return nil
}
