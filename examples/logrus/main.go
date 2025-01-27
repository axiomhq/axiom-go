// The purpose of this example is to show how to integrate with logrus.
package main

import (
	"log"

	"github.com/sirupsen/logrus"

	adapter "github.com/axiomhq/axiom-go/adapters/logrus"
)

func main() {
	// Export "AXIOM_DATASET" in addition to the required environment variables.

	// 1. Setup the Axiom hook for logrus.
	hook, err := adapter.New()
	if err != nil {
		log.Fatal(err)
	}

	// 2. Register an exit handler to have all logs flushed before the
	// application exits in case of a "fatal" log operation.
	logrus.RegisterExitHandler(hook.Close)

	// 3. This makes sure logrus calls the registered exit handler. Alternaively
	// hook.Close() can be called manually. It is safe to call multiple times.
	//
	// ❗THIS IS IMPORTANT❗ Without it, the logs will not be sent to Axiom as
	// the buffer will not be flushed when the application exits.
	defer logrus.Exit(0)

	// 4. Spawn the logger.
	logger := logrus.New()

	// 5. Attach the Axiom hook.
	logger.AddHook(hook)

	// 6. Log ⚡
	logger.WithField("mood", "hyped").Info("This is awesome!")
	logger.WithField("mood", "worried").Warn("This is not that awesome...")
	logger.WithField("mood", "depressed").Error("This is rather bad.")
}
