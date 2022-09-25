// The purpose of this example is to show how to send OpenTelemetry traces to
// Axiom.
package main

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	axiotel "github.com/axiomhq/axiom-go/axiom/otel"
)

func main() {
	// Export `AXIOM_TOKEN` and `AXIOM_ORG_ID` (when using a personal token) for
	// Axiom Cloud.
	// Export `AXIOM_URL` and `AXIOM_TOKEN` for Axiom Selfhost.

	ctx := context.Background()

	// 1. Initialize OpenTelemetry.
	close, err := axiotel.InitTracing(ctx, "axiom-otel-example", "v1.0.0")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := close(); err != nil {
			log.Fatal(err)
		}
	}()

	// 2. Instrument ⚡
	tr := otel.Tracer("main")

	ctx, span := tr.Start(ctx, "foo")
	defer span.End()

	bar(ctx)
}

func bar(ctx context.Context) {
	tr := otel.Tracer("bar")
	_, span := tr.Start(ctx, "bar")
	span.SetAttributes(attribute.Key("testset").String("value"))
	defer span.End()

	time.Sleep(time.Millisecond * 100)
}
