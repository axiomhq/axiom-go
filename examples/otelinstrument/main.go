// The purpose of this example is to show how to instrument the Axiom Go client
// using OpenTelemetry.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/axiomhq/axiom-go/axiom"
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
	// Note: You can setup OpenTelemetry however you like! This example uses the
	// helper package axiom/otel to initialize OpenTelemetry with Axiom
	// configured as a backend for convenience.
	stop, err := axiotel.InitTracing(ctx, dataset, "axiom-otel-example", "v1.0.0")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if stopErr := stop(); stopErr != nil {
			log.Fatal(stopErr)
		}
	}()

	// 2. Initialize the Axiom API client.
	client, err := axiom.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// 3. Use the client as usual ⚡
	// This will send traces to the configured OpenTelemetry collector (in this
	// case Axiom itself).
	user, err := client.Users.Current(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Hello %s!\n", user.Name)
}
