package apex

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/axiom"
)

// TestNew makes sure New() picks up the `AXIOM_DATASET` environment variable.
func TestNew(t *testing.T) {
	os.Clearenv()

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

func TestHandler(t *testing.T) {
	now := time.Now()

	exp := fmt.Sprintf(`{"_time":"%s","severity":"info","key":"value","message":"my message"}`,
		now.Format(time.RFC3339Nano))

	var hasRun uint64
	hf := func(w http.ResponseWriter, r *http.Request) {
		gzr, err := gzip.NewReader(r.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(gzr)
		assert.NoError(t, err)

		JSONEqExp(t, exp, string(b), []string{axiom.TimestampField})

		atomic.AddUint64(&hasRun, 1)

		_, _ = w.Write([]byte("{}"))
	}

	logger, teardown := setup(t, hf)
	defer teardown()

	logger.
		WithField("key", "value").
		Info("my message")

	// Wait for timer based handler flush.
	time.Sleep(1250 * time.Millisecond)

	assert.EqualValues(t, 1, atomic.LoadUint64(&hasRun))
}

func TestHandler_FlushFullBatch(t *testing.T) {
	var lines uint64
	hf := func(w http.ResponseWriter, r *http.Request) {
		gzr, err := gzip.NewReader(r.Body)
		require.NoError(t, err)

		s := bufio.NewScanner(gzr)
		for s.Scan() {
			atomic.AddUint64(&lines, 1)
		}
		assert.NoError(t, s.Err())

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

	// Wait for timer based handler flush.
	time.Sleep(1250 * time.Millisecond)

	// Should have received the last event.
	assert.EqualValues(t, 1025, atomic.LoadUint64(&lines))
}

// setup sets up a test HTTP server along with a apex logger that is
// configured to talk to that test server through an Axiom handler. Tests should
// pass a handler function which provides the response for the API method being
// tested.
func setup(t *testing.T, h http.HandlerFunc) (*log.Logger, func()) {
	t.Helper()

	srv := httptest.NewServer(h)

	client, err := axiom.NewClient(
		axiom.SetURL(srv.URL),
		axiom.SetAccessToken("xaat-test"),
		axiom.SetClient(srv.Client()),
	)
	require.NoError(t, err)

	handler, err := New(
		SetClient(client),
		SetDataset("test"),
	)
	require.NoError(t, err)

	logger := &log.Logger{
		Handler: handler,
		Level:   log.InfoLevel,
	}

	return logger, func() { handler.Close(); srv.Close() }
}

// JSONEqExp is like assert.JSONEq() but excludes the given fields.
func JSONEqExp(t assert.TestingT, expected string, actual string, excludedFields []string, msgAndArgs ...interface{}) bool {
	type tHelper interface {
		Helper()
	}

	if h, ok := t.(tHelper); ok {
		h.Helper()
	}

	var expectedJSONAsInterface, actualJSONAsInterface map[string]interface{}

	if err := json.Unmarshal([]byte(expected), &expectedJSONAsInterface); err != nil {
		return assert.Fail(t, fmt.Sprintf("Expected value ('%s') is not valid json.\nJSON parsing error: '%s'", expected, err.Error()), msgAndArgs...)
	}

	if err := json.Unmarshal([]byte(actual), &actualJSONAsInterface); err != nil {
		return assert.Fail(t, fmt.Sprintf("Input ('%s') needs to be valid json.\nJSON parsing error: '%s'", actual, err.Error()), msgAndArgs...)
	}

	for _, excludedField := range excludedFields {
		delete(expectedJSONAsInterface, excludedField)
		delete(actualJSONAsInterface, excludedField)
	}

	return assert.Equal(t, expectedJSONAsInterface, actualJSONAsInterface, msgAndArgs...)
}
