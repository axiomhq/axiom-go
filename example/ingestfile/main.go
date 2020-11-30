// The purpose of this example is to show how to stream the contents of a JSON
// logfile and gzip it on the fly.
package main

import (
	"compress/gzip"
	"context"
	"log"
	"os"

	"github.com/axiomhq/axiom-go/axiom"
)

func main() {
	f, err := os.Open("logs.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r, err := axiom.GZIPStreamer(f, gzip.BestSpeed)
	if err != nil {
		log.Fatal(err)
	}

	client, err := axiom.NewClient("https://my-axiom.example.com", "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX")
	if err != nil {
		log.Fatal(err)
	}

	res, err := client.Datasets.Ingest(context.Background(), "test", r, axiom.JSON, axiom.GZIP, axiom.IngestOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for _, fail := range res.Failures {
		log.Print(fail.Error)
	}
}
