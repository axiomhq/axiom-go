//go:build integration

package apex_test

import (
	"context"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/adapters"
	adapter "github.com/axiomhq/axiom-go/adapters/apex"
	"github.com/axiomhq/axiom-go/axiom"
)

func Test(t *testing.T) {
	adapters.TestAdapter(t, "apex", func(_ context.Context, dataset string, client *axiom.Client) {
		handler, err := adapter.New(
			adapter.SetClient(client),
			adapter.SetDataset(dataset),
		)
		require.NoError(t, err)

		defer handler.Close()

		log.SetHandler(handler)

		log.WithField("mood", "hyped").Info("This is awesome!")
		log.WithField("mood", "worried").Warn("This is no that awesome...")
		log.WithField("mood", "depressed").Error("This is rather bad.")
	})
}
