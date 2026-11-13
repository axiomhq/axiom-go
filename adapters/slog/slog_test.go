package slog

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime"
	"sync"
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

	var hasRun atomic.Uint64
	hf := func(w http.ResponseWriter, r *http.Request) {
		zsr, err := zstd.NewReader(r.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zsr)
		assert.NoError(t, err)

		testhelper.JSONEqExp(t, exp, string(b), []string{ingest.TimestampField})

		hasRun.Add(1)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
	}

	logger, closeHandler := adapters.Setup(t, hf, setup(t))

	logger.
		With("key", "value").
		Info("my message")

	closeHandler()

	assert.EqualValues(t, 1, hasRun.Load())
}

func TestHandler_Source(t *testing.T) {
	_, file, line, _ := runtime.Caller(0)
	line += 29 // Account for the lines until the log line.
	sourceStr := fmt.Sprintf(`{"file":"%s", "function":"github.com/axiomhq/axiom-go/adapters/slog.TestHandler_Source", "line":%d}`, file, line)

	exp := fmt.Sprintf(`{"_time":"%s","level":"INFO","key":"value","msg":"my message", "source":%s}`,
		time.Now().Format(time.RFC3339Nano), sourceStr)

	var hasRun atomic.Uint64
	hf := func(w http.ResponseWriter, r *http.Request) {
		zsr, err := zstd.NewReader(r.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zsr)
		assert.NoError(t, err)

		str := string(b)

		testhelper.JSONEqExp(t, exp, str, []string{ingest.TimestampField})

		hasRun.Add(1)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
	}

	logger, closeHandler := adapters.Setup(t, hf, setupWithSource(t))

	logger.
		With("key", "value").
		Info("my message")

	closeHandler()

	assert.EqualValues(t, 1, hasRun.Load())
}

func TestHandler_WithError(t *testing.T) {
	exp := fmt.Sprintf(`{"_time":"%s","level":"INFO","key":"value","msg":"my message","error":"this is an error"}`,
		time.Now().Format(time.RFC3339Nano))

	var hasRun atomic.Uint64
	hf := func(w http.ResponseWriter, r *http.Request) {
		zsr, err := zstd.NewReader(r.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zsr)
		assert.NoError(t, err)

		testhelper.JSONEqExp(t, exp, string(b), []string{ingest.TimestampField})

		hasRun.Add(1)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
	}

	logger, closeHandler := adapters.Setup(t, hf, setup(t))

	err := fmt.Errorf("this is an error")

	logger.
		With("key", "value").
		Info("my message", slog.Any("error", err))

	closeHandler()

	assert.EqualValues(t, 1, hasRun.Load())
}

