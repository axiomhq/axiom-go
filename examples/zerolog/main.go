// The purpose of this example is to show how to integrate with zerolog.
package main

import (
	"io"
	"log"
	"os"

	adapter "github.com/axiomhq/axiom-go/adapters/zerolog"
	"github.com/rs/zerolog"
	l "github.com/rs/zerolog/log"
)

func main() {
	// Export "AXIOM_DATASET" in addition to the required environment variables.

	writer, err := adapter.New()
	if err != nil {
		log.Fatal(err)
	}

	l.Logger = zerolog.New(io.MultiWriter(writer, os.Stderr)).With().Logger()

	l.Logger.Info().Str("mood", "hyped").Msg("This is awesome!")
	l.Logger.Warn().Str("mood", "worried").Msg("This is no that awesome...")
	l.Logger.Error().Str("mood", "depressed").Msg("This is rather bad.")
}
