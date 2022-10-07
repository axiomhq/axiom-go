// The purpose of this example is to show how to send events to Axiom.
package main

import (
	"context"
	"log"
	"os"

	"github.com/axiomhq/axiom-go/axiom"
)

func main() {
	// Export `AXIOM_DATASET` in addition to the required environment variables.

	dataset := os.Getenv("AXIOM_DATASET")
	if dataset == "" {
		log.Fatal("AXIOM_DATASET is required")
	}

	// 1. Initialize the Axiom API client.
	client, err := axiom.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// 2. Ingest ⚡
	events := []axiom.Event{
		{"foo": "bar"},
		{"bar": "foo"},
	}
	res, err := client.Datasets.IngestEvents(context.Background(), dataset, events)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Make sure everything went smoothly.
	for _, fail := range res.Failures {
		log.Print(fail.Error)
	}
}
