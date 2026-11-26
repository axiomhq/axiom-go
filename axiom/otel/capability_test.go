package otel_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	axiotel "github.com/axiomhq/axiom-go/axiom/otel"
)

func newTestTracerProvider(t *testing.T, recorder *tracetest.SpanRecorder) *sdktrace.TracerProvider {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	tp, err := axiotel.TracerProvider(t.Context(), "test-dataset", "test", "v0.1.0",
		axiotel.SetURL(srv.URL),
		axiotel.SetToken("xaat-test-token"),
		axiotel.SetNoEnv(),
	)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, tp.Shutdown(context.Background())) })

	tp.RegisterSpanProcessor(recorder)
	return tp
}

func TestCapability(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := newTestTracerProvider(t, recorder)

	ctx := axiotel.WithCapability(t.Context(), "test-cap")

	_, span := tp.Tracer("test").Start(ctx, "op")
	span.End()

	require.NoError(t, tp.ForceFlush(t.Context()))

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	assertHasAttribute(t, spans[0], "gen_ai.capability.name", "test-cap")
}

func TestStep(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := newTestTracerProvider(t, recorder)

	ctx := axiotel.WithStep(t.Context(), "test-step")

	_, span := tp.Tracer("test").Start(ctx, "op")
	span.End()

	require.NoError(t, tp.ForceFlush(t.Context()))

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	assertHasAttribute(t, spans[0], "gen_ai.step.name", "test-step")
}

func TestNoBaggage(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := newTestTracerProvider(t, recorder)

	_, span := tp.Tracer("test").Start(t.Context(), "op")
	span.End()

	require.NoError(t, tp.ForceFlush(t.Context()))

	spans := recorder.Ended()
	require.Len(t, spans, 1)

	for _, attr := range spans[0].Attributes() {
		assert.NotEqual(t, "gen_ai.capability.name", string(attr.Key))
		assert.NotEqual(t, "gen_ai.step.name", string(attr.Key))
	}
}

func TestCapabilityAndStep(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := newTestTracerProvider(t, recorder)

	ctx := axiotel.WithCapability(t.Context(), "my-cap")
	ctx = axiotel.WithStep(ctx, "my-step")

	_, span := tp.Tracer("test").Start(ctx, "op")
	span.End()

	require.NoError(t, tp.ForceFlush(t.Context()))

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	assertHasAttribute(t, spans[0], "gen_ai.capability.name", "my-cap")
	assertHasAttribute(t, spans[0], "gen_ai.step.name", "my-step")
}

func assertHasAttribute(t *testing.T, span sdktrace.ReadOnlySpan, key, expected string) {
	t.Helper()
	for _, attr := range span.Attributes() {
		if string(attr.Key) == key {
			assert.Equal(t, expected, attr.Value.AsString())
			return
		}
	}
	t.Errorf("attribute %q not found on span", key)
}
