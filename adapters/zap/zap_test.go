package zap

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/klauspost/compress/gzip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/internal/test/testhelper"
)

// TestNew makes sure New() picks up the `AXIOM_DATASET` environment variable.
func TestNew(t *testing.T) {
	testhelper.SafeClearEnv(t)

	os.Setenv("AXIOM_TOKEN", "xaat-test")
	os.Setenv("AXIOM_ORG_ID", "123")

	core, err := New()
	require.ErrorIs(t, err, ErrMissingDatasetName)
	require.Nil(t, core)

	os.Setenv("AXIOM_DATASET", "test")

	core, err = New()
	require.NoError(t, err)
	require.NotNil(t, core)
}

func TestCore(t *testing.T) {
	now := time.Now()

	exp := fmt.Sprintf(`{"_time":"%s","level":"info","key":"value","msg":"my message"}`,
		now.Format(time.RFC3339Nano))

	hasRun := false
	hf := func(w http.ResponseWriter, r *http.Request) {
		gzr, err := gzip.NewReader(r.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(gzr)
		assert.NoError(t, err)

		assert.JSONEq(t, exp, string(b))

		hasRun = true

		w.Header().Set("Content-Type", "application/json")
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

	client, err := axiom.NewClient(
		axiom.SetNoEnv(),
		axiom.SetURL(srv.URL),
		axiom.SetAccessToken("xaat-test"),
		axiom.SetClient(srv.Client()),
	)
	require.NoError(t, err)

	core, err := New(
		SetClient(client),
		SetDataset("test"),
	)
	require.NoError(t, err)

	logger := zap.New(core)

	return logger, func() {
		defer srv.Close()
		require.NoError(t, logger.Sync())
	}
}
