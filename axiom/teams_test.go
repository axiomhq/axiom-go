package axiom

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamsService_List(t *testing.T) {
	exp := []*Team{
		{
			ID:   "CcXzGSwIFeshgnHTmD",
			Name: "Test",
			Members: []string{
				"7debe8bb-69f1-436f-94f6-a2fe23e71cf5",
			},
			Datasets: []string{
				"logs",
			},
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `[
			{
				"id": "CcXzGSwIFeshgnHTmD",
				"name": "Test",
				"members": [
					"7debe8bb-69f1-436f-94f6-a2fe23e71cf5"
				],
				"datasets": [
					"logs"
				]
			}
		]`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/teams", hf)
	defer teardown()

	res, err := client.Teams.List(context.Background())
	require.NoError(t, err)

	assert.EqualValues(t, exp, res)
}

func TestTeamsService_Get(t *testing.T) {
	exp := &Team{
		ID:   "CcXzGSwIFeshgnHTmD",
		Name: "Test",
		Members: []string{
			"7debe8bb-69f1-436f-94f6-a2fe23e71cf5",
		},
		Datasets: []string{
			"logs",
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `{
			"id": "CcXzGSwIFeshgnHTmD",
			"name": "Test",
			"members": [
				"7debe8bb-69f1-436f-94f6-a2fe23e71cf5"
			],
			"datasets": [
				"logs"
			]
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/teams/CcXzGSwIFeshgnHTmD", hf)
	defer teardown()

	res, err := client.Teams.Get(context.Background(), "CcXzGSwIFeshgnHTmD")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestTeamsService_Create(t *testing.T) {
	exp := &Team{
		ID:   "4miTfZKp29VByAQgTd",
		Name: "Server Team",
		Datasets: []string{
			"test",
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		_, err := fmt.Fprint(w, `{
			"id": "4miTfZKp29VByAQgTd",
			"name": "Server Team",
			"members": null,
			"datasets": [
				"test"
			]
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/teams", hf)
	defer teardown()

	res, err := client.Teams.Create(context.Background(), TeamCreateRequest{
		Name: "Server Team",
		Datasets: []string{
			"test",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestTeamsService_Update(t *testing.T) {
	exp := &Team{
		ID:   "4miTfZKp29VByAQgTd",
		Name: "Server Team",
		Members: []string{
			"7debe8bb-69f1-436f-94f6-a2fe23e71cf5",
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		_, err := fmt.Fprint(w, `{
			"id": "4miTfZKp29VByAQgTd",
			"name": "Server Team",
			"members": [
				"7debe8bb-69f1-436f-94f6-a2fe23e71cf5"
			],
			"datasets": null
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/teams/4miTfZKp29VByAQgTd", hf)
	defer teardown()

	res, err := client.Teams.Update(context.Background(), "4miTfZKp29VByAQgTd", Team{
		Members: []string{
			"7debe8bb-69f1-436f-94f6-a2fe23e71cf5",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestTeamsService_Delete(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}

	client, teardown := setup(t, "/api/v1/teams/4miTfZKp29VByAQgTd", hf)
	defer teardown()

	err := client.Teams.Delete(context.Background(), "4miTfZKp29VByAQgTd")
	require.NoError(t, err)
}
