package adapters

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/axiom"
)

// Setup sets up a test http server that serves the given handler function. It
// uses the given setup function to retrieve the adapter to be tested that must
// be configured to talk to the client passed to the setup function.
func Setup[T any](t *testing.T, hf http.HandlerFunc, setupFunc func(dataset string, client *axiom.Client) (T, func())) (T, func()) {
	t.Helper()

	client := SetupClient(t, hf)

	return setupFunc("test", client)
}

// SetupClient sets up a test http server and returns a client configured to
// talk to it. Keep-alive connections are disabled to ensure compatibility with
// synctest.
func SetupClient(t *testing.T, hf http.HandlerFunc) *axiom.Client {
	t.Helper()

	srv := httptest.NewServer(hf)
	t.Cleanup(srv.Close)

	httpClient := srv.Client()
	httpClient.Transport.(*http.Transport).DisableKeepAlives = true

	client, err := axiom.NewClient(
		axiom.SetNoEnv(),
		axiom.SetURL(srv.URL),
		axiom.SetToken("xaat-test"),
		axiom.SetClient(httpClient),
	)
	require.NoError(t, err)

	return client
}
