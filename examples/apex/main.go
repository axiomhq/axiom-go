// The purpose of this example is to show how to integrate with apex.
package main

import (
	stdlibLog "log"

	"github.com/apex/log"

	adapter "github.com/axiomhq/axiom-go/adapters/apex"
)

func main() {
	// Export `AXIOM_TOKEN`, `AXIOM_ORG_ID` and `AXIOM_DATASET` for Axiom Cloud.
	// Export `AXIOM_URL`, `AXIOM_TOKEN` and `AXIOM_DATASET` for Axiom Selfhost.

	// 1. Setup the Axiom handler for apex.
	handler, err := adapter.New()
	if err != nil {
		stdlibLog.Fatal(err)
	}

	// 2. Have all logs flushed before the application exits.
	defer handler.Close()

	// 3. Set the Axiom handler as handler for apex.
	log.SetHandler(handler)

	// 4. Log ⚡
	log.WithField("mood", "hyped").Info("This is awesome!")
}
