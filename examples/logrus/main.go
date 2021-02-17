// The purpose of this example is to show how to integrate with logrus.
package main

import (
	"log"
	"os"

	"github.com/sirupsen/logrus"

	adapter "github.com/axiomhq/axiom-go/adapters/logrus"
)

func main() {
	var (
		deploymentURL = os.Getenv("AXM_DEPLOYMENT_URL")
		accessToken   = os.Getenv("AXM_ACCESS_TOKEN")
		dataset       = os.Getenv("AXM_DATASET")
	)

	// 1. Setup the Axiom hook for logrus.
	hook, err := adapter.New(deploymentURL, accessToken, dataset)
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

	// 6. This makes sure logrus calls the registered exit handler. Alternaively
	// hook.Close() can be called manually. It is safe to call multiple times.
	logrus.Exit(0)
}
