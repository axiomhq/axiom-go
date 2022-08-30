package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/axiom/auth"
	"github.com/axiomhq/axiom-go/axiom/auth/pkce"
)

func TestLogin(t *testing.T) {
	var (
		globalRedirectURI   string
		globalCodeChallenge string
	)
	authHf := func(w http.ResponseWriter, r *http.Request) {
		// Correct query parameters are present (or not)?
		assert.Equal(t, "123", r.FormValue("client_id"))
		assert.Empty(t, r.FormValue("client_secret"))
		assert.Equal(t, "*", r.FormValue("scope"))
		assert.Equal(t, "code", r.FormValue("response_type"))
		assert.Contains(t, r.URL.Query(), "redirect_uri")
		assert.Contains(t, r.URL.Query(), "state")
		assert.Contains(t, r.URL.Query(), "code_challenge")
		assert.Contains(t, "S256", r.FormValue("code_challenge_method"))

		// Save some global state.
		globalRedirectURI = r.FormValue("redirect_uri")
		globalCodeChallenge = r.FormValue("code_challenge")

		redirectURI, err := url.ParseRequestURI(r.FormValue("redirect_uri"))
		require.NoError(t, err)

		q := redirectURI.Query()
		q.Set("code", "test-code")
		q.Set("state", r.FormValue("state"))
		redirectURI.RawQuery = q.Encode()

		http.Redirect(w, r, redirectURI.String(), http.StatusFound)
	}

	tokenHf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "123", r.FormValue("client_id"))
		assert.Equal(t, "authorization_code", r.FormValue("grant_type"))
		assert.Equal(t, "test-code", r.FormValue("code"))
		assert.Equal(t, globalRedirectURI, r.FormValue("redirect_uri"))
		assert.Contains(t, r.Form, "code_verifier")

		// Server side PKCE verification.
		codeVerifier := pkce.VerifierFromString(r.FormValue("code_verifier"))
		codeChallenge := pkce.ChallengeFromString(globalCodeChallenge)

		assert.True(t, codeChallenge.Verify(codeVerifier, pkce.MethodS256))

		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Content-Type", "application/json")

		_, _ = w.Write([]byte(`{
			"access_token": "test-token",
			"token_type": "bearer"
		}`))
	}

	successCh := make(chan struct{})

	r := http.NewServeMux()
	r.Handle("/oauth/authorize", http.HandlerFunc(authHf))
	r.Handle("/oauth/token", http.HandlerFunc(tokenHf))
	r.Handle("/oauth/success", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(successCh)
	}))
	r.Handle("/oauth/error", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not be called")
	}))

	srv := httptest.NewServer(r)
	defer srv.Close()

	loginFunc := func(_ context.Context, loginURL string) error {
		// Assume the user opens the login URL and gives consent.
		go func() {
			resp, err := http.Get(loginURL) //nolint:gosec // This is a test.
			require.NoError(t, err)
			assert.NoError(t, resp.Body.Close())
		}()
		return nil
	}

	token, err := auth.Login(context.Background(), "123", srv.URL, loginFunc)
	require.NoError(t, err)

	assert.Equal(t, "test-token", token)

	<-successCh
}

func TestLogin_AuthorizationError(t *testing.T) {
	authHf := func(w http.ResponseWriter, r *http.Request) {
		redirectURI, err := url.ParseRequestURI(r.FormValue("redirect_uri"))
		require.NoError(t, err)

		q := redirectURI.Query()
		q.Set("error", "access_denied")
		q.Set("error_description", "user denied access")
		q.Set("state", r.FormValue("state"))
		redirectURI.RawQuery = q.Encode()

		http.Redirect(w, r, redirectURI.String(), http.StatusFound)
	}

	errCh := make(chan struct{})

	r := http.NewServeMux()
	r.Handle("/oauth/authorize", http.HandlerFunc(authHf))
	r.Handle("/oauth/token", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not be called")
	}))
	r.Handle("/oauth/success", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not be called")
	}))
	r.Handle("/oauth/error", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(errCh)
	}))

	srv := httptest.NewServer(r)
	defer srv.Close()

	loginFunc := func(_ context.Context, loginURL string) error {
		// Assume the user opens the login URL and gives consent.
		go func() {
			resp, err := http.Get(loginURL) //nolint:gosec // This is a test.
			require.NoError(t, err)
			assert.NoError(t, resp.Body.Close())
		}()
		return nil
	}

	token, err := auth.Login(context.Background(), "123", srv.URL, loginFunc)
	assert.EqualError(t, err, "oauth2 authorization error \"access_denied\": user denied access")
	assert.Empty(t, token)

	<-errCh
}

func TestLogin_ExchangeError(t *testing.T) {
	authHf := func(w http.ResponseWriter, r *http.Request) {
		redirectURI, err := url.ParseRequestURI(r.FormValue("redirect_uri"))
		require.NoError(t, err)

		q := redirectURI.Query()
		q.Set("code", "test-code")
		q.Set("state", r.FormValue("state"))
		redirectURI.RawQuery = q.Encode()

		http.Redirect(w, r, redirectURI.String(), http.StatusFound)
	}

	tokenHf := func(w http.ResponseWriter, r *http.Request) {
		code := http.StatusInternalServerError
		http.Error(w, http.StatusText(code), code)
	}

	errCh := make(chan struct{})

	r := http.NewServeMux()
	r.Handle("/oauth/authorize", http.HandlerFunc(authHf))
	r.Handle("/oauth/token", http.HandlerFunc(tokenHf))
	r.Handle("/oauth/success", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not be called")
	}))
	r.Handle("/oauth/error", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(errCh)
	}))

	srv := httptest.NewServer(r)
	defer srv.Close()

	loginFunc := func(_ context.Context, loginURL string) error {
		// Assume the user opens the login URL and gives consent.
		go func() {
			resp, err := http.Get(loginURL) //nolint:gosec // This is a test.
			require.NoError(t, err)
			assert.NoError(t, resp.Body.Close())
		}()
		return nil
	}

	token, err := auth.Login(context.Background(), "123", srv.URL, loginFunc)
	assert.EqualError(t, err, "oauth2: cannot fetch token: 500 Internal Server Error\nResponse: Internal Server Error\n")
	assert.Empty(t, token)

	<-errCh
}
