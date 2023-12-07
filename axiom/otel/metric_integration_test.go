//go:build integration

package otel_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"

	"github.com/axiomhq/axiom-go/axiom"
	axiotel "github.com/axiomhq/axiom-go/axiom/otel"
)

func TestMetricsIntegration(t *testing.T) {
	ctx := context.Background()

	datasetSuffix := os.Getenv("AXIOM_DATASET_SUFFIX")
	if datasetSuffix == "" {
		datasetSuffix = "local"
	}
	dataset := fmt.Sprintf("test-axiom-go-otel-metric-%s", datasetSuffix)

	client, err := axiom.NewClient()
	require.NoError(t, err)

	_, err = client.Datasets.Create(ctx, axiom.DatasetCreateRequest{
		Name:        dataset,
		Description: "This is a test dataset for otel metric integration tests.",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		err = client.Datasets.Delete(ctx, dataset)
		require.NoError(t, err)
	})

	stop, err := axiotel.InitMetrics(ctx, dataset, "axiom-go-otel-test-metric", "v1.0.0")
	require.NoError(t, err)
	require.NotNil(t, stop)

	t.Cleanup(func() { require.NoError(t, stop()) })

	meter := otel.Meter("main")

	counter, err := meter.Int64Counter("test")
	require.NoError(t, err)

	counter.Add(ctx, 1)
}
