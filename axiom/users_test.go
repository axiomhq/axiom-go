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
		ID:   "e9cffaad-60e7-4b04-8d27-185e1808c38c",
		Name: "Lukas Malkmus",
		Emails: []string{
			"lukas@axiom.co",
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
			"id": "e9cffaad-60e7-4b04-8d27-185e1808c38c",
			"name": "Lukas Malkmus",
			"emails": [
				"lukas@axiom.co"
			]
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/user", hf)
	defer teardown()

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
	assert.Empty(t, UserRole(0).String())
	assert.Empty(t, emptyUserRole.String())
	assert.Equal(t, emptyUserRole, UserRole(0))
	assert.Contains(t, (RoleOwner + 1).String(), "UserRole(")

	for typ := RoleReadOnly; typ <= RoleOwner; typ++ {
		s := typ.String()
		assert.NotEmpty(t, s)
		assert.NotContains(t, s, "UserRole(")
	}
}

func TestUserRoleFromString(t *testing.T) {
	for typ := RoleReadOnly; typ <= RoleOwner; typ++ {
		s := typ.String()

		parsedUserRole, err := userRoleFromString(s)
		assert.NoError(t, err)

		assert.NotEmpty(t, s)
		assert.Equal(t, typ, parsedUserRole)
	}
}
