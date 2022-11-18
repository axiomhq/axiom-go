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

	// 1. Initialize the Axiom API client.
	client, err := axiom.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// 2. Query all events using APL ⚡
	apl := fmt.Sprintf("['%s']", dataset) // E.g. ['test']
	res, err := client.Datasets.Query(context.Background(), apl)
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
