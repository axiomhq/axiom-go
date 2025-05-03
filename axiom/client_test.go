package axiom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/internal/config"
	"github.com/axiomhq/axiom-go/internal/test/testhelper"
)

const (
	endpoint       = "http://api.axiom.local"
	apiToken       = "xaat-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
	personalToken  = "xapt-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX" //nolint:gosec // Chill, it's just testing.
	organizationID = "awkward-identifier-c3po"
)

var tokenRe = regexp.MustCompile("xa(a|p)t-[a-zA-z0-9]{8}-[a-zA-z0-9]{4}-[a-zA-z0-9]{4}-[a-zA-z0-9]{4}-[a-zA-z0-9]{12}")

// SetStrictDecoding is a special testing-only client option that - when set to
// 'true' - failes JSON response decoding if fields not present in the
// destination struct are encountered.
func SetStrictDecoding(b bool) Option {
	return func(c *Client) error {
		c.strictDecoding = b
		return nil
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		environment map[string]string
		options     []Option
		err         error
	}{
		{
			name: "no environment no options",
			err:  config.ErrMissingToken,
		},
		{
			name: "no environment token option",
			options: []Option{
				SetToken(personalToken),
			},
			err: config.ErrMissingOrganizationID,
		},
		{
			name: "no environment token option with API token",
			options: []Option{
				SetToken(apiToken),
			},
		},
		{
			name: "organizationID environment no options",
			environment: map[string]string{
				"AXIOM_TOKEN": personalToken,
			},
			err: config.ErrMissingOrganizationID,
		},
		{
			name: "organizationID environment no options with API token",
			environment: map[string]string{
				"AXIOM_TOKEN": apiToken,
			},
		},
		{
			name: "no environment token and organizationID option",
			options: []Option{
				SetToken(personalToken),
				SetOrganizationID(organizationID),
			},
		},
		{
			name: "token and organizationID environment no options",
			environment: map[string]string{
				"AXIOM_TOKEN":  personalToken,
				"AXIOM_ORG_ID": organizationID,
			},
		},
		{
			name: "token environment organizationID option",
			environment: map[string]string{
				"AXIOM_TOKEN": personalToken,
			},
			options: []Option{
				SetOrganizationID(organizationID),
			},
		},
		{
			name: "organizationID environment token option",
			environment: map[string]string{
				"AXIOM_ORG_ID": organizationID,
			},
			options: []Option{
				SetToken(personalToken),
			},
		},
		{
			name: "no environment url and token option",
			options: []Option{
				SetURL(endpoint),
				SetToken(personalToken),
			},
		},
		{
			name: "url and token environment no options",
			environment: map[string]string{
				"AXIOM_URL":   endpoint,
				"AXIOM_TOKEN": personalToken,
			},
		},
		{
			name: "token and organizationID environment apiUrl option",
			environment: map[string]string{
				"AXIOM_TOKEN":  personalToken,
				"AXIOM_ORG_ID": organizationID,
			},
			options: []Option{
				SetURL(config.APIURL().String()),
			},
		},
		{
			name: "token and organizationID environment enhanced apiUrl option",
			environment: map[string]string{
				"AXIOM_TOKEN":  personalToken,
				"AXIOM_ORG_ID": organizationID,
			},
			options: []Option{
				SetURL(config.APIURL().String() + "/"),
			},
		},
		{
			name: "apiUrl token and organizationID environment no options",
			environment: map[string]string{
				"AXIOM_URL":    config.APIURL().String(),
				"AXIOM_TOKEN":  personalToken,
				"AXIOM_ORG_ID": organizationID,
			},
		},
		{
			name: "enhanced apiUrl, token and organizationID environment no options",
			environment: map[string]string{
				"AXIOM_URL":    config.APIURL().String() + "/",
				"AXIOM_TOKEN":  personalToken,
				"AXIOM_ORG_ID": organizationID,
			},
		},
		{
			name: "dev url, token and organizationID environment no options",
			environment: map[string]string{
				"AXIOM_URL":    "https://dev.axiom.co",
				"AXIOM_TOKEN":  personalToken,
				"AXIOM_ORG_ID": organizationID,
			},
		},
		{
			name: "apiUrl and token environment organizationID option",
			environment: map[string]string{
				"AXIOM_URL":   config.APIURL().String(),
				"AXIOM_TOKEN": personalToken,
			},
			options: []Option{
				SetOrganizationID(organizationID),
			},
		},
		{
			name: "token and organizationID environment noEnv option",
			environment: map[string]string{
				"AXIOM_TOKEN":  personalToken,
				"AXIOM_ORG_ID": organizationID,
			},
			options: []Option{
				SetNoEnv(),
			},
			err: config.ErrMissingToken,
		},
		{
			name: "no environment noEnv, apiUrl and token option with API token",
			options: []Option{
				SetNoEnv(),
				SetURL(config.APIURL().String()),
				SetToken(apiToken),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.SafeClearEnv(t)

			for k, v := range tt.environment {
				t.Setenv(k, v)
			}

			client, err := NewClient(tt.options...)
			if tt.err != nil {
				assert.ErrorIs(t, err, tt.err)
			} else if assert.NoError(t, err) {
				assert.Regexp(t, tokenRe, client.config.Token())
				assert.NotEmpty(t, client.config.BaseURL())
			}
		})
	}
}

