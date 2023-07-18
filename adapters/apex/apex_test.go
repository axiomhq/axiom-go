package apex

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/ingest"
	"github.com/axiomhq/axiom-go/internal/test/adapters"
	"github.com/axiomhq/axiom-go/internal/test/testhelper"
)

// TestNew makes sure New() picks up the "AXIOM_DATASET" environment variable.
func TestNew(t *testing.T) {
	testhelper.SafeClearEnv(t)

	t.Setenv("AXIOM_TOKEN", "xaat-test")
	t.Setenv("AXIOM_ORG_ID", "123")

	handler, err := New()
	require.ErrorIs(t, err, ErrMissingDatasetName)
	require.Nil(t, handler)

	t.Setenv("AXIOM_DATASET", "test")

	handler, err = New()
	require.NoError(t, err)
	require.NotNil(t, handler)
	handler.Close()

	assert.Equal(t, "test", handler.datasetName)
}

func TestHandler(t *testing.T) {
	exp := fmt.Sprintf(`{"_time":"%s","severity":"info","key":"value","message":"my message"}`,
		time.Now().Format(time.RFC3339Nano))

	var hasRun uint64
	hf := func(w http.ResponseWriter, r *http.Request) {
		zsr, err := zstd.NewReader(r.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zsr)
		require.NoError(t, err)

		testhelper.JSONEqExp(t, exp, string(b), []string{ingest.TimestampField})

		atomic.AddUint64(&hasRun, 1)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
	}

	logger, flush := adapters.Setup(t, hf, setup(t))

	logger.
		WithField("key", "value").
		Info("my message")

	flush()

	assert.EqualValues(t, 1, atomic.LoadUint64(&hasRun))
}

func TestHandler_FlushFullBatch(t *testing.T) {
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

	logger, _ := adapters.Setup(t, hf, setup(t))

	for i := 0; i <= 1000; i++ {
		logger.Info("my message")
	}

	// Let the server process.
	time.Sleep(250 * time.Millisecond)

	// Should have a full batch right away.
	assert.EqualValues(t, 1000, atomic.LoadUint64(&lines))

	// Wait for timer based handler flush.
	time.Sleep(1250 * time.Millisecond)

	// Should have received the last event.
	assert.EqualValues(t, 1001, atomic.LoadUint64(&lines))
}

func setup(t *testing.T) func(dataset string, client *axiom.Client) (*log.Logger, func()) {
	return func(dataset string, client *axiom.Client) (*log.Logger, func()) {
		t.Helper()

		handler, err := New(
			SetClient(client),
			SetDataset(dataset),
		)
		require.NoError(t, err)
		t.Cleanup(handler.Close)

		logger := &log.Logger{
			Handler: handler,
			Level:   log.InfoLevel,
		}

		return logger, handler.Close
	}
}
