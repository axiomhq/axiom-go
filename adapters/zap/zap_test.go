package zap_test

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	adapter "github.com/axiomhq/axiom-go/adapters/zap"
	"github.com/axiomhq/axiom-go/axiom"
)

func TestCore(t *testing.T) {
	now := time.Now()

	exp := fmt.Sprintf(`{"_time":"%s","level":"info","key":"value","msg":"my message"}`,
		now.Format(time.RFC3339Nano))

	hasRun := false
	hf := func(w http.ResponseWriter, r *http.Request) {
		gzr, err := gzip.NewReader(r.Body)
		require.NoError(t, err)

		b, err := ioutil.ReadAll(gzr)
		assert.NoError(t, err)

		assert.JSONEq(t, exp, string(b))

		hasRun = true

		_, _ = w.Write([]byte("{}"))
	}

	logger, teardown := setup(t, hf)
	defer teardown()

	// Timestamp field is set manually to make the JSONEq assertion pass.
	logger.Info("my message", zap.String("key", "value"), zap.Time(axiom.TimestampField, now))

	require.NoError(t, logger.Sync())

	assert.True(t, hasRun)
}

// setup sets up a test HTTP server along with a zap logger that is configured
// to talk to that test server through an Axiom WriteSyncer. Tests should pass a
// handler function which provides the response for the API method being tested.
func setup(t *testing.T, h http.HandlerFunc) (*zap.Logger, func()) {
	t.Helper()

	srv := httptest.NewServer(h)

	client, err := axiom.NewClient(srv.URL, "", "", axiom.SetClient(srv.Client()))
	require.NoError(t, err)

	core, err := adapter.NewWithClient(client, "test")
	require.NoError(t, err)

	logger := zap.New(core)

	return logger, func() {
		defer srv.Close()
		require.NoError(t, logger.Sync())
	}
}
