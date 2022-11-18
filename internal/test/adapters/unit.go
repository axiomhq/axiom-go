package adapters

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/axiom"
)

// Setup sets up a test HTTP server that serves the given handler function. It
// uses the given setup function to retrieve the adapter to be tested that must
// be configured to talk to the client passed to the setup function. The test
func Setup[T any](t *testing.T, hf http.HandlerFunc, setupFunc func(dataset string, client *axiom.Client) T) T {
	t.Helper()

	srv := httptest.NewServer(hf)
	t.Cleanup(srv.Close)

	client, err := axiom.NewClient(
		axiom.SetNoEnv(),
		axiom.SetURL(srv.URL),
		axiom.SetToken("xaat-test"),
		axiom.SetClient(srv.Client()),
	)
	require.NoError(t, err)

	return setupFunc("test", client)
}
