package otel_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"

	axiotel "github.com/axiomhq/axiom-go/axiom/otel"
)

func TestMetrics(t *testing.T) {
	var handlerCalled uint32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint32(&handlerCalled, 1)

		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/metrics", r.URL.Path)
		assert.Equal(t, "application/x-protobuf", r.Header.Get("Content-Type"))
		assert.Equal(t, "metric-test-dataset", r.Header.Get("X-Axiom-Dataset"))

		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	ctx := context.Background()

	stop, err := axiotel.InitMetrics(ctx, "metric-test-dataset", "axiom-go-otel-metrics-test", "v1.0.0",
		axiotel.SetURL(srv.URL),
		axiotel.SetToken("xaat-test-token"),
		axiotel.SetNoEnv(),
	)
	require.NoError(t, err)
	require.NotNil(t, stop)

	t.Cleanup(func() { _ = stop() })

	meter := otel.Meter("main")

	counter, err := meter.Int64Counter("test")
	require.NoError(t, err)

	counter.Add(ctx, 1)

	// Stop meter which flushes all metrics.
	require.NoError(t, stop())

	assert.EqualValues(t, 1, atomic.LoadUint32(&handlerCalled))
}
