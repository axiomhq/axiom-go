// The purpose of this example is to show how to integrate with slog.
package main

import (
	"log"
	"log/slog"

	adapter "github.com/axiomhq/axiom-go/adapters/slog"
)

func main() {
	// Export "AXIOM_DATASET" in addition to the required environment variables.

	// 1. Setup the Axiom handler for slog.
	handler, err := adapter.New()
	if err != nil {
		log.Fatal(err.Error())
	}

	// 2. Have all logs flushed before the application exits.
	//
	// ❗THIS IS IMPORTANT❗ Without it, the logs will not be sent to Axiom as
	// the buffer will not be flushed when the application exits.
	defer handler.Close()

	// 3. Create the logger.
	//
	// Go 1.26+: Use slog.NewMultiHandler to fan out to multiple handlers.
	// multi := slog.NewMultiHandler(handler, slog.NewTextHandler(os.Stderr, nil))
	// logger := slog.New(multi)
	logger := slog.New(handler)

	// 4. 💡 Optional: Make the Go log package use the structured logger.
	slog.SetDefault(logger)

	// 5. Log ⚡
	logger.Info("This is awesome!", "mood", "hyped")
	logger.With("mood", "worried").Warn("This is not that awesome...")
	logger.Error("This is rather bad.", slog.String("mood", "depressed"))
}
