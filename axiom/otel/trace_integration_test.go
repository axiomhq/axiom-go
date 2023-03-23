//go:build integration

package otel_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	axiotel "github.com/axiomhq/axiom-go/axiom/otel"
)

func TestTracingIntegration(t *testing.T) {
	ctx := context.Background()

	stop, err := axiotel.InitTracing(ctx, "axiom-go-otel-test", "v1.0.0")
	require.NoError(t, err)
	require.NotNil(t, stop)

	t.Cleanup(func() {
		require.NoError(t, stop())
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
}
