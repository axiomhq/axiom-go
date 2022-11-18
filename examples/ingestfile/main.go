// The purpose of this example is to show how to stream the contents of a JSON
// logfile and gzip them on the fly.
package main

import (
	"context"
	"log"
	"os"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/ingest"
)

func main() {
	// Export `AXIOM_DATASET` in addition to the required environment variables.

	dataset := os.Getenv("AXIOM_DATASET")
	if dataset == "" {
		log.Fatal("AXIOM_DATASET is required")
	}

	// 1. Open the file to ingest.
	f, err := os.Open("logs.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// 2. Wrap it in a gzip enabled reader.
	r, err := axiom.GzipEncoder(f)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Initialize the Axiom API client.
	client, err := axiom.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// 4. Ingest ⚡
	// Note the JSON content type and Gzip content encoding being set because
	// the client does not auto sense them.
	res, err := client.Datasets.Ingest(context.Background(), dataset, r, axiom.JSON, axiom.Gzip,
		// Set a custom timestamp field (default used by the server is "_time").
		ingest.SetTimestampField("timestamp"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// 5. Make sure everything went smoothly.
	for _, fail := range res.Failures {
		log.Print(fail.Error)
	}
}
