package config

import (
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/internal/test/testhelper"
)

const (
	endpoint         = "http://axiom.local"
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
				baseURL:        cloudURL,
				accessToken:    personalToken,
				organizationID: organizationID,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.SafeClearEnv(t)

			for k, v := range tt.environment {
				os.Setenv(k, v)
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
			expErr: ErrMissingAccessToken,
		},
		{
			name: "missing organization id",
			config: Config{
				accessToken: personalToken,
			},
			expErr: ErrMissingOrganizationID,
		},
		{
			name: "missing nothing",
			config: Config{
				accessToken:    personalToken,
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
