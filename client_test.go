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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// endpoint is a test url that won't be called.
	endpoint = "http://axiom.local/"
	// accessToken is a dummy access token.
	accessToken = "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient(endpoint, accessToken)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Are endpoints/resources present?
	assert.Equal(t, &datasetsService{client: client}, client.Datasets)

	// Is default configuration present?
	assert.Equal(t, endpoint, client.baseURL.String())
	assert.NotEmpty(t, client.userAgent)
	assert.NotNil(t, client.httpClient)
}

func TestNewRequest_BadURL(t *testing.T) {
	client, err := NewClient(endpoint, accessToken)
	require.NoError(t, err)

	_, err = client.newRequest(context.Background(), http.MethodGet, ":", nil)
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
func TestNewRequest_EmptyBody(t *testing.T) {
	client, err := NewClient(endpoint, accessToken)
	require.NoError(t, err)

	req, err := client.newRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	assert.Empty(t, req.Body)
}

func TestDo(t *testing.T) {
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

func TestDo_ioWriter(t *testing.T) {
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

func TestDo_HTTPError(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		httpErr := Error{
			StatusCode: http.StatusBadRequest,
			Message:    http.StatusText(http.StatusBadRequest),
		}
		err := json.NewEncoder(w).Encode(httpErr)
		require.NoError(t, err)
	}

	client, teardown := setup(t, "/", hf)
	defer teardown()

	req, err := client.newRequest(context.Background(), http.MethodGet, "/", nil)
	require.NoError(t, err)

	err = client.do(req, nil)
	require.NoError(t, err)
}

func TestDo_RedirectLoop(t *testing.T) {
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
		require.NotEmpty(t, r.Header.Get("Authorization"), "no authentication present on the request")

		handler.ServeHTTP(w, r)
	}))
	srv := httptest.NewServer(r)

	client, err := NewClient(srv.URL, accessToken, SetClient(srv.Client()))
	require.NoError(t, err)

	return client, func() { srv.Close() }
}