func TestNewClient_Valid(t *testing.T) {
	client := newClient(t)

	// Are endpoints/resources present?
	assert.NotNil(t, client.Datasets)
	assert.NotNil(t, client.Organizations)
	assert.NotNil(t, client.Users)

	// Is default configuration present?
	assert.Equal(t, endpoint, client.config.BaseURL().String())
	assert.Equal(t, personalToken, client.config.Token())
	assert.Empty(t, client.config.OrganizationID())
	assert.NotNil(t, client.httpClient)
	assert.NotEmpty(t, client.userAgent)
	assert.False(t, client.strictDecoding)
	assert.True(t, client.noEnv) // Disabled for testing.
	assert.False(t, client.noRetry)
}

func TestClient_Options_SetToken(t *testing.T) {
	client := newClient(t)

	exp := personalToken
	opt := SetToken(exp)

	err := client.Options(opt)
	assert.NoError(t, err)

	assert.Equal(t, exp, client.config.Token())
}

func TestClient_Options_SetClient(t *testing.T) {
	client := newClient(t)

	exp := &http.Client{
		Timeout: time.Second,
	}
	opt := SetClient(exp)

	err := client.Options(opt)
	assert.NoError(t, err)

	assert.Equal(t, exp, client.httpClient)
}

func TestClient_Options_SetPersonalTokenConfig(t *testing.T) {
	client := newClient(t)

	opt := SetPersonalTokenConfig(personalToken, organizationID)

	err := client.Options(opt)
	assert.NoError(t, err)

	assert.Equal(t, personalToken, client.config.Token())
	assert.Equal(t, organizationID, client.config.OrganizationID())
}

func TestClient_Options_SetAPITokenConfig(t *testing.T) {
	client := newClient(t)

	opt := SetAPITokenConfig(apiToken)

	err := client.Options(opt)
	assert.NoError(t, err)

	assert.Equal(t, apiToken, client.config.Token())
	assert.Empty(t, client.config.OrganizationID())
}

func TestClient_Options_SetOrganizationID(t *testing.T) {
	client := newClient(t)

	exp := organizationID
	opt := SetOrganizationID(exp)

	err := client.Options(opt)
	assert.NoError(t, err)

	assert.Equal(t, exp, client.config.OrganizationID())
}

func TestClient_Options_SetURL(t *testing.T) {
	client := newClient(t)

	exp := endpoint
	opt := SetURL(exp)

	err := client.Options(opt)
	assert.NoError(t, err)

	assert.Equal(t, exp, client.config.BaseURL().String())
}

func TestClient_Options_SetUserAgent(t *testing.T) {
	client := newClient(t)

	exp := "axiom-go/1.0.0"
	opt := SetUserAgent(exp)

	err := client.Options(opt)
	assert.NoError(t, err)

	assert.Equal(t, exp, client.userAgent)
}

func TestClient_newRequest_BadURL(t *testing.T) {
	client := newClient(t)

	_, err := client.NewRequest(context.Background(), http.MethodGet, ":", nil)
	assert.Error(t, err)

	if assert.IsType(t, new(url.Error), err) {
		urlErr := err.(*url.Error)
		assert.Equal(t, urlErr.Op, "parse")
	}
}

// If a nil body is passed to NewRequest, make sure that nil is also passed to
// http.NewRequest. In most cases, passing an io.Reader that returns no content
// is fine, since there is no difference between an HTTP request body that is an
// empty string versus one that is not set at all. However in certain cases,
// intermediate systems may treat these differently resulting in subtle errors.
func TestClient_newRequest_EmptyBody(t *testing.T) {
	client := newClient(t)

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	assert.Empty(t, req.Body)
}

