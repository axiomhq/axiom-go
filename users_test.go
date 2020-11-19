package axiom

import (
	"context"
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
		require.NoError(t, err)
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
			Role:        "owner",
			Permissions: []string{},
		},
		{
			ID:    "e9cffaad-60e7-4b04-8d27-185e1808c38c",
			Name:  "Michael Doe",
			Email: "michael@example.com",
			Role:  "owner",
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

		limit := r.URL.Query().Get("limit")
		assert.Empty(t, limit)

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
		require.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/users", hf)
	defer teardown()

	res, err := client.Users.List(context.Background(), ListOptions{})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestUsersService_List_OptionsLimit(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		limit := r.URL.Query().Get("limit")
		assert.Equal(t, limit, "1")

		_, err := fmt.Fprint(w, `[
			{
				"id": "20475220-20e4-4080-b2f4-68315e21f5ec",
				"name": "John Doe",
				"email": "john@example.com",
				"role": "owner",
				"permissions": []
			}
		]`)
		require.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/users", hf)
	defer teardown()

	_, err := client.Users.List(context.Background(), ListOptions{
		Limit: 1,
	})
	require.NoError(t, err)
}

func TestUsersService_Get(t *testing.T) {
	exp := &User{
		ID:          "20475220-20e4-4080-b2f4-68315e21f5ec",
		Name:        "John Doe",
		Email:       "john@example.com",
		Role:        "owner",
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
		require.NoError(t, err)
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
		Role:  "user",
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		_, err := fmt.Fprint(w, `{
			"id": "7debe8bb-69f1-436f-94f6-a2fe23e71cf5",
			"name": "John Doe",
			"email": "john@example.com",
			"role": "user",
			"permissions": null
		}`)
		require.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/users", hf)
	defer teardown()

	res, err := client.Users.Create(context.Background(), CreateUserRequest{
		Name:  "John Doe",
		Email: "john@example.com",
		Role:  "user",
	})
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
