package zap_test

import (
	"log"

	"go.uber.org/zap"

	adapter "github.com/axiomhq/axiom-go/adapters/zap"
)

func Example() {
	// Export `AXIOM_TOKEN`, `AXIOM_ORG_ID` (when using a personal token) and
	// `AXIOM_DATASET` for Axiom Cloud.
	// Export `AXIOM_URL`, `AXIOM_TOKEN` and `AXIOM_DATASET` for Axiom Selfhost.

	core, err := adapter.New()
	if err != nil {
		log.Fatal(err)
	}

	logger := zap.New(core)
	defer func() {
		if syncErr := logger.Sync(); syncErr != nil {
			log.Fatal(syncErr)
		}
	}()

	logger.Info("This is awesome!", zap.String("mood", "hyped"))
	logger.Warn("This is no that awesome...", zap.String("mood", "worried"))
	logger.Error("This is rather bad.", zap.String("mood", "depressed"))
}
