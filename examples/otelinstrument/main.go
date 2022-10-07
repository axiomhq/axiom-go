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
	close, err := axiotel.InitTracing(ctx, "axiom-otel-example", "v1.0.0")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := close(); err != nil {
			log.Fatal(err)
		}
	}()

	// 3. Initialize the Axiom API client.
	client, err := axiom.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// 4. Use the client as usual ⚡
	// This will send traces to the configured OpenTelemetry collector (in this
	// case Axiom itself).
	user, err := client.Users.Current(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Hello %s!\n", user.Name)
}
