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
	exp := &AuthenticatedUser{
		ID:   "e9cffaad-60e7-4b04-8d27-185e1808c38c",
		Name: "Lukas Malkmus",
		Emails: []string{
			"lukas@axiom.co",
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

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
func TestUsersService_List(t *testing.T) {
	exp := []*User{
		{
			ID:          "20475220-20e4-4080-b2f4-68315e21f5ec",
			Name:        "John Doe",
			Email:       "john@example.com",
			Role:        RoleOwner,
			Permissions: []string{},
		},
		{
			ID:    "e9cffaad-60e7-4b04-8d27-185e1808c38c",
			Name:  "Michael Doe",
			Email: "michael@example.com",
			Role:  RoleOwner,
			Permissions: []string{
				"CanUpdate",
				"ChangeAccess",
				"ChangeApiKeys",
				"ChangeAuthentication",
				"ChangeDashboards",
				"ChangeIntegrations",
				"ChangeMonitorsAndNotifiers",
				"ChangeSavedQueries",
				"ChangeVirtualFields",
				"ManageBilling",
				"ManageDatasets",
				"ManageIngestTokens",
			},
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `[
			{
				"id": "20475220-20e4-4080-b2f4-68315e21f5ec",
				"name": "John Doe",
				"email": "john@example.com",
				"role": "owner",
				"permissions": []
			},
			{
				"id": "e9cffaad-60e7-4b04-8d27-185e1808c38c",
				"name": "Michael Doe",
				"email": "michael@example.com",
				"role": "owner",
				"permissions": [
					"CanUpdate",
					"ChangeAccess",
					"ChangeApiKeys",
					"ChangeAuthentication",
					"ChangeDashboards",
					"ChangeIntegrations",
					"ChangeMonitorsAndNotifiers",
					"ChangeSavedQueries",
					"ChangeVirtualFields",
					"ManageBilling",
					"ManageDatasets",
					"ManageIngestTokens"
				]
			}
		]`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/users", hf)
	defer teardown()

	res, err := client.Users.List(context.Background())
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestUsersService_Get(t *testing.T) {
	exp := &User{
		ID:          "20475220-20e4-4080-b2f4-68315e21f5ec",
		Name:        "John Doe",
		Email:       "john@example.com",
		Role:        RoleOwner,
		Permissions: []string{},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `{
			"id": "20475220-20e4-4080-b2f4-68315e21f5ec",
			"name": "John Doe",
			"email": "john@example.com",
			"role": "owner",
			"permissions": []
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/users/20475220-20e4-4080-b2f4-68315e21f5ec", hf)
	defer teardown()

	res, err := client.Users.Get(context.Background(), "20475220-20e4-4080-b2f4-68315e21f5ec")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestUsersService_Create(t *testing.T) {
	exp := &User{
		ID:    "7debe8bb-69f1-436f-94f6-a2fe23e71cf5",
		Name:  "John Doe",
		Email: "john@example.com",
		Role:  RoleUser,
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		_, err := fmt.Fprint(w, `{
			"id": "7debe8bb-69f1-436f-94f6-a2fe23e71cf5",
			"name": "John Doe",
			"email": "john@example.com",
			"role": "user",
			"permissions": null
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/users", hf)
	defer teardown()

	res, err := client.Users.Create(context.Background(), UserCreateRequest{
		Name:  "John Doe",
		Email: "john@example.com",
		Role:  RoleUser,
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestUsersService_Update(t *testing.T) {
	exp := &User{
		ID:    "7debe8bb-69f1-436f-94f6-a2fe23e71cf5",
		Name:  "Michael Doe",
		Email: "john@example.com",
		Role:  RoleUser,
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		_, err := fmt.Fprint(w, `{
			"id": "7debe8bb-69f1-436f-94f6-a2fe23e71cf5",
			"name": "Michael Doe",
			"email": "john@example.com",
			"role": "user",
			"permissions": null
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/users/7debe8bb-69f1-436f-94f6-a2fe23e71cf5", hf)
	defer teardown()

	res, err := client.Users.Update(context.Background(), "7debe8bb-69f1-436f-94f6-a2fe23e71cf5", UserUpdateRequest{
		Name: "Michael Doe",
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestUsersService_UpdateRole(t *testing.T) {
	exp := &User{
		ID:    "7debe8bb-69f1-436f-94f6-a2fe23e71cf5",
		Name:  "Michael Doe",
		Email: "john@example.com",
		Role:  RoleAdmin,
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		_, err := fmt.Fprint(w, `{
			"id": "7debe8bb-69f1-436f-94f6-a2fe23e71cf5",
			"name": "Michael Doe",
			"email": "john@example.com",
			"role": "admin",
			"permissions": null
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/users/7debe8bb-69f1-436f-94f6-a2fe23e71cf5/role", hf)
	defer teardown()

	res, err := client.Users.UpdateRole(context.Background(), "7debe8bb-69f1-436f-94f6-a2fe23e71cf5", RoleAdmin)
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestUsersService_Delete(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}

	client, teardown := setup(t, "/api/v1/users/7debe8bb-69f1-436f-94f6-a2fe23e71cf5", hf)
	defer teardown()

	err := client.Users.Delete(context.Background(), "7debe8bb-69f1-436f-94f6-a2fe23e71cf5")
	require.NoError(t, err)
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
	assert.Contains(t, (RoleReadOnly - 1).String(), "UserRole(")
	assert.Contains(t, (RoleOwner + 1).String(), "UserRole(")

	for r := RoleReadOnly; r <= RoleOwner; r++ {
		s := r.String()
		assert.NotEmpty(t, s)
		assert.NotContains(t, s, "UserRole(")
	}
}
