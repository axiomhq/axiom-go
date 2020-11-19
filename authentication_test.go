package axiom

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticationService_Valid(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.WriteHeader(http.StatusOK)
	}

	client, teardown := setup(t, "/api/v1/version", hf)
	defer teardown()

	valid, err := client.Authentication.Valid(context.Background())
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestAuthenticationService_Valid_Unauthenticated(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.WriteHeader(http.StatusForbidden)
	}

	client, teardown := setup(t, "/api/v1/version", hf)
	defer teardown()

	valid, err := client.Authentication.Valid(context.Background())
	assert.Error(t, err)
	assert.False(t, valid)
}
