//go:build integration

package slog_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"log/slog"

	adapter "github.com/axiomhq/axiom-go/adapters/slog"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/internal/test/adapters"
)

func Test(t *testing.T) {
	adapters.IntegrationTest(t, "slog", func(_ context.Context, dataset string, client *axiom.Client) {
		handler, err := adapter.New(
			adapter.SetClient(client),
			adapter.SetDataset(dataset),
		)
		require.NoError(t, err)

		defer handler.Close()

		logger := slog.New(handler)

		logger.Info("This is awesome!", slog.String("mood", "hyped"))
		logger.Warn("This is no that awesome...", slog.String("mood", "worried"))
		logger.Error("This is rather bad.", slog.String("mood", "depressed"))
	})
}
