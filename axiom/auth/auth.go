package auth

import (
	"bytes"
	"context"
	_ "embed"
	"html/template"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/axiomhq/axiom-go/axiom/auth/pkce"
)

const (
	authPath  = "/oauth/authorize"
	tokenPath = "/oauth/token" //nolint:gosec // Sigh, this is not a hardcoded credential...

	clientID = "13c885a8-f46a-4424-82d2-883cf7ccfe49"
)

//go:embed callback.html.tmpl
var callbackTmpl string

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
		code   string
		doneCh = make(chan struct{})
	)
	srv := http.Server{
		Addr:        lis.Addr().String(),
		Handler:     callbackHandler(&code, state.String(), doneCh),
		BaseContext: func(net.Listener) context.Context { return ctx },
	}
	defer srv.Close()

	errCh := make(chan error)
	go func() {
		if serveErr := srv.Serve(lis); serveErr != nil && serveErr != http.ErrServerClosed {
			errCh <- serveErr
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		close(doneCh)
		return "", ctx.Err()
	case err = <-errCh:
		close(doneCh)
		return "", err
	case <-doneCh:
		if shutdownErr := srv.Shutdown(ctx); shutdownErr != nil {
			return "", err
		}
	}

	token, err := config.Exchange(ctx, code, codeVerifier.AuthCodeOption())
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

func callbackHandler(code *string, state string, doneCh chan<- struct{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			code := http.StatusMethodNotAllowed
			http.Error(w, http.StatusText(code), code)
			return
		}

		// Make sure state matches.
		if r.FormValue("state") != state {
			code := http.StatusBadRequest
			http.Error(w, http.StatusText(code), code)
			return
		}

		tmpl, err := template.New("callback").Parse(callbackTmpl)
		if err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		data := make(map[string]string)
		if oauthErr := r.FormValue("error"); oauthErr != "" {
			data["error"] = oauthErr
			data["error_description"] = r.FormValue("error_description")
		} else if *code = r.FormValue("code"); *code == "" {
			code := http.StatusBadRequest
			http.Error(w, http.StatusText(code), code)
			return
		}

		var buf bytes.Buffer
		if err = tmpl.Execute(&buf, data); err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		_, _ = io.Copy(w, &buf)

		close(doneCh)
	}
}
