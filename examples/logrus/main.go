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

	// 3. Spawn the logger.
	logger := logrus.New()

	// 4. Attach the Axiom hook.
	logger.AddHook(hook)

	// 5. Log ⚡
	logger.WithField("mood", "hyped").Info("This is awesome!")
	logger.WithField("mood", "worried").Warn("This is no that awesome...")
	logger.WithField("mood", "depressed").Error("This is rather bad.")

	// 6. This makes sure logrus calls the registered exit handler. Alternaively
	// hook.Close() can be called manually. It is safe to call multiple times.
	//
	// ❗THIS IS IMPORTANT❗ Without it, the logs will not be sent to Axiom as
	// the buffer will not be flushed when the application exits.
	logrus.Exit(0)
}