func TestClient_do(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, _ = fmt.Fprint(w, `{"A":"a"}`)
	}

	client := setup(t, "GET /", hf)

	type foo struct {
		A string
	}

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	var body foo
	_, err = client.Do(req, &body)
	require.NoError(t, err)

	assert.Equal(t, foo{"a"}, body)
}

func TestClient_do_ioWriter(t *testing.T) {
	content := `{"A":"a"}`

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, _ = fmt.Fprint(w, content)
	}

	client := setup(t, "GET /", hf)

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	var buf bytes.Buffer
	_, err = client.Do(req, &buf)
	require.NoError(t, err)

	assert.Equal(t, content, buf.String())
}

func TestClient_do_HTTPError(t *testing.T) {
	hf := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Axiom-Trace-Id", "abc")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(http.StatusText(http.StatusBadRequest)))
	}

	client := setup(t, "GET /", hf)

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	if _, err = client.Do(req, nil); assert.ErrorIs(t, err, HTTPError{
		Status:  http.StatusBadRequest,
		Message: http.StatusText(http.StatusBadRequest),
	}) {
		assert.EqualError(t, err, "API error 400: Bad Request")
		assert.Equal(t, "abc", err.(HTTPError).TraceID)
	}
}

func TestClient_do_HTTPError_Typed(t *testing.T) {
	hf := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(http.StatusText(http.StatusForbidden)))
	}

	client := setup(t, "GET /", hf)

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	if _, err = client.Do(req, nil); assert.ErrorIs(t, err, ErrUnauthorized) {
		assert.EqualError(t, err, "API error 403: Forbidden")
	}
}

func TestClient_do_HTTPError_JSON(t *testing.T) {
	hf := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", mediaTypeJSON)
		w.Header().Set("X-Axiom-Trace-Id", "abc")
		w.WriteHeader(http.StatusBadRequest)

		assert.NoError(t, json.NewEncoder(w).Encode(HTTPError{
			Message: "This is a Bad Request error",
		}))
	}

	client := setup(t, "GET /", hf)

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	if _, err = client.Do(req, nil); assert.ErrorIs(t, err, HTTPError{
		Status:  http.StatusBadRequest,
		Message: "This is a Bad Request error",
	}) {
		assert.EqualError(t, err, "API error 400: This is a Bad Request error")
		assert.Equal(t, "abc", err.(HTTPError).TraceID)
	}
}

func TestClient_do_HTTPError_Unauthenticated(t *testing.T) {
	hf := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", mediaTypeJSON)
		w.WriteHeader(http.StatusUnauthorized)

		assert.NoError(t, json.NewEncoder(w).Encode(HTTPError{
			Message: "You are not allowed here!",
		}))
	}

	client := setup(t, "GET /", hf)

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	_, err = client.Do(req, nil)
	assert.ErrorIs(t, err, ErrUnauthenticated)
}

func TestClient_do_RateLimit(t *testing.T) {
	// Truncated time for testing as the [LimitError.Error] method uses
	// [time.Until] which will yield different milliseconds when comparing the
	// time values with [errors.Is].
	reset := time.Now().Add(time.Hour).Truncate(time.Second)

	expErr := LimitError{
		HTTPError: HTTPError{
			Status:  http.StatusTooManyRequests,
			Message: "limit exceeded",
			TraceID: "abc",
		},

		Limit: Limit{
			Scope:     LimitScopeAnonymous,
			Limit:     1_000,
			Remaining: 0,
			Reset:     reset,

			limitType: limitRate,
		},
	}

	hf := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", mediaTypeJSON)
		w.Header().Set("X-RateLimit-Scope", "anonymous")
		w.Header().Set("X-RateLimit-Limit", "1000")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))
		w.Header().Set("X-Axiom-Trace-Id", "abc")
		w.WriteHeader(http.StatusTooManyRequests)
		assert.NoError(t, json.NewEncoder(w).Encode(HTTPError{
			Message: "limit exceeded",
		}))
	}

	client := setup(t, "GET /", hf)

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	// Request should fail with a "*LimitError".
	resp, err := client.Do(req, nil)

	if assert.ErrorIs(t, err, expErr) {
		assert.EqualError(t, err, "rate limit exceeded: try again in 59m59s")
		assert.Equal(t, "abc", err.(LimitError).TraceID)
	}
	assert.Equal(t, expErr.Limit, resp.Limit)
}

