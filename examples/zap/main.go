// The purpose of this example is to show how to integrate with zap.
package main

import (
	"log"
	"os"

	"go.uber.org/zap"

	adapter "github.com/axiomhq/axiom-go/adapters/zap"
)

func main() {
	var (
		deploymentURL = os.Getenv("AXM_DEPLOYMENT_URL")
		accessToken   = os.Getenv("AXM_ACCESS_TOKEN")
		dataset       = os.Getenv("AXM_DATASET")
	)

	// 1. Setup the Axiom core for zap.
	core, err := adapter.New(deploymentURL, accessToken, dataset)
	if err != nil {
		log.Fatal(err)
	}

	// 2. Spawn the logger.
	logger := zap.New(core)

	// 3. Have all logs flushed before the application exits.
	defer func() {
		// Make sure to handle this error, just in case syncing fails.
		_ = logger.Sync()
	}()

	// 4. Log ⚡
	logger.Info("This is awesome!", zap.String("mood", "hyped"))
}
