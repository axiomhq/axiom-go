package auth

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/axiomhq/axiom-go/axiom/auth/pkce"
)

const (
	authPath          = "/oauth/authorize"
	tokenPath         = "/oauth/token" //nolint:gosec // Sigh, this is not a hardcoded credential...
	finalRedirectPath = "/oauth/done"

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

	finalRedirectURL, err := u.Parse(finalRedirectPath)
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
			code := http.StatusMethodNotAllowed
			http.Error(w, http.StatusText(code), code)
			callbackErrCh <- errors.New("invalid method")
			return
		}

		// Make sure state matches.
		if r.FormValue("state") != state.String() {
			code := http.StatusBadRequest
			http.Error(w, http.StatusText(code), code)
			callbackErrCh <- errors.New("invalid state")
			return
		}

		if r.Form.Has("error") {
			q := r.URL.Query()
			q.Set("error", r.FormValue("error"))
			q.Set("error_description", r.FormValue("error_description"))
			finalRedirectURL.RawQuery = q.Encode()
		} else {
			code := r.FormValue("code")
			if code == "" {
				code := http.StatusBadRequest
				http.Error(w, http.StatusText(code), code)
				callbackErrCh <- errors.New("missing authorization code")
				return
			}

			var exchangeErr error
			if token, exchangeErr = config.Exchange(r.Context(), code, codeVerifier.AuthCodeOption()); exchangeErr != nil {
				code := http.StatusBadRequest
				http.Error(w, exchangeErr.Error(), code)
				callbackErrCh <- exchangeErr
				return
			}
		}

		http.Redirect(w, r, finalRedirectURL.String(), http.StatusFound)
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
	case err = <-callbackErrCh:
		if err != nil {
			return "", err
		} else if shutdownErr := srv.Shutdown(ctx); shutdownErr != nil {
			return "", err
		}
	}

	return token.AccessToken, nil
}
