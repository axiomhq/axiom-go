// The purpose of this example is to show how to send OpenTelemetry metrics to
// Axiom.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/runtime"

	axiotel "github.com/axiomhq/axiom-go/axiom/otel"
)

func main() {
	// Export "AXIOM_DATASET" in addition to the required environment variables.

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	dataset := os.Getenv("AXIOM_DATASET")
	if dataset == "" {
		log.Fatal("AXIOM_DATASET is required")
	}

	// 1. Initialize OpenTelemetry.
	shutdown, err := axiotel.InitMetrics(ctx, dataset, "axiom-otel-example", "v1.0.0")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if shutdownErr := shutdown(); shutdownErr != nil {
			log.Fatal(shutdownErr)
		}
	}()

	// 2. Instrument ⚡
	//
	// For manual instrumentation, refer to the documentation of the
	// OpenTelemetry Go SDK.
	if err = runtime.Start(); err != nil {
		log.Fatal(err)
	} else if err = host.Start(); err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()
}
