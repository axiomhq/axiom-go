package logrus

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/internal/test/testhelper"
)

// TestNew makes sure New() picks up the `AXIOM_DATASET` environment variable.
func TestNew(t *testing.T) {
	testhelper.SafeClearEnv(t)

	os.Setenv("AXIOM_TOKEN", "xaat-test")
	os.Setenv("AXIOM_ORG_ID", "123")

	handler, err := New()
	require.ErrorIs(t, err, ErrMissingDatasetName)
	require.Nil(t, handler)

	os.Setenv("AXIOM_DATASET", "test")

	handler, err = New()
	require.NoError(t, err)
	require.NotNil(t, handler)
	handler.Close()

	assert.Equal(t, "test", handler.datasetName)
}

func TestHook(t *testing.T) {
	now := time.Now()

	exp := fmt.Sprintf(`{"_time":"%s","severity":"info","key":"value","message":"my message"}`,
		now.Format(time.RFC3339Nano))

	var hasRun uint64
	hf := func(w http.ResponseWriter, r *http.Request) {
		zsr, err := zstd.NewReader(r.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zsr)
		assert.NoError(t, err)

		assert.JSONEq(t, exp, string(b))

		atomic.AddUint64(&hasRun, 1)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
	}

	logger, teardown := setup(t, hf)
	defer teardown()

	logger.
		WithTime(now).
		WithField("key", "value").
		Info("my message")

	// Wait for timer based hook flush.
	time.Sleep(1250 * time.Millisecond)

	assert.EqualValues(t, 1, atomic.LoadUint64(&hasRun))
}

func TestHook_FlushFullBatch(t *testing.T) {
	var lines uint64
	hf := func(w http.ResponseWriter, r *http.Request) {
		zsr, err := zstd.NewReader(r.Body)
		require.NoError(t, err)

		s := bufio.NewScanner(zsr)
		for s.Scan() {
			atomic.AddUint64(&lines, 1)
		}
		assert.NoError(t, s.Err())

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
	}

	logger, teardown := setup(t, hf)
	defer teardown()

	for i := 0; i <= 1024; i++ {
		logger.Info("my message")
	}

	// Let the server process.
	time.Sleep(250 * time.Millisecond)

	// Should have a full batch right away.
	assert.EqualValues(t, 1024, atomic.LoadUint64(&lines))

	// Wait for timer based hook flush.
	time.Sleep(1250 * time.Millisecond)

	// Should have received the last event.
	assert.EqualValues(t, 1025, atomic.LoadUint64(&lines))
}

// setup sets up a test HTTP server along with a logrus logger that is
// configured to talk to that test server through an Axiom hook. Tests should
// pass a handler function which provides the response for the API method being
// tested.
func setup(t *testing.T, h http.HandlerFunc) (*logrus.Logger, func()) {
	t.Helper()

	srv := httptest.NewServer(h)

	client, err := axiom.NewClient(
		axiom.SetURL(srv.URL),
		axiom.SetAccessToken("xaat-test"),
		axiom.SetClient(srv.Client()),
	)
	require.NoError(t, err)

	hook, err := New(
		SetClient(client),
		SetDataset("test"),
	)
	require.NoError(t, err)

	logger := logrus.New()
	logger.AddHook(hook)

	// We don't want output in tests.
	logger.Out = io.Discard

	return logger, func() { hook.Close(); srv.Close() }
}
