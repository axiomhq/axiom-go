package otel_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/log"
	logglobal "go.opentelemetry.io/otel/log/global"

	axiotel "github.com/axiomhq/axiom-go/axiom/otel"
)

func TestLogging(t *testing.T) {
	var handlerCalled uint32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint32(&handlerCalled, 1)

		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/logs", r.URL.Path)
		assert.Equal(t, "application/x-protobuf", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-dataset", r.Header.Get("X-Axiom-Dataset"))

		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	ctx := t.Context()

	stop, err := axiotel.InitLogging(ctx, "test-dataset", "axiom-go-otel-test", "v1.0.0",
		axiotel.SetURL(srv.URL),
		axiotel.SetToken("xaat-test-token"),
		axiotel.SetNoEnv(),
	)
	require.NoError(t, err)
	require.NotNil(t, stop)

	t.Cleanup(func() {
		assert.NoError(t, stop())
	})

	logger := logglobal.GetLoggerProvider().Logger("test")

	var record log.Record
	record.SetBody(log.StringValue("test log message"))
	record.SetSeverity(log.SeverityInfo)
	logger.Emit(ctx, record)

	// Stop logger provider which flushes all log records.
	require.NoError(t, stop())

	assert.EqualValues(t, 1, atomic.LoadUint32(&handlerCalled))
}
