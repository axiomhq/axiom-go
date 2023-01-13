// The purpose of this example is to show how to integrate with apex/log.
package main

import (
	"github.com/apex/log"

	adapter "github.com/axiomhq/axiom-go/adapters/apex"
)

func main() {
	// Export "AXIOM_DATASET" in addition to the required environment variables.

	// 1. Setup the Axiom handler for apex.
	handler, err := adapter.New()
	if err != nil {
		log.Fatal(err.Error())
	}

	// 2. Have all logs flushed before the application exits.
	//
	// ❗THIS IS IMPORTANT❗ Without it, the logs will not be sent to Axiom as
	// the buffer will not be flushed when the application exits.
	defer handler.Close()

	// 3. Set the Axiom handler as handler for apex.
	log.SetHandler(handler)

	// 4. Log ⚡
	log.WithField("mood", "hyped").Info("This is awesome!")
	log.WithField("mood", "worried").Warn("This is no that awesome...")
	log.WithField("mood", "depressed").Error("This is rather bad.")
}
