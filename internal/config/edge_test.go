package config

import (
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_EdgeIngestURL_WithEdgeURL_NoPath(t *testing.T) {
	edgeURL, err := url.Parse("https://api.eu.axiom.co")
	require.NoError(t, err)

	cfg := Config{
		edgeURL: edgeURL,
	}

	// No path provided - should append legacy format for backwards compatibility
	result := cfg.EdgeIngestURL("test-dataset")
	require.NotNil(t, result)
	assert.Equal(t, "https://api.eu.axiom.co/v1/datasets/test-dataset/ingest", result.String())
}

func TestConfig_EdgeIngestURL_WithEdgeURL_WithPath(t *testing.T) {
	edgeURL, err := url.Parse("http://localhost:3400/ingest")
	require.NoError(t, err)

	cfg := Config{
		edgeURL: edgeURL,
	}

	// URL has a custom path - should use as-is
	result := cfg.EdgeIngestURL("test-dataset")
	require.NotNil(t, result)
	assert.Equal(t, "http://localhost:3400/ingest", result.String())
}

func TestConfig_EdgeIngestURL_WithEdgeURL_TrailingSlash(t *testing.T) {
	edgeURL, err := url.Parse("https://api.eu.axiom.co/")
	require.NoError(t, err)

	cfg := Config{
		edgeURL: edgeURL,
	}

	// Trailing slash only - should append legacy format
	result := cfg.EdgeIngestURL("test-dataset")
	require.NotNil(t, result)
	assert.Equal(t, "https://api.eu.axiom.co/v1/datasets/test-dataset/ingest", result.String())
}

func TestConfig_EdgeIngestURL_WithEdgeRegion(t *testing.T) {
	cfg := Config{
		edgeRegion: "eu-central-1.aws.edge.axiom.co",
	}

	result := cfg.EdgeIngestURL("my-dataset")
	require.NotNil(t, result)
	assert.Equal(t, "https://eu-central-1.aws.edge.axiom.co/v1/ingest/my-dataset", result.String())
}

func TestConfig_EdgeIngestURL_EdgeURLTakesPrecedence(t *testing.T) {
	edgeURL, err := url.Parse("https://custom-edge.example.com/custom/path")
	require.NoError(t, err)

	cfg := Config{
		edgeURL:    edgeURL,
		edgeRegion: "eu-central-1.aws.edge.axiom.co",
	}

	// edgeURL takes precedence over edgeRegion, and custom path is used as-is
	result := cfg.EdgeIngestURL("test-dataset")
	require.NotNil(t, result)
	assert.Equal(t, "https://custom-edge.example.com/custom/path", result.String())
}

func TestConfig_EdgeIngestURL_NoEdgeConfigured(t *testing.T) {
	cfg := Config{}

	result := cfg.EdgeIngestURL("test-dataset")
	assert.Nil(t, result)
}

func TestConfig_EdgeQueryURL_WithEdgeURL_NoPath(t *testing.T) {
	edgeURL, err := url.Parse("https://api.eu.axiom.co")
	require.NoError(t, err)

	cfg := Config{
		edgeURL: edgeURL,
	}

	// No path provided - should append legacy format for backwards compatibility
	result := cfg.EdgeQueryURL()
	require.NotNil(t, result)
	assert.Equal(t, "https://api.eu.axiom.co/v1/datasets/_apl", result.String())
}

func TestConfig_EdgeQueryURL_WithEdgeURL_WithPath(t *testing.T) {
	edgeURL, err := url.Parse("http://localhost:3400/query")
	require.NoError(t, err)

	cfg := Config{
		edgeURL: edgeURL,
	}

	// URL has a custom path - should use as-is
	result := cfg.EdgeQueryURL()
	require.NotNil(t, result)
	assert.Equal(t, "http://localhost:3400/query", result.String())
}

func TestConfig_EdgeQueryURL_WithEdgeRegion(t *testing.T) {
	cfg := Config{
		edgeRegion: "mumbai.axiom.co",
	}

	result := cfg.EdgeQueryURL()
	require.NotNil(t, result)
	assert.Equal(t, "https://mumbai.axiom.co/v1/query/_apl", result.String())
}

func TestConfig_EdgeQueryURL_NoEdgeConfigured(t *testing.T) {
	cfg := Config{}

	result := cfg.EdgeQueryURL()
	assert.Nil(t, result)
}

func TestConfig_IsEdgeConfigured(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected bool
	}{
		{
			name:     "no edge configured",
			config:   Config{},
			expected: false,
		},
		{
			name: "edge URL configured",
			config: Config{
				edgeURL: &url.URL{Scheme: "https", Host: "edge.example.com"},
			},
			expected: true,
		},
		{
			name: "edge region configured",
			config: Config{
				edgeRegion: "eu-central-1.aws.edge.axiom.co",
			},
			expected: true,
		},
		{
			name: "both configured",
			config: Config{
				edgeURL:    &url.URL{Scheme: "https", Host: "edge.example.com"},
				edgeRegion: "eu-central-1.aws.edge.axiom.co",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.IsEdgeConfigured())
		})
	}
}

