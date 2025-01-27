package apex_test

import (
	"github.com/apex/log"

	adapter "github.com/axiomhq/axiom-go/adapters/apex"
)

func Example() {
	// Export "AXIOM_DATASET" in addition to the required environment variables.

	handler, err := adapter.New()
	if err != nil {
		log.Fatal(err.Error())
	}
	defer handler.Close()

	log.SetHandler(handler)

	log.WithField("mood", "hyped").Info("This is awesome!")
	log.WithField("mood", "worried").Warn("This is not that awesome...")
	log.WithField("mood", "depressed").Error("This is rather bad.")
}
