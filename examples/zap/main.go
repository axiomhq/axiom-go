// The purpose of this example is to show how to integrate with zap.
package main

import (
	"log"

	"go.uber.org/zap"

	adapter "github.com/axiomhq/axiom-go/adapters/zap"
)

func main() {
	// Export `AXIOM_TOKEN`, `AXIOM_ORG_ID` and `AXIOM_DATASET` for Axiom Cloud
	// Export `AXIOM_URL`, `AXIOM_TOKEN` and `AXIOM_DATASET` for Axiom Selfhost

	// 1. Setup the Axiom core for zap.
	core, err := adapter.New()
	if err != nil {
		log.Fatal(err)
	}

	// 2. Spawn the logger.
	logger := zap.New(core)

	// 3. Have all logs flushed before the application exits.
	defer func() {
		// Make sure to handle this error in production, just in case syncing
		// fails.
		_ = logger.Sync()
	}()

	// 4. Log ⚡
	logger.Info("This is awesome!", zap.String("mood", "hyped"))
}
