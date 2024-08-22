package slog

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

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

	assert.Equal(t, "test", handler.datasetName)
}

func TestHandler(t *testing.T) {
	exp := fmt.Sprintf(`{"_time":"%s","level":"INFO","key":"value","msg":"my message"}`,
		time.Now().Format(time.RFC3339Nano))

	var hasRun uint64
	hf := func(w http.ResponseWriter, r *http.Request) {
		zsr, err := zstd.NewReader(r.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zsr)
		assert.NoError(t, err)

		testhelper.JSONEqExp(t, exp, string(b), []string{ingest.TimestampField})

		atomic.AddUint64(&hasRun, 1)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
	}

	logger, closeHandler := adapters.Setup(t, hf, setup(t))

	logger.
		With("key", "value").
		Info("my message")

	closeHandler()

	assert.EqualValues(t, 1, atomic.LoadUint64(&hasRun))
}

func TestHandler_NoPanicAfterClose(t *testing.T) {
	exp := fmt.Sprintf(`{"_time":"%s","level":"INFO","key":"value","msg":"my message"}`,
		time.Now().Format(time.RFC3339Nano))

	var lines uint64
	hf := func(w http.ResponseWriter, r *http.Request) {
		zsr, err := zstd.NewReader(r.Body)
		require.NoError(t, err)

		s := bufio.NewScanner(zsr)
		for s.Scan() {
			testhelper.JSONEqExp(t, exp, s.Text(), []string{ingest.TimestampField})
			atomic.AddUint64(&lines, 1)
		}
		assert.NoError(t, s.Err())

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
	}

	logger, closeHandler := adapters.Setup(t, hf, setup(t))

	logger.
		With("key", "value").
		Info("my message")

	closeHandler()

	// This should be a no-op.
	logger.
		With("key", "value").
		Info("my message")

	assert.EqualValues(t, 1, atomic.LoadUint64(&lines))
}

func TestHandler_Groups(t *testing.T) {
	exp := fmt.Sprintf(`{"_time":"%s","level":"INFO","s":{"a":1,"b":2},"msg":"my message"}`,
		time.Now().Format(time.RFC3339Nano))

	var lines uint64
	hf := func(w http.ResponseWriter, r *http.Request) {
		zsr, err := zstd.NewReader(r.Body)
		require.NoError(t, err)

		s := bufio.NewScanner(zsr)
		for s.Scan() {
			testhelper.JSONEqExp(t, exp, s.Text(), []string{ingest.TimestampField})
			atomic.AddUint64(&lines, 1)
		}
		assert.NoError(t, s.Err())

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
	}

	logger, closeHandler := adapters.Setup(t, hf, setup(t))

	ctx := context.Background()

	logger.WithGroup("s").LogAttrs(ctx, slog.LevelInfo, "my message", slog.Int("a", 1), slog.Int("b", 2))
	logger.LogAttrs(ctx, slog.LevelInfo, "my message", slog.Group("s", slog.Int("a", 1), slog.Int("b", 2)))

	closeHandler()

	assert.EqualValues(t, 2, atomic.LoadUint64(&lines))
}

func setup(t *testing.T) func(dataset string, client *axiom.Client) (*slog.Logger, func()) {
	return func(dataset string, client *axiom.Client) (*slog.Logger, func()) {
		t.Helper()

		handler, err := New(
			SetClient(client),
			SetDataset(dataset),
		)
		require.NoError(t, err)
		t.Cleanup(handler.Close)

		return slog.New(handler), handler.Close
	}
}
