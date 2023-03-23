package otel_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	axiotel "github.com/axiomhq/axiom-go/axiom/otel"
)

func TestTracing(t *testing.T) {
	var handlerCalled uint32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint32(&handlerCalled, 1)

		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/traces", r.URL.Path)
		assert.Equal(t, "application/x-protobuf", r.Header.Get("Content-Type"))

		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	ctx := context.Background()

	stop, err := axiotel.InitTracing(ctx, "axiom-go-otel-test", "v1.0.0",
		axiotel.SetURL(srv.URL),
		axiotel.SetToken("xaat-test-token"),
		axiotel.SetNoEnv(),
	)
	require.NoError(t, err)
	require.NotNil(t, stop)

	t.Cleanup(func() {
		_ = stop()
	})

	bar := func(ctx context.Context) {
		tr := otel.Tracer("bar")
		_, span := tr.Start(ctx, "bar")
		span.SetAttributes(attribute.Key("testset").String("value"))
		defer span.End()

		time.Sleep(time.Millisecond * 100)
	}

	tr := otel.Tracer("main")

	ctx, span := tr.Start(ctx, "foo")
	defer span.End()

	bar(ctx)

	// Stop tracer which flushes all spans.
	require.NoError(t, stop())

	assert.EqualValues(t, 1, atomic.LoadUint32(&handlerCalled))
}
