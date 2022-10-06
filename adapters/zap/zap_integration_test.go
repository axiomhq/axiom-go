//go:build integration

package zap_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	adapter "github.com/axiomhq/axiom-go/adapters/zap"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/internal/test/adapters"
)

func Test(t *testing.T) {
	adapters.IntegrationTest(t, "zap", func(_ context.Context, dataset string, client *axiom.Client) {
		core, err := adapter.New(
			adapter.SetClient(client),
			adapter.SetDataset(dataset),
		)
		require.NoError(t, err)

		logger := zap.New(core)
		defer func() {
			err := logger.Sync()
			assert.NoError(t, err)
		}()

		logger.Info("This is awesome!", zap.String("mood", "hyped"))
		logger.Warn("This is no that awesome...", zap.String("mood", "worried"))
		logger.Error("This is rather bad.", zap.String("mood", "depressed"))
	})
}
