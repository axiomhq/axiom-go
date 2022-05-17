package auth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/axiomhq/axiom-go/axiom/auth/pkce"
)

const (
	authPath    = "/oauth/authorize"
	tokenPath   = "/oauth/token" //nolint:gosec // Sigh, this is not a hardcoded credential...
	successPath = "/oauth/success"
	errorPath   = "/oauth/error"

	clientID = "13c885a8-f46a-4424-82d2-883cf7ccfe49"
)

// LoginFunc is a function that is called with the URL the user has to visit in
// order to authenticate.
type LoginFunc func(ctx context.Context, loginURL string) error

// Login to the given Axiom deployment and retrieve a Personal Access Token in
// exchange. This will execute the OAuth2 Authorization Code Flow with Proof Key
// for Code Exchange (PKCE).
func Login(ctx context.Context, baseURL string, loginFunc LoginFunc) (string, error) {
	u, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return "", err
	}

	authURL, err := u.Parse(authPath)
	if err != nil {
		return "", err
	}

	tokenURL, err := u.Parse(tokenPath)
	if err != nil {
		return "", err
	}

	successURL, err := u.Parse(successPath)
	if err != nil {
		return "", err
	}

	errorURL, err := u.Parse(errorPath)
	if err != nil {
		return "", err
	}

	// Start a listener for the callback. We need to do this early in order to
	// construct the correct URL for the callback.
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", err
	}
	defer lis.Close()

	redirectURL, err := url.Parse("http://" + lis.Addr().String())
	if err != nil {
		return "", err
	}

	config := &oauth2.Config{
		ClientID: clientID,
		Endpoint: oauth2.Endpoint{
			AuthURL:   authURL.String(),
			TokenURL:  tokenURL.String(),
			AuthStyle: oauth2.AuthStyleInParams,
		},
		RedirectURL: redirectURL.String(),
		Scopes:      []string{"*"},
	}

	// Create the PKCE Code Verifier and S256 Code Challenge.
	method := pkce.MethodS256
	codeVerifier, err := pkce.New()
	if err != nil {
		return "", err
	}

	state, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	loginURL := config.AuthCodeURL(state.String(),
		codeVerifier.Challenge(method).AuthCodeOption(),
		method.AuthCodeOption(),
	)

	if err = loginFunc(ctx, loginURL); err != nil {
		return "", err
	}

	var (
		token         *oauth2.Token
		callbackErrCh = make(chan error)
	)
	callbackHandlerHf := func(w http.ResponseWriter, r *http.Request) {
		defer close(callbackErrCh)

		if r.Method != http.MethodGet {
			callbackErrCh <- errors.New("invalid method")
			http.Redirect(w, r, errorURL.String(), http.StatusFound)
			return
		}

		// Make sure state matches.
		if r.FormValue("state") != state.String() {
			callbackErrCh <- errors.New("invalid state")
			http.Redirect(w, r, errorURL.String(), http.StatusFound)
			return
		}

		// In case we have an error from the authorization server, return it and
		// redirect to the error page.
		if r.Form.Has("error") {
			serverErr := fmt.Errorf("oauth2 authorization error %q: %s", r.FormValue("error"), r.FormValue("error_description"))
			callbackErrCh <- serverErr
			http.Redirect(w, r, errorURL.String(), http.StatusFound)
			return
		}

		code := r.FormValue("code")
		if code == "" {
			callbackErrCh <- errors.New("missing authorization code")
			http.Redirect(w, r, errorURL.String(), http.StatusFound)
			return
		}

		var exchangeErr error
		if token, exchangeErr = config.Exchange(r.Context(), code, codeVerifier.AuthCodeOption()); exchangeErr != nil {
			callbackErrCh <- exchangeErr
			http.Redirect(w, r, errorURL.String(), http.StatusFound)
			return
		}

		http.Redirect(w, r, successURL.String(), http.StatusFound)
	}

	srv := http.Server{
		Addr:        lis.Addr().String(),
		Handler:     http.HandlerFunc(callbackHandlerHf),
		BaseContext: func(net.Listener) context.Context { return ctx },
	}
	defer srv.Close()

	srvErrCh := make(chan error)
	go func() {
		if serveErr := srv.Serve(lis); serveErr != nil && serveErr != http.ErrServerClosed {
			srvErrCh <- serveErr
		}
		close(srvErrCh)
	}()

	select {
	case <-ctx.Done():
		close(callbackErrCh)
		return "", ctx.Err()
	case err = <-srvErrCh:
		close(callbackErrCh)
		return "", err
	case callbackErr := <-callbackErrCh:
		if shutdownErr := srv.Shutdown(ctx); callbackErr != nil {
			return "", callbackErr
		} else if shutdownErr != nil {
			return "", shutdownErr
		}
	}

	return token.AccessToken, nil
}
