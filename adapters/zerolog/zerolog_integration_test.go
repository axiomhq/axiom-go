//go:build integration

package zerolog_test

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	adapter "github.com/axiomhq/axiom-go/adapters/zerolog"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/internal/test/adapters"
)

func TestRealDataset(t *testing.T) {
	adapters.IntegrationTest(t, "zerolog", func(_ context.Context, dataset string, client *axiom.Client) {
		zerolog.TimeFieldFormat = time.RFC3339Nano
		ws, err := adapter.New(
			adapter.SetClient(client),
			adapter.SetDataset(dataset),
		)
		require.NoError(t, err)

		logger := zerolog.New(io.MultiWriter(ws, os.Stderr)).With().Timestamp().Logger()
		defer ws.Close()

		// test can log after closing the adapter
		logger.Info().Str("mood", "hyped").Msg("my seen message")
		logger.Info().Str("mood", "worried").Msg("my seen message 2")
		logger.Info().Str("mood", "depressed").Msg("my seen message 3")
	})
}

func TestCanHandleNoTime(t *testing.T) {
	adapters.IntegrationTest(t, "zerolog", func(_ context.Context, dataset string, client *axiom.Client) {
		ws, err := adapter.New(
			adapter.SetClient(client),
			adapter.SetDataset(dataset),
		)
		require.NoError(t, err)

		logger := zerolog.New(io.MultiWriter(ws, os.Stderr))
		defer ws.Close()

		// test can log after closing the adapter
		logger.Info().Str("mood", "hyped").Msg("my seen message")
		logger.Info().Str("mood", "worried").Msg("my seen message 2")
		logger.Info().Str("mood", "depressed").Msg("my seen message 3")
	})
}
