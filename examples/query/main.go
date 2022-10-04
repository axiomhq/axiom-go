// The purpose of this example is to show how to query a dataset.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/query"
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

	// 2. Query all events in the last minute ⚡
	res, err := client.Datasets.Query(context.Background(), dataset, query.Query{
		StartTime: time.Now().Add(-time.Minute),
		EndTime:   time.Now(),
	}, query.Options{})
	if err != nil {
		log.Fatal(err)
	} else if len(res.Matches) == 0 {
		log.Fatal("No matches found")
	}

	// 3. Print the queried results.
	for _, match := range res.Matches {
		fmt.Println(match.Data)
	}
}
