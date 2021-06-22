// The purpose of this example is to show how to query a dataset using the Axiom
// Processing Language (APL).
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/apl"
)

func main() {
	// Export `AXIOM_TOKEN`, `AXIOM_ORG_ID` and `AXIOM_DATASET` for Axiom Cloud
	// Export `AXIOM_URL`, `AXIOM_TOKEN` and `AXIOM_DATASET` for Axiom Selfhost

	dataset := os.Getenv("AXIOM_DATASET")
	if dataset == "" {
		log.Fatal("AXIOM_DATASET is required")
	}

	// 1. Initialize the Axiom API client.
	client, err := axiom.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// 2. Query all events using APL ⚡
	aplQuery := fmt.Sprintf("['%s']", dataset) // E.g. ['test']
	res, err := client.Datasets.APLQuery(context.Background(), aplQuery, apl.Options{})
	if err != nil {
		log.Fatal(err)
	}

	// 3. Print the queried results.
	for _, match := range res.Result.Matches {
		log.Print(match.Data)
	}
}
