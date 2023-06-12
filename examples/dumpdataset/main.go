// The purpose of this example is to show how to dump a whole dataset into a
// file. Events are ND-JSON encoded.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/query"
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

	// 2. Open the file to write to.
	f, err := os.Create(fmt.Sprintf("%s_dump.json", dataset))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			log.Fatal(closeErr)
		}
	}()

	// 3. Query in a loop until we have no more results.
	var (
		enc = json.NewEncoder(f)
		apl = fmt.Sprintf("['%s'] | sort by _time asc", dataset) // E.g. ['test'] | sort by _time asc
		res *query.Result
	)
	for {
		// 4. If we have a cursor, we need to set it in the query options.
		var opts []query.Option
		if res != nil && res.Status.MaxCursor != "" {
			opts = append(opts, query.SetCursor(res.Status.MaxCursor, false))
		}

		// 5. Query all events using APL ⚡
		if res, err = client.Query(context.Background(), apl, opts...); err != nil {
			log.Fatal(err)
		} else if len(res.Matches) == 0 {
			break
		}

		// 6. Write the queried results to the file.
		for _, match := range res.Matches {
			if err = enc.Encode(match); err != nil {
				log.Fatal(err)
			}
		}
	}
}
