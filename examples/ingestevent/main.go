// The purpose of this example is to show how to send events to Axiom.
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/ingest"
)

func main() {
	// Export "AXIOM_DATASET" in addition to the required environment variables.

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
	//
	// Set the events timestamp by specifying the "_time" field the server uses
	// by default. Can be changed by using the [ingest.SetTimestampField] option
	// when ingesting.
	events := []axiom.Event{
		{ingest.TimestampField: time.Now(), "foo": "bar"},
		{ingest.TimestampField: time.Now(), "bar": "foo"},
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
