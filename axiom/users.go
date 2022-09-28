package axiom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=UserRole -linecomment -output=users_string.go

// UserRole represents the role of a user.
type UserRole uint8

// All available user roles.
const (
	emptyUserRole UserRole = iota //

	RoleReadOnly // read-only
	RoleUser     // user
	RoleAdmin    // admin
	RoleOwner    // owner
)

func userRoleFromString(s string) (ur UserRole, err error) {
	switch s {
	case emptyUserRole.String():
		ur = emptyUserRole
	case RoleReadOnly.String():
		ur = RoleReadOnly
	case RoleUser.String():
		ur = RoleUser
	case RoleAdmin.String():
		ur = RoleAdmin
	case RoleOwner.String():
		ur = RoleOwner
	default:
		err = fmt.Errorf("unknown user role %q", s)
	}

	return ur, err
}

// MarshalJSON implements `json.Marshaler`. It is in place to marshal the
// UserRole to its string representation because that's what the server expects.
func (ur UserRole) MarshalJSON() ([]byte, error) {
	return json.Marshal(ur.String())
}

// UnmarshalJSON implements `json.Unmarshaler`. It is in place to unmarshal the
// UserRole from the string representation the server returns.
func (ur *UserRole) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}

	*ur, err = userRoleFromString(s)

	return err
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
// Axiom API Reference: /api/v1/users
type UsersService service

// Current retrieves the authenticated user.
func (s *UsersService) Current(ctx context.Context) (*User, error) {
	path := "/api/v1/user"

	var res User
	if err := s.client.Call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
