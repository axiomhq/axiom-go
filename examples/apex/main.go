// The purpose of this example is to show how to integrate with apex.
package main

import (
	"fmt"
	"os"

	"github.com/apex/log"

	adapter "github.com/axiomhq/axiom-go/adapters/apex"
)

func main() {
	var (
		deploymentURL = os.Getenv("AXIOM_URL")
		accessToken   = os.Getenv("AXIOM_TOKEN")
		dataset       = os.Getenv("AXIOM_DATASET")
	)

	// 1. Setup the Axiom handler for apex.
	handler, err := adapter.New(deploymentURL, accessToken, dataset)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	// 2. Have all logs flushed before the application exits.
	defer handler.Close()

	// 3. Set the Axiom handler as handler for apex.
	log.SetHandler(handler)

	// 4. Log ⚡
	log.WithField("mood", "hyped").Info("This is awesome!")
}
