package slog_test

import (
	"log"
	"log/slog"

	adapter "github.com/axiomhq/axiom-go/adapters/slog"
)

func Example() {
	// Export "AXIOM_DATASET" in addition to the required environment variables.

	handler, err := adapter.New()
	if err != nil {
		log.Fatal(err.Error())
	}
	defer handler.Close()

	logger := slog.New(handler)

	logger.Info("This is awesome!", "mood", "hyped")
	logger.With("mood", "worried").Warn("This is no that awesome...")
	logger.Error("This is rather bad.", slog.String("mood", "depressed"))
}
