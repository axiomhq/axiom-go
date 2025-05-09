package apex_test

import (
	"context"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/require"

	adapter "github.com/axiomhq/axiom-go/adapters/apex"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/internal/test/adapters"
)

func Test(t *testing.T) {
	adapters.IntegrationTest(t, "apex", func(_ context.Context, dataset string, client *axiom.Client) {
		handler, err := adapter.New(
			adapter.SetClient(client),
			adapter.SetDataset(dataset),
		)
		require.NoError(t, err)

		defer handler.Close()

		log.SetHandler(handler)

		log.WithField("mood", "hyped").Info("This is awesome!")
		log.WithField("mood", "worried").Warn("This is not that awesome...")
		log.WithField("mood", "depressed").Error("This is rather bad.")
	})
}