func TestSetEdgeURL(t *testing.T) {
	cfg := Default()
	err := cfg.Options(SetEdgeURL("https://edge.example.com"))
	require.NoError(t, err)

	assert.NotNil(t, cfg.EdgeURL())
	assert.Equal(t, "https://edge.example.com", cfg.EdgeURL().String())
}

func TestSetEdgeURL_Invalid(t *testing.T) {
	cfg := Default()
	err := cfg.Options(SetEdgeURL("not a valid url"))
	assert.Error(t, err)
}

func TestSetEdgeRegion(t *testing.T) {
	cfg := Default()
	err := cfg.Options(SetEdgeRegion("eu-central-1.aws.edge.axiom.co"))
	require.NoError(t, err)

	assert.Equal(t, "eu-central-1.aws.edge.axiom.co", cfg.EdgeRegion())
}

func TestIncorporateEnvironment_EdgeVariables(t *testing.T) {
	t.Run("AXIOM_EDGE_URL", func(t *testing.T) {
		t.Setenv("AXIOM_EDGE_URL", "https://edge.example.com")

		cfg := Default()
		err := cfg.IncorporateEnvironment()
		require.NoError(t, err)

		assert.NotNil(t, cfg.EdgeURL())
		assert.Equal(t, "https://edge.example.com", cfg.EdgeURL().String())
	})

	t.Run("AXIOM_EDGE_REGION", func(t *testing.T) {
		t.Setenv("AXIOM_EDGE_REGION", "mumbai.axiom.co")

		cfg := Default()
		err := cfg.IncorporateEnvironment()
		require.NoError(t, err)

		assert.Equal(t, "mumbai.axiom.co", cfg.EdgeRegion())
	})

	t.Run("both edge variables", func(t *testing.T) {
		t.Setenv("AXIOM_EDGE_URL", "https://edge.example.com")
		t.Setenv("AXIOM_EDGE_REGION", "mumbai.axiom.co")

		cfg := Default()
		err := cfg.IncorporateEnvironment()
		require.NoError(t, err)

		// Both should be set, URL takes precedence in EdgeIngestURL
		assert.NotNil(t, cfg.EdgeURL())
		assert.Equal(t, "mumbai.axiom.co", cfg.EdgeRegion())
	})
}

func TestEdgeRegionalFormats(t *testing.T) {
	tests := []struct {
		name           string
		region         string
		dataset        string
		expectedIngest string
		expectedQuery  string
	}{
		{
			name:           "AWS production edge",
			region:         "eu-central-1.aws.edge.axiom.co",
			dataset:        "my-dataset",
			expectedIngest: "https://eu-central-1.aws.edge.axiom.co/v1/ingest/my-dataset",
			expectedQuery:  "https://eu-central-1.aws.edge.axiom.co/v1/query/_apl",
		},
		{
			name:           "staging edge",
			region:         "us-east-1.edge.staging.axiomdomain.co",
			dataset:        "test-dataset",
			expectedIngest: "https://us-east-1.edge.staging.axiomdomain.co/v1/ingest/test-dataset",
			expectedQuery:  "https://us-east-1.edge.staging.axiomdomain.co/v1/query/_apl",
		},
		{
			name:           "dev edge",
			region:         "eu-west-1.edge.dev.axiomdomain.co",
			dataset:        "dev-dataset",
			expectedIngest: "https://eu-west-1.edge.dev.axiomdomain.co/v1/ingest/dev-dataset",
			expectedQuery:  "https://eu-west-1.edge.dev.axiomdomain.co/v1/query/_apl",
		},
		{
			name:           "simple regional domain",
			region:         "mumbai.axiom.co",
			dataset:        "logs",
			expectedIngest: "https://mumbai.axiom.co/v1/ingest/logs",
			expectedQuery:  "https://mumbai.axiom.co/v1/query/_apl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				edgeRegion: tt.region,
			}

			ingestURL := cfg.EdgeIngestURL(tt.dataset)
			require.NotNil(t, ingestURL)
			assert.Equal(t, tt.expectedIngest, ingestURL.String())

			queryURL := cfg.EdgeQueryURL()
			require.NotNil(t, queryURL)
			assert.Equal(t, tt.expectedQuery, queryURL.String())
		})
	}
}

func TestMain(m *testing.M) {
	// Clear edge-related env vars before running tests
	os.Unsetenv("AXIOM_EDGE_URL")
	os.Unsetenv("AXIOM_EDGE_REGION")
	os.Exit(m.Run())
}
