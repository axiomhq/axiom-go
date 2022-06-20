// The purpose of this example is to show how to stream the contents of a JSON
// logfile and gzip them on the fly.
package main

import (
	"context"
	"log"
	"os"

	"github.com/axiomhq/axiom-go/axiom"
)

func main() {
	// Export `AXIOM_TOKEN`, `AXIOM_ORG_ID` (when using a personal token) and
	// `AXIOM_DATASET` for Axiom Cloud.
	// Export `AXIOM_URL`, `AXIOM_TOKEN` and `AXIOM_DATASET` for Axiom Selfhost.

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
	ev := axiom.Event{
		"foo": "bar",
	}
	res, err := client.Datasets.IngestEvents(context.Background(), dataset, axiom.IngestOptions{}, ev)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Make sure everything went smoothly.
	for _, fail := range res.Failures {
		log.Print(fail.Error)
	}
}
