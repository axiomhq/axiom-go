package logrus_test

import (
	"log"

	"github.com/sirupsen/logrus"

	adapter "github.com/axiomhq/axiom-go/adapters/logrus"
)

func Example() {
	// Export "AXIOM_DATASET" in addition to the required environment variables.

	hook, err := adapter.New()
	if err != nil {
		log.Fatal(err)
	}
	logrus.RegisterExitHandler(hook.Close)

	logger := logrus.New()
	logger.AddHook(hook)

	logger.WithField("mood", "hyped").Info("This is awesome!")
	logger.WithField("mood", "worried").Warn("This is not that awesome...")
	logger.WithField("mood", "depressed").Error("This is rather bad.")

	logrus.Exit(0)
}
