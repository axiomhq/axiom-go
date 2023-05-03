//go:build integration

package otel_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/axiomhq/axiom-go/axiom"
	axiotel "github.com/axiomhq/axiom-go/axiom/otel"
)

func TestTracingIntegration(t *testing.T) {
	ctx := context.Background()

	datasetSuffix := os.Getenv("AXIOM_DATASET_SUFFIX")
	if datasetSuffix == "" {
		datasetSuffix = "local"
	}
	dataset := fmt.Sprintf("test-axiom-go-otel-%s", datasetSuffix)

	client, err := axiom.NewClient()
	require.NoError(t, err)

	_, err = client.Datasets.Create(ctx, axiom.DatasetCreateRequest{
		Name:        dataset,
		Description: "This is a test dataset for datasets integration tests.",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		err = client.Datasets.Delete(ctx, dataset)
		require.NoError(t, err)
	})

	stop, err := axiotel.InitTracing(ctx, dataset, "axiom-go-otel-test", "v1.0.0")
	require.NoError(t, err)
	require.NotNil(t, stop)

	t.Cleanup(func() {
		require.NoError(t, stop())
	})

	bar := func(ctx context.Context) {
		tr := otel.Tracer("bar")
		_, span := tr.Start(ctx, "bar")
		span.SetAttributes(attribute.Key("testset").String("value"))
		defer span.End()

		time.Sleep(time.Millisecond * 100)
	}

	tr := otel.Tracer("main")

	ctx, span := tr.Start(ctx, "foo")
	defer span.End()

	bar(ctx)
}
