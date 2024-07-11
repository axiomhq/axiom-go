package axiom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersService_Current(t *testing.T) {
	exp := &User{
		ID:    "e9cffaad-60e7-4b04-8d27-185e1808c38c",
		Name:  "Lukas Malkmus",
		Email: "lukas@axiom.co",
		Role: struct {
			ID   string `json:"id,omitempty"`
			Name string `json:"name,omitempty"`
		}{
			ID:   "80f1b217-c142-404c-82e7-f4e48f8f8b78",
			Name: "super-user",
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
			"id": "e9cffaad-60e7-4b04-8d27-185e1808c38c",
			"name": "Lukas Malkmus",
			"email": "lukas@axiom.co",
			"role": {
				"id": "80f1b217-c142-404c-82e7-f4e48f8f8b78",
				"name": "super-user"
			}
		}`)
		assert.NoError(t, err)
	}

	client := setup(t, "/v2/user", hf)

	res, err := client.Users.Current(context.Background())
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestUserRole_Marshal(t *testing.T) {
	exp := `{
		"role": "read-only"
	}`

	b, err := json.Marshal(struct {
		Role UserRole `json:"role"`
	}{
		Role: RoleReadOnly,
	})
	require.NoError(t, err)
	require.NotEmpty(t, b)

	assert.JSONEq(t, exp, string(b))
}

func TestUserRole_Unmarshal(t *testing.T) {
	var act struct {
		Role UserRole `json:"role"`
	}
	err := json.Unmarshal([]byte(`{ "role": "read-only" }`), &act)
	require.NoError(t, err)

	assert.Equal(t, RoleReadOnly, act.Role)
}

func TestUserRole_String(t *testing.T) {
	// Check outer bounds.
	assert.Equal(t, RoleCustom, UserRole(0))
	assert.Contains(t, (RoleOwner + 1).String(), "UserRole(")

	for u := RoleCustom; u <= RoleOwner; u++ {
		s := u.String()
		assert.NotEmpty(t, s)
		assert.NotContains(t, s, "UserRole(")
	}
}

func TestUserRoleFromString(t *testing.T) {
	for r := RoleCustom; r <= RoleOwner; r++ {
		parsed := userRoleFromString(r.String())
		assert.Equal(t, r, parsed)
	}
}

func TestUserRole_Custom(t *testing.T) {
	r := userRoleFromString("badboys")
	assert.Equal(t, RoleCustom, r)
}
