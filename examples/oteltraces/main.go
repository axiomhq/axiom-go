// The purpose of this example is to show how to send OpenTelemetry traces to
// Axiom.
package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	axiotel "github.com/axiomhq/axiom-go/axiom/otel"
)

func main() {
	// Export "AXIOM_DATASET" in addition to the required environment variables.

	ctx := context.Background()

	dataset := os.Getenv("AXIOM_DATASET")
	if dataset == "" {
		log.Fatal("AXIOM_DATASET is required")
	}

	// 1. Initialize OpenTelemetry.
	stop, err := axiotel.InitTracing(ctx, dataset, "axiom-otel-example", "v1.0.0")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if stopErr := stop(); stopErr != nil {
			log.Fatal(stopErr)
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
	defer span.End()

	span.SetAttributes(attribute.Key("testset").String("value"))

	time.Sleep(time.Millisecond * 100)
}
