// The purpose of this example is to show how to instrument the Axiom Go client
// using OpenTelemetry.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/axiomhq/axiom-go/axiom"
	axiotel "github.com/axiomhq/axiom-go/axiom/otel"
)

func main() {
	ctx := context.Background()

	// 1. Initialize OpenTelemetry.
	// Note: You can setup OpenTelemetry however you like. This example uses
	// helper package axiom/otel to initialize OpenTelemetry with Axiom
	// configured as a backend for convenience.
	close, err := axiotel.InitTracing(ctx, "axiom-otel-example", "v1.0.0")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if closeErr := close(); closeErr != nil {
			log.Fatal(closeErr)
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
