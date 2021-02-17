// The purpose of this example is to show how to stream the contents of a JSON
// logfile and gzip them on the fly.
package main

import (
	"compress/gzip"
	"context"
	"log"
	"os"

	"github.com/axiomhq/axiom-go/axiom"
)

func main() {
	var (
		deploymentURL = os.Getenv("AXM_DEPLOYMENT_URL")
		accessToken   = os.Getenv("AXM_ACCESS_TOKEN")
	)

	// 1. Open the file to ingest.
	f, err := os.Open("logs.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// 2. Wrap it in a gzip enabled reader.
	r, err := axiom.GZIPStreamer(f, gzip.BestSpeed)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Initialize the Axiom API client.
	client, err := axiom.NewClient(deploymentURL, accessToken)
	if err != nil {
		log.Fatal(err)
	}

	// 4. Ingest ⚡
	// Note the JSON content type and GZIP content encoding being set because
	// the client does not auto sense them.
	res, err := client.Datasets.Ingest(context.Background(), "test", r, axiom.JSON, axiom.GZIP, axiom.IngestOptions{})
	if err != nil {
		log.Fatal(err)
	}

	// 5. Make sure everything went smoothly.
	for _, fail := range res.Failures {
		log.Print(fail.Error)
	}
}
