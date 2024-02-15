package otel_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/axiomhq/axiom-go/axiom"
	axiotel "github.com/axiomhq/axiom-go/axiom/otel"
	"github.com/axiomhq/axiom-go/internal/config"
	"github.com/axiomhq/axiom-go/internal/test/testhelper"
)

func TestTracingIntegration(t *testing.T) {
	cfg := config.Default()
	if err := cfg.IncorporateEnvironment(); err != nil {
		t.Fatal(err)
	}

	if cfg.Token() == "" || cfg.OrganizationID() == "" || cfg.BaseURL() == nil {
		t.Skip("missing required environment variables to run integration tests")
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("invalid configuration: %s", err)
	}

	datasetSuffix := os.Getenv("AXIOM_DATASET_SUFFIX")
	if datasetSuffix == "" {
		datasetSuffix = "local"
	}

	// Clear the environment to avoid unexpected behavior.
	testhelper.SafeClearEnv(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	t.Cleanup(cancel)

	userAgent := fmt.Sprintf("axiom-go-otel-integration-test/%s", datasetSuffix)
	client, err := axiom.NewClient(
		axiom.SetNoEnv(),
		axiom.SetURL(cfg.BaseURL().String()),
		axiom.SetToken(cfg.Token()),
		axiom.SetOrganizationID(cfg.OrganizationID()),
		axiom.SetUserAgent(userAgent),
	)
	require.NoError(t, err)

	// Get some info on the user that runs the test.
	testUser, err := client.Users.Current(ctx)
	require.NoError(t, err)

	t.Logf("using account %q", testUser.Name)

	// Create the dataset to use...
	dataset, err := client.Datasets.Create(ctx, axiom.DatasetCreateRequest{
		Name:        fmt.Sprintf("test-axiom-go-otel-%s", datasetSuffix),
		Description: "This is a test dataset for datasets integration tests.",
	})
	require.NoError(t, err)

	// ... and make sure it's deleted after the test.
	t.Cleanup(func() {
		teardownCtx := teardownContext(t, time.Second*15)
		deleteErr := client.Datasets.Delete(teardownCtx, dataset.ID)
		assert.NoError(t, deleteErr)
	})

	stop, err := axiotel.InitTracing(ctx, dataset.ID, "axiom-go-otel-test", "v1.0.0",
		axiotel.SetNoEnv(),
		axiotel.SetURL(cfg.BaseURL().String()),
		axiotel.SetToken(cfg.Token()),
		axiotel.SetOrganizationID(cfg.OrganizationID()),
	)
	require.NoError(t, err)
	require.NotNil(t, stop)

	t.Cleanup(func() {
		assert.NoError(t, stop())
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

func teardownContext(t *testing.T, timeout time.Duration) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)
	return ctx
}