func TestHandler_NoPanicAfterClose(t *testing.T) {
	exp := fmt.Sprintf(`{"_time":"%s","level":"INFO","key":"value","msg":"my message"}`,
		time.Now().Format(time.RFC3339Nano))

	var lines atomic.Uint64
	hf := func(w http.ResponseWriter, r *http.Request) {
		zsr, err := zstd.NewReader(r.Body)
		require.NoError(t, err)

		s := bufio.NewScanner(zsr)
		for s.Scan() {
			testhelper.JSONEqExp(t, exp, s.Text(), []string{ingest.TimestampField})
			lines.Add(1)
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

	assert.EqualValues(t, 1, lines.Load())
}

func TestHandler_Groups(t *testing.T) {
	exp := fmt.Sprintf(`{"_time":"%s","level":"INFO","s":{"a":1,"b":2},"msg":"my message"}`,
		time.Now().Format(time.RFC3339Nano))

	var lines atomic.Uint64
	hf := func(w http.ResponseWriter, r *http.Request) {
		zsr, err := zstd.NewReader(r.Body)
		require.NoError(t, err)

		s := bufio.NewScanner(zsr)
		for s.Scan() {
			testhelper.JSONEqExp(t, exp, s.Text(), []string{ingest.TimestampField})
			lines.Add(1)
		}
		assert.NoError(t, s.Err())

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
	}

	logger, closeHandler := adapters.Setup(t, hf, setup(t))

	ctx := t.Context()

	logger.WithGroup("s").LogAttrs(ctx, slog.LevelInfo, "my message", slog.Int("a", 1), slog.Int("b", 2))
	logger.LogAttrs(ctx, slog.LevelInfo, "my message", slog.Group("s", slog.Int("a", 1), slog.Int("b", 2)))

	closeHandler()

	assert.EqualValues(t, 2, lines.Load())
}

func TestHandler_GroupScopes(t *testing.T) {
	tests := []struct {
		name string
		emit func(*slog.Logger)
	}{
		{
			name: "interleaved attributes and groups",
			emit: func(logger *slog.Logger) {
				base := logger.With("service", "api")
				request := base.WithGroup("request").With("id", "r1")
				request.WithGroup("handler").Info("served", "status", 200)
				request.Info("request", "status", 201)
				base.Info("base", "status", 202)
				logger.WithGroup("request").Info("control", "status", 203)
			},
		},
		{
			name: "empty group and attributes",
			emit: func(logger *slog.Logger) {
				logger.With("service", "api").WithGroup("").With().Info("empty name", "status", 200)
				logger.With("service", "api").WithGroup("request").WithGroup("handler").Info("no attributes")
				logger.WithGroup("request").LogAttrs(t.Context(), slog.LevelInfo, "zero attribute", slog.Attr{})
			},
		},
		{
			name: "explicit named groups",
			emit: func(logger *slog.Logger) {
				logger.With("service", "api").WithGroup("request").
					With(slog.Group("headers", slog.String("accept", "application/json"))).
					WithGroup("handler").Info("named groups", slog.Group("response", slog.Int("status", 200)))
			},
		},
		{
			name: "parent and sibling groups",
			emit: func(logger *slog.Logger) {
				base := logger.With("service", "api").With("region", "east").With("release", "v1")
				left := base.WithGroup("left").With("id", "left")
				right := base.WithGroup("right").With("id", "right")
				left.WithGroup("handler").Info("left", "status", 200)
				right.Info("right", "status", 201)
				base.Info("parent", "status", 202)
			},
		},
		{
			name: "concurrent derived handlers",
			emit: func(logger *slog.Logger) {
				base := logger.With("service", "api").WithGroup("request").
					With("id", "r1").With("method", "GET").With("host", "example")
				var wg sync.WaitGroup
				for i := range 16 {
					wg.Go(func() {
						base.With("worker", i).WithGroup("handler").With("operation", "serve").
							Info(fmt.Sprintf("worker-%d", i), "status", 200)
					})
				}
				wg.Wait()
				base.Info("parent", "status", 201)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reference bytes.Buffer
			tt.emit(slog.New(slog.NewJSONHandler(&reference, nil)))
			expected := make(map[string]map[string]any)
			decoder := json.NewDecoder(&reference)
			for decoder.More() {
				var event map[string]any
				require.NoError(t, decoder.Decode(&event))
				delete(event, slog.TimeKey)
				expected[event[slog.MessageKey].(string)] = event
			}

			handler := &Handler{rootHandler: &rootHandler{
				eventCh: make(chan axiom.Event, len(expected)),
				closeCh: make(chan struct{}),
			}}
			tt.emit(slog.New(handler))
			require.Len(t, handler.eventCh, len(expected))
			for range len(expected) {
				event := <-handler.eventCh
				delete(event, ingest.TimestampField)
				encoded, err := json.Marshal(event)
				require.NoError(t, err)
				var actual map[string]any
				require.NoError(t, json.Unmarshal(encoded, &actual))
				assert.Equal(t, expected[event[slog.MessageKey].(string)], actual)
			}
		})
	}
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

func setupWithSource(t *testing.T) func(dataset string, client *axiom.Client) (*slog.Logger, func()) {
	return func(dataset string, client *axiom.Client) (*slog.Logger, func()) {
		t.Helper()

		handler, err := New(
			SetClient(client),
			SetDataset(dataset),
			SetAddSource(),
		)
		require.NoError(t, err)
		t.Cleanup(handler.Close)

		return slog.New(handler), handler.Close
	}
}
