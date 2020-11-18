package axiom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// endpoint is a test url that won't be called.
	endpoint = "http://axiom.local"
	// accessToken is a placeholder access token.
	accessToken = "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
)

// SetStrictDecoding is a special testing only client option that failes JSON
// response decoding if fields not present in the destination struct are
// encountered.
func SetStrictDecoding() Option {
	return func(c *Client) error {
		c.strictDecoding = true
		return nil
	}
}

func TestNewClient(t *testing.T) {
	client, err := NewClient(endpoint, accessToken)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Are endpoints/resources present?
	assert.NotNil(t, client.Dashboards)
	assert.NotNil(t, client.Datasets)
	assert.NotNil(t, client.Monitors)
	assert.NotNil(t, client.Notifiers)
	assert.NotNil(t, client.StarredQueries)
	assert.NotNil(t, client.Teams)
	assert.NotNil(t, client.Tokens.Ingest)
	assert.NotNil(t, client.Tokens.Personal)
	assert.NotNil(t, client.Users)
	assert.NotNil(t, client.Version)
	assert.NotNil(t, client.VirtualFields)

	// Is default configuration present?
	assert.Equal(t, endpoint, client.baseURL.String())
	assert.NotEmpty(t, client.userAgent)
	assert.NotEmpty(t, client.accessToken)
	assert.NotNil(t, client.httpClient)
}

func TestNewCloudClient(t *testing.T) {
	client, err := NewCloudClient(accessToken)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Is default configuration present?
	assert.Equal(t, CloudURL, client.baseURL.String())
}

func TestClient_Options_SetClient(t *testing.T) {
	client, _ := NewClient(endpoint, accessToken)

	exp := &http.Client{
		Timeout: 0,
	}
	opt := SetClient(exp)

	err := client.Options(opt)
	assert.NoError(t, err)

	assert.Equal(t, exp, client.httpClient)
}

func TestClient_Options_SetUserAgent(t *testing.T) {
	client, _ := NewClient(endpoint, accessToken)

	exp := "axiom-go/1.0.0"
	opt := SetUserAgent(exp)

	err := client.Options(opt)
	assert.NoError(t, err)

	assert.Equal(t, exp, client.userAgent)
}

func TestClient_newRequest_BadURL(t *testing.T) {
	client, _ := NewClient(endpoint, accessToken)

	_, err := client.newRequest(context.Background(), http.MethodGet, ":", nil)
	assert.Error(t, err)

	if assert.IsType(t, err, new(url.Error)) {
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
	client, _ := NewClient(endpoint, accessToken)

	req, err := client.newRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	assert.Empty(t, req.Body)
}

func TestClient_do(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		_, _ = fmt.Fprint(w, `{"A":"a"}`)
	}

	client, teardown := setup(t, "/", hf)
	defer teardown()

	type foo struct {
		A string
	}

	req, err := client.newRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	var body foo
	err = client.do(req, &body)
	require.NoError(t, err)

	assert.Equal(t, foo{"a"}, body)
}

func TestClient_do_ioWriter(t *testing.T) {
	content := `{"A":"a"}`

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		_, _ = fmt.Fprint(w, content)
	}

	client, teardown := setup(t, "/", hf)
	defer teardown()

	req, err := client.newRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = client.do(req, &buf)
	require.NoError(t, err)

	assert.Equal(t, content, buf.String())
}

func TestClient_do_HTTPError(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		httpErr := Error{
			Message: http.StatusText(http.StatusBadRequest),
		}
		err := json.NewEncoder(w).Encode(httpErr)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/", hf)
	defer teardown()

	req, err := client.newRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	err = client.do(req, nil)
	require.NoError(t, err)
}

func TestClient_do_Unauthenticated(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}

	client, teardown := setup(t, "/", hf)
	defer teardown()

	req, err := client.newRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	err = client.do(req, nil)
	require.Equal(t, err, ErrUnauthenticated)
}

func TestClient_do_RedirectLoop(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	client, teardown := setup(t, "/", hf)
	defer teardown()

	req, err := client.newRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	err = client.do(req, nil)
	require.Error(t, err)

	assert.IsType(t, err, new(url.Error))
}

// setup sets up a test HTTP server along with a client that is configured to
// talk to that test server. Tests should pass a handler function which provides
// the response for the API method being tested.
func setup(t *testing.T, path string, handler http.HandlerFunc) (*Client, func()) {
	t.Helper()

	r := http.NewServeMux()
	r.HandleFunc(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.Header.Get("authorization"), "no authorization header present on the request")
		assert.Equal(t, r.Header.Get("accept"), "application/json", "bad accept header present on the request")
		assert.Equal(t, r.Header.Get("user-agent"), "axiom-go", "bad user-agent header present on the request")

		if r.ContentLength > 0 {
			assert.NotEmpty(t, r.Header.Get("Content-Type"), "no Content-Type header present on the request")
		}

		handler.ServeHTTP(w, r)
	}))
	srv := httptest.NewServer(r)

	client, err := NewClient(srv.URL, accessToken, SetClient(srv.Client()), SetStrictDecoding())
	require.NoError(t, err)

	return client, func() { srv.Close() }
}

func mustTimeParse(t *testing.T, layout, value string) time.Time {
	ts, err := time.Parse(layout, value)
	require.NoError(t, err)
	require.False(t, ts.IsZero())
	return ts
}
