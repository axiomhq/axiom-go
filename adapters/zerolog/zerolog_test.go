package zerolog

import (
	"bufio"
	"io"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/rs/zerolog"
	l "github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/internal/test/adapters"
	"github.com/axiomhq/axiom-go/internal/test/testhelper"
)

// TestNew makes sure New() picks up the "AXIOM_DATASET" environment variable.
func TestNew(t *testing.T) {
	testhelper.SafeClearEnv(t)

	t.Setenv("AXIOM_TOKEN", "xaat-test")
	t.Setenv("AXIOM_ORG_ID", "123")

	writer, err := New()
	require.ErrorIs(t, err, ErrMissingDataset)
	require.Nil(t, writer)

	t.Setenv("AXIOM_DATASET", "test")

	writer, err = New()
	require.NoError(t, err)
	require.NotNil(t, writer)

	assert.Equal(t, "test", writer.dataset)
}

func TestBasicHook(t *testing.T) {
	exp := `{"key":"value", "level":"info", "logger":"zerolog", "message":"my message"}`

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

	logger, closeHook := adapters.Setup(t, hf, setup(t))

	logger.Info().Str("key", "value").Msg("my message")

	closeHook()

	assert.EqualValues(t, 1, atomic.LoadUint64(&hasRun))

	// test can log after closing the adapter
	logger.Info().Str("key", "value").Msg("my unseen message")
	logger.Info().Str("key", "value").Msg("my unseen message 2")
	logger.Info().Str("key", "value").Msg("my unseen message 3")
}

func TestHook_FlushFullBatch(t *testing.T) {
	exp := `{"key":"value", "level":"info", "logger":"zerolog", "message":"my message"}`

	var lines uint64
	hf := func(w http.ResponseWriter, r *http.Request) {
		zsr, err := zstd.NewReader(r.Body)
		require.NoError(t, err)

		s := bufio.NewScanner(zsr)
		for s.Scan() {
			assert.JSONEq(t, exp, s.Text())
			atomic.AddUint64(&lines, 1)
		}
		assert.NoError(t, s.Err())

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
	}

	logger, _ := adapters.Setup(t, hf, setup(t))

	for range 10_001 {
		logger.Info().Str("key", "value").Msg("my message")
	}

	// Let the server process.
	time.Sleep(time.Millisecond * 750)

	// Should have a full batch right away.
	assert.EqualValues(t, 10_000, atomic.LoadUint64(&lines))

	// Wait for timer based hook flush.
	time.Sleep(time.Second + time.Millisecond*250)

	// Should have received the last event.
	assert.EqualValues(t, 10_001, atomic.LoadUint64(&lines))
}

func setup(t *testing.T) func(dataset string, client *axiom.Client) (*zerolog.Logger, func()) {
	return func(dataset string, client *axiom.Client) (*zerolog.Logger, func()) {
		t.Helper()

		writer, err := New(
			SetClient(client),
			SetDataset(dataset),
		)
		require.NoError(t, err)
		t.Cleanup(func() { writer.Close() })

		// use io.Discard here to squash logs in tests
		l.Logger = zerolog.New(io.MultiWriter(writer, io.Discard)).With().Logger()

		return &l.Logger, func() { writer.Close() }
	}
}