func TestClient_do_RedirectLoop(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	client := setup(t, "GET /", hf)

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	_, err = client.Do(req, nil)
	assert.IsType(t, new(url.Error), err)
}

func TestClient_do_ValidOnlyAPITokenPaths(t *testing.T) {
	hf := func(http.ResponseWriter, *http.Request) {}

	tests := []string{
		"/v2/datasets/test/query",
		"/v2/datasets/_apl",
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			client := setup(t, tt, hf)

			err := client.Options(SetToken("xaat-123"))
			require.NoError(t, err)

			req, err := client.NewRequest(context.Background(), http.MethodGet, tt, nil)
			require.Nil(t, err)

			_, err = client.Do(req, nil)
			require.NoError(t, err)
		})
	}
}

func TestClient_do_Backoff(t *testing.T) {
	payload := `{"foo":"bar"}`

	var (
		internalServerErrorCalled bool
		badGatewayCalled          bool
		gatewayTimeoutCalled      bool
	)
	hf := func(w http.ResponseWriter, r *http.Request) {
		header := http.StatusOK
		switch {
		case !internalServerErrorCalled:
			internalServerErrorCalled = true
			header = http.StatusInternalServerError
		case !badGatewayCalled:
			badGatewayCalled = true
			header = http.StatusBadGateway
		case !gatewayTimeoutCalled:
			gatewayTimeoutCalled = true
			header = http.StatusGatewayTimeout
		}

		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		assert.Equal(t, payload, string(b))

		w.WriteHeader(header)
	}

	client := setup(t, "POST /", hf)

	// Wrap with an io.TeeReader as http.NewRequest checks for some special
	// readers it can read in full to optimize the request.
	var r io.Reader = strings.NewReader(payload)
	r = io.TeeReader(r, io.Discard)
	req, err := client.NewRequest(context.Background(), http.MethodPost, "/", r)
	require.NoError(t, err)

	// Make sure the request body can be re-read.
	getBodyCounter := 0
	req.GetBody = func() (io.ReadCloser, error) {
		getBodyCounter++
		return io.NopCloser(strings.NewReader(payload)), nil
	}

	resp, err := client.Do(req, nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, internalServerErrorCalled)
	assert.True(t, badGatewayCalled)
	assert.True(t, gatewayTimeoutCalled)
	assert.Equal(t, 3, getBodyCounter)
}

func TestClient_do_Backoff_NoRetryOn400(t *testing.T) {
	var currentCalls int
	hf := func(w http.ResponseWriter, _ *http.Request) {
		currentCalls++
		w.WriteHeader(http.StatusBadRequest)
	}

	client := setup(t, "GET /", hf)

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	resp, err := client.Do(req, nil)
	require.Error(t, err, "got status code 400")

	assert.Equal(t, 1, currentCalls)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// setup sets up a test HTTP server along with a client that is configured to
// talk to that test server. Tests should pass a handler function which provides
// the response for the API method being tested.
func setup(t *testing.T, path string, handler http.HandlerFunc) *Client {
	t.Helper()

	r := http.NewServeMux()
	r.HandleFunc(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.Header.Get("Authorization"), "no authorization header present on the request")
		assert.Equal(t, mediaTypeJSON, r.Header.Get("Accept"), "bad accept header present on the request")
		assert.Equal(t, "axiom-go", r.Header.Get("User-Agent"), "bad user-agent header present on the request")
		if organizationIDHeader := r.Header.Get("X-Axiom-Org-Id"); organizationIDHeader != "" {
			assert.Equal(t, organizationID, organizationIDHeader, "bad x-axiom-org-id header present on the request")
		}

		if r.ContentLength > 0 {
			assert.NotEmpty(t, r.Header.Get("Content-Type"), "no Content-Type header present on the request")
		}

		handler.ServeHTTP(w, r)
	}))

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	client, err := NewClient(
		SetURL(srv.URL),
		SetToken(personalToken),
		SetOrganizationID(organizationID),
		SetClient(srv.Client()),
		SetStrictDecoding(true),
		SetNoEnv(),
	)
	require.NoError(t, err)

	return client
}

// newClient returns a new client with stub properties for testing methods that
// don't actually make a http call.
func newClient(t *testing.T) *Client {
	t.Helper()

	client, err := NewClient(
		SetURL(endpoint),
		SetToken(personalToken),
		SetNoEnv(),
	)
	require.NoError(t, err)

	return client
}
