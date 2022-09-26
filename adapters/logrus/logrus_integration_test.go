//go:build integration

package logrus_test

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	adapter "github.com/axiomhq/axiom-go/adapters/logrus"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/internal/test/adapters"
)

func Test(t *testing.T) {
	adapters.IntegrationTest(t, "logrus", func(_ context.Context, dataset string, client *axiom.Client) {
		hook, err := adapter.New(
			adapter.SetClient(client),
			adapter.SetDataset(dataset),
		)
		require.NoError(t, err)

		defer hook.Close()

		logger := logrus.New()
		logger.AddHook(hook)

		logger.WithField("mood", "hyped").Info("This is awesome!")
		logger.WithField("mood", "worried").Warn("This is no that awesome...")
		logger.WithField("mood", "depressed").Error("This is rather bad.")
	})
}
