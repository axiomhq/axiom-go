package axiom

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HINT(lukasmalkmus): The tests below just test against the "personal"
// endpoint. However, the "ingest" implementation is the same. Under the hood,
// they both use the TokenService. The integration tests make sure this
// implementation works against both endpoints.

func TestTokensService_List(t *testing.T) {
	exp := []*Token{
		{
			ID:   "08fceb797a467c3c23151f3584c31cfaea962e3ca306e3af69c2dab28e8c2e6e",
			Name: "Test",
			Scopes: []string{
				"*",
			},
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `[
			{
				"id": "08fceb797a467c3c23151f3584c31cfaea962e3ca306e3af69c2dab28e8c2e6e",
				"name": "Test",
				"scopes": [
            		"*"
        		]
			}
		]`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/tokens/personal", hf)
	defer teardown()

	res, err := client.Tokens.Personal.List(context.Background())
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestTokensService_Get(t *testing.T) {
	exp := &Token{
		ID:   "08fceb797a467c3c23151f3584c31cfaea962e3ca306e3af69c2dab28e8c2e6e",
		Name: "Test",
		Scopes: []string{
			"*",
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `{
			"id": "08fceb797a467c3c23151f3584c31cfaea962e3ca306e3af69c2dab28e8c2e6e",
			"name": "Test",
			"scopes": [
				"*"
			]
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/tokens/personal/08fceb797a467c3c23151f3584c31cfaea962e3ca306e3af69c2dab28e8c2e6e", hf)
	defer teardown()

	res, err := client.Tokens.Personal.Get(context.Background(), "08fceb797a467c3c23151f3584c31cfaea962e3ca306e3af69c2dab28e8c2e6e")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestTokensService_View(t *testing.T) {
	exp := &RawToken{
		Token: "ae51e8d9-5fa2-4957-9847-3c1ccfa5ffe9",
		Scopes: []string{
			"*",
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `{
			"token": "ae51e8d9-5fa2-4957-9847-3c1ccfa5ffe9",
			"scopes": [
				"*"
			]
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/tokens/personal/08fceb797a467c3c23151f3584c31cfaea962e3ca306e3af69c2dab28e8c2e6e/token", hf)
	defer teardown()

	res, err := client.Tokens.Personal.View(context.Background(), "08fceb797a467c3c23151f3584c31cfaea962e3ca306e3af69c2dab28e8c2e6e")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestTokensService_Create(t *testing.T) {
	exp := &Token{
		ID:          "08fceb797a467c3c23151f3584c31cfaea962e3ca306e3af69c2dab28e8c2e6e",
		Name:        "Test",
		Description: "A test token",
		Scopes: []string{
			"*",
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		_, err := fmt.Fprint(w, `{
			"id": "08fceb797a467c3c23151f3584c31cfaea962e3ca306e3af69c2dab28e8c2e6e",
			"name": "Test",
			"description": "A test token",
			"scopes": [
				"*"
			]
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/tokens/personal", hf)
	defer teardown()

	res, err := client.Tokens.Personal.Create(context.Background(), TokenCreateRequest{
		Name:        "Test",
		Description: "A test token",
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestTokensService_Update(t *testing.T) {
	exp := &Token{
		ID:          "08fceb797a467c3c23151f3584c31cfaea962e3ca306e3af69c2dab28e8c2e6e",
		Name:        "Test",
		Description: "A very good test token",
		Scopes: []string{
			"*",
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		_, err := fmt.Fprint(w, `{
			"id": "08fceb797a467c3c23151f3584c31cfaea962e3ca306e3af69c2dab28e8c2e6e",
			"name": "Test",
			"description": "A very good test token",
			"scopes": [
				"*"
			]
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/tokens/personal/08fceb797a467c3c23151f3584c31cfaea962e3ca306e3af69c2dab28e8c2e6e", hf)
	defer teardown()

	res, err := client.Tokens.Personal.Update(context.Background(), "08fceb797a467c3c23151f3584c31cfaea962e3ca306e3af69c2dab28e8c2e6e", Token{
		Name:        "Michael Doe",
		Description: "A very good test token",
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestTokensService_Delete(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}

	client, teardown := setup(t, "/api/v1/tokens/personal/08fceb797a467c3c23151f3584c31cfaea962e3ca306e3af69c2dab28e8c2e6e", hf)
	defer teardown()

	err := client.Tokens.Personal.Delete(context.Background(), "08fceb797a467c3c23151f3584c31cfaea962e3ca306e3af69c2dab28e8c2e6e")
	require.NoError(t, err)
}
