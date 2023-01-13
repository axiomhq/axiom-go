// The purpose of this example is to show how to integrate with zap.
package main

import (
	"log"

	"go.uber.org/zap"

	adapter "github.com/axiomhq/axiom-go/adapters/zap"
)

func main() {
	// Export "AXIOM_DATASET" in addition to the required environment variables.

	// 1. Setup the Axiom core for zap.
	core, err := adapter.New()
	if err != nil {
		log.Fatal(err)
	}

	// 2. Spawn the logger.
	logger := zap.New(core)

	// 3. Have all logs flushed before the application exits.
	//
	// ❗THIS IS IMPORTANT❗ Without it, the logs will not be sent to Axiom as
	// the buffer will not be flushed when the application exits.
	defer func() {
		if syncErr := logger.Sync(); syncErr != nil {
			log.Fatal(syncErr)
		}
	}()

	// 4. Log ⚡
	logger.Info("This is awesome!", zap.String("mood", "hyped"))
	logger.Warn("This is no that awesome...", zap.String("mood", "worried"))
	logger.Error("This is rather bad.", zap.String("mood", "depressed"))
}
