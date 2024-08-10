//go:build go1.23

// The purpose of this example is to show how to query a dataset using the Axiom
// Processing Language (APL).
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/axiomhq/axiom-go/axiom"
)

func main() {
	// Export "AXIOM_DATASET" in addition to the required environment variables.

	dataset := os.Getenv("AXIOM_DATASET")
	if dataset == "" {
		log.Fatal("AXIOM_DATASET is required")
	}

	ctx := context.Background()

	// 1. Initialize the Axiom API client.
	client, err := axiom.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// 2. Query all events using APL ⚡
	apl := fmt.Sprintf("['%s']", dataset) // E.g. ['test']
	res, err := client.Query(ctx, apl)
	if err != nil {
		log.Fatal(err)
	} else if res.Status.RowsMatched == 0 {
		log.Fatal("No matches found")
	}

	// 3. Print the queried results by creating a iterator for the rows from the
	// tabular query result (as it is organized in columns) and iterating over
	// the rows.
	for row := range res.Tables[0].Rows() {
		_, _ = fmt.Println(row)
	}
}
