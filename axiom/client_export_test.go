package axiom_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/axiom"
)

// TestClient makes sure a user of the package can construct his own requests to
// use with the clients methods.
func TestClient_Manual(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	t.Cleanup(srv.Close)

	client, err := axiom.NewClient(
		axiom.SetURL(srv.URL),
		axiom.SetToken("xapt-123"),
		axiom.SetOrganizationID("123"),
		axiom.SetClient(srv.Client()),
		axiom.SetStrictDecoding(true),
		axiom.SetNoEnv(),
	)
	require.NoError(t, err)

	opts := struct {
		test string `url:"test"`
	}{
		test: "test",
	}

	path, err := axiom.AddURLOptions("/v2/test", opts)
	require.NoError(t, err)

	req, err := client.NewRequest(context.Background(), http.MethodGet, path, nil)
	require.NoError(t, err)

	resp, err := client.Do(req, nil)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestClient makes sure a user of the package can construct his own requests to
// use with the clients Call method.
func TestClient_Call(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	t.Cleanup(srv.Close)

	client, err := axiom.NewClient(
		axiom.SetURL(srv.URL),
		axiom.SetToken("xapt-123"),
		axiom.SetOrganizationID("123"),
		axiom.SetClient(srv.Client()),
		axiom.SetStrictDecoding(true),
		axiom.SetNoEnv(),
	)
	require.NoError(t, err)

	opts := struct {
		test string `url:"test"`
	}{
		test: "test",
	}

	path, err := axiom.AddURLOptions("/v2/test", opts)
	require.NoError(t, err)

	err = client.Call(context.Background(), http.MethodGet, path, nil, nil)
	require.NoError(t, err)
}
