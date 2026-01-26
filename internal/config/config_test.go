package config

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/internal/test/testhelper"
)

const (
	endpoint         = "http://api.axiom.local"
	apiToken         = "xaat-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
	personalToken    = "xapt-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX" //nolint:gosec // Chill, it's just testing.
	unspecifiedToken = "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
	organizationID   = "awkward-identifier-c3po"
)

func TestConfig_IncorporateEnvironment(t *testing.T) {
	tests := []struct {
		name        string
		baseConfig  Config
		environment map[string]string
		want        Config
		expErr      error
	}{
		{
			name: "no environment, no preset",
		},
		{
			name: "url environment, no preset",
			environment: map[string]string{
				"AXIOM_URL": endpoint,
			},
			want: Config{
				baseURL: mustParseURL(t, endpoint),
			},
		},
		{
			name: "url environment; url preset",
			baseConfig: Config{
				baseURL: mustParseURL(t, endpoint),
			},
			environment: map[string]string{
				"AXIOM_URL": "http://some-new-url",
			},
			want: Config{
				baseURL: mustParseURL(t, "http://some-new-url"),
			},
		},
		{
			name:       "token, org id environment; default preset",
			baseConfig: Default(),
			environment: map[string]string{
				"AXIOM_TOKEN":  personalToken,
				"AXIOM_ORG_ID": organizationID,
			},
			want: Config{
				baseURL:        apiURL,
				token:          personalToken,
				organizationID: organizationID,
			},
		},
		{
			name:       "edge url environment; default preset",
			baseConfig: Default(),
			environment: map[string]string{
				"AXIOM_EDGE_URL": "https://edge.example.com",
			},
			want: Config{
				baseURL: apiURL,
				edgeURL: mustParseURL(t, "https://edge.example.com"),
			},
		},
		{
			name:       "edge domain environment; default preset",
			baseConfig: Default(),
			environment: map[string]string{
				"AXIOM_EDGE": "eu-central-1.aws.edge.axiom.co",
			},
			want: Config{
				baseURL: apiURL,
				edge:    "eu-central-1.aws.edge.axiom.co",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.SafeClearEnv(t)

			for k, v := range tt.environment {
				t.Setenv(k, v)
			}

			assert.Equal(t, tt.expErr, tt.baseConfig.IncorporateEnvironment())
			assert.Equal(t, tt.want, tt.baseConfig)
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		expErr error
	}{
		{
			name:   "no nothing",
			expErr: ErrMissingToken,
		},
		{
			name: "missing organization id",
			config: Config{
				token: personalToken,
			},
			expErr: ErrMissingOrganizationID,
		},
		{
			name: "missing nothing",
			config: Config{
				token:          personalToken,
				organizationID: organizationID,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expErr, tt.config.Validate())
		})
	}
}

func mustParseURL(tb testing.TB, urlStr string) *url.URL {
	tb.Helper()

	u, err := url.ParseRequestURI(urlStr)
	require.NoError(tb, err)

	return u
}

func TestConfig_EdgeIngestURL(t *testing.T) {
	tests := []struct {
		name     string
		edgeURL  string
		edge     string
		dataset  string
		expected string
	}{
		{
			name:     "no edge configured",
			dataset:  "test-dataset",
			expected: "",
		},
		{
			name:     "edge URL without path",
			edgeURL:  "https://eu-central-1.aws.edge.axiom.co",
			dataset:  "test-dataset",
			expected: "https://eu-central-1.aws.edge.axiom.co/v1/ingest/test-dataset",
		},
		{
			name:     "edge URL with trailing slash",
			edgeURL:  "https://eu-central-1.aws.edge.axiom.co/",
			dataset:  "test-dataset",
			expected: "https://eu-central-1.aws.edge.axiom.co/v1/ingest/test-dataset",
		},
		{
			name:     "edge URL with custom path",
			edgeURL:  "http://localhost:3400/ingest",
			dataset:  "test-dataset",
			expected: "http://localhost:3400/ingest",
		},
		{
			name:     "edge domain only",
			edge:     "eu-central-1.aws.edge.axiom.co",
			dataset:  "my-dataset",
			expected: "https://eu-central-1.aws.edge.axiom.co/v1/ingest/my-dataset",
		},
		{
			name:     "edge URL takes precedence over edge domain",
			edgeURL:  "https://primary.edge.axiom.co",
			edge:     "secondary.edge.axiom.co",
			dataset:  "test-dataset",
			expected: "https://primary.edge.axiom.co/v1/ingest/test-dataset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{edge: tt.edge}
			if tt.edgeURL != "" {
				cfg.edgeURL = mustParseURL(t, tt.edgeURL)
			}

			result := cfg.EdgeIngestURL(tt.dataset)
			if tt.expected == "" {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected, result.String())
			}
		})
	}
}

func TestConfig_EdgeQueryURL(t *testing.T) {
	tests := []struct {
		name     string
		edgeURL  string
		edge     string
		expected string
	}{
		{
			name:     "no edge configured",
			expected: "",
		},
		{
			name:     "edge URL without path",
			edgeURL:  "https://eu-central-1.aws.edge.axiom.co",
			expected: "https://eu-central-1.aws.edge.axiom.co/v1/query/_apl",
		},
		{
			name:     "edge URL with trailing slash",
			edgeURL:  "https://eu-central-1.aws.edge.axiom.co/",
			expected: "https://eu-central-1.aws.edge.axiom.co/v1/query/_apl",
		},
		{
			name:     "edge URL with custom path",
			edgeURL:  "http://localhost:3400/query",
			expected: "http://localhost:3400/query",
		},
		{
			name:     "edge domain only",
			edge:     "eu-central-1.aws.edge.axiom.co",
			expected: "https://eu-central-1.aws.edge.axiom.co/v1/query/_apl",
		},
		{
			name:     "edge URL takes precedence over edge domain",
			edgeURL:  "https://primary.edge.axiom.co",
			edge:     "secondary.edge.axiom.co",
			expected: "https://primary.edge.axiom.co/v1/query/_apl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{edge: tt.edge}
			if tt.edgeURL != "" {
				cfg.edgeURL = mustParseURL(t, tt.edgeURL)
			}

			result := cfg.EdgeQueryURL()
			if tt.expected == "" {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected, result.String())
			}
		})
	}
}

func TestConfig_IsEdgeConfigured(t *testing.T) {
	tests := []struct {
		name     string
		edgeURL  string
		edge     string
		expected bool
	}{
		{
			name:     "no edge configured",
			expected: false,
		},
		{
			name:     "edge URL configured",
			edgeURL:  "https://edge.example.com",
			expected: true,
		},
		{
			name:     "edge domain configured",
			edge:     "edge.example.com",
			expected: true,
		},
		{
			name:     "both configured",
			edgeURL:  "https://edge.example.com",
			edge:     "edge.example.com",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{edge: tt.edge}
			if tt.edgeURL != "" {
				cfg.edgeURL = mustParseURL(t, tt.edgeURL)
			}
			assert.Equal(t, tt.expected, cfg.IsEdgeConfigured())
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

func TestSetEdge(t *testing.T) {
	cfg := Default()
	err := cfg.Options(SetEdge("eu-central-1.aws.edge.axiom.co"))
	require.NoError(t, err)

	assert.Equal(t, "eu-central-1.aws.edge.axiom.co", cfg.Edge())
}
