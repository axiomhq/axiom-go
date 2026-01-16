package config

import (
	"net/url"
	"os"
	"strings"
)

// Config is the configuration for Axiom related functionality. It should never
// be created manually but always via the [Default] function.
type Config struct {
	// baseURL of the Axiom instance. Defaults to [CloudURL].
	baseURL *url.URL
	// token is the authentication token that will be set as 'Bearer' on the
	// 'Authorization' header. It must be an api or a personal token.
	token string
	// organizationID is the Axiom organization ID that will be set on the
	// 'X-Axiom-Org-Id' header. Not required for API tokens.
	organizationID string

	// edgeURL is an explicit edge endpoint URL for ingest and query operations.
	// Takes precedence over edgeRegion if both are set.
	edgeURL *url.URL
	// edgeRegion is the regional edge domain (e.g., "eu-central-1.aws.edge.axiom.co").
	// When set, edge URLs are built as "https://{edgeRegion}/v1/ingest/{dataset}".
	edgeRegion string
}

// Default returns a default configuration with the base URL set.
func Default() Config {
	return Config{
		baseURL: apiURL,
	}
}

// BaseURL returns the base URL.
func (c Config) BaseURL() *url.URL {
	return c.baseURL
}

// Token returns the token.
func (c Config) Token() string {
	return c.token
}

// OrganizationID returns the organization ID.
func (c Config) OrganizationID() string {
	return c.organizationID
}

// SetBaseURL sets the base URL.
func (c *Config) SetBaseURL(baseURL *url.URL) {
	c.baseURL = baseURL
}

// SetToken sets the token.
func (c *Config) SetToken(token string) {
	c.token = token
}

// SetOrganizationID sets the organization ID.
func (c *Config) SetOrganizationID(organizationID string) {
	c.organizationID = organizationID
}

// EdgeURL returns the edge URL.
func (c Config) EdgeURL() *url.URL {
	return c.edgeURL
}

// EdgeRegion returns the edge region.
func (c Config) EdgeRegion() string {
	return c.edgeRegion
}

// SetEdgeURL sets the edge URL.
func (c *Config) SetEdgeURL(edgeURL *url.URL) {
	c.edgeURL = edgeURL
}

// SetEdgeRegion sets the edge region.
func (c *Config) SetEdgeRegion(edgeRegion string) {
	c.edgeRegion = edgeRegion
}

// IsEdgeConfigured returns true if an edge endpoint is configured.
func (c Config) IsEdgeConfigured() bool {
	return c.edgeURL != nil || c.edgeRegion != ""
}

// EdgeIngestURL returns the URL for edge-based ingestion for the given dataset.
// Returns nil if no edge configuration is set.
//
// URL handling follows this priority:
//   - If edgeURL has a custom path, it is used as-is
//   - If edgeURL has no path (or only "/"), "/v1/datasets/{dataset}/ingest" is appended for backwards compatibility
//   - If edgeRegion is set, builds "https://{region}/v1/ingest/{dataset}"
func (c Config) EdgeIngestURL(dataset string) *url.URL {
	if c.edgeURL != nil {
		u := *c.edgeURL
		path := strings.TrimSuffix(u.Path, "/")

		// If URL has a custom path, use as-is
		if path != "" {
			return &u
		}

		// No path provided - append legacy format for backwards compatibility
		u.Path = "/v1/datasets/" + dataset + "/ingest"
		return &u
	}

	if c.edgeRegion != "" {
		return &url.URL{
			Scheme: "https",
			Host:   c.edgeRegion,
			Path:   "/v1/ingest/" + dataset,
		}
	}

	return nil
}

// EdgeQueryURL returns the URL for edge-based query operations.
// Returns nil if no edge configuration is set.
//
// URL handling follows this priority:
//   - If edgeURL has a custom path, it is used as-is
//   - If edgeURL has no path (or only "/"), "/v1/datasets/_apl" is appended for backwards compatibility
//   - If edgeRegion is set, builds "https://{region}/v1/query/_apl"
func (c Config) EdgeQueryURL() *url.URL {
	if c.edgeURL != nil {
		u := *c.edgeURL
		path := strings.TrimSuffix(u.Path, "/")

		// If URL has a custom path, use as-is
		if path != "" {
			return &u
		}

		// No path provided - append legacy format for backwards compatibility
		u.Path = "/v1/datasets/_apl"
		return &u
	}

	if c.edgeRegion != "" {
		return &url.URL{
			Scheme: "https",
			Host:   c.edgeRegion,
			Path:   "/v1/query/_apl",
		}
	}

	return nil
}

// Options applies options to the configuration.
func (c *Config) Options(options ...Option) error {
	for _, option := range options {
		if option == nil {
			continue
		} else if err := option(c); err != nil {
			return err
		}
	}
	return nil
}

// IncorporateEnvironment loads configuration from environment variables. It
// will reject invalid values.
func (c *Config) IncorporateEnvironment() error {
	var (
		envURL            = os.Getenv("AXIOM_URL")
		envToken          = os.Getenv("AXIOM_TOKEN")
		envOrganizationID = os.Getenv("AXIOM_ORG_ID")
		envEdgeURL        = os.Getenv("AXIOM_EDGE_URL")
		envEdgeRegion     = os.Getenv("AXIOM_EDGE_REGION")

		options   = make([]Option, 0, 5)
		addOption = func(option Option) { options = append(options, option) }
	)

	if envURL != "" {
		addOption(SetURL(envURL))
	}

	if envToken != "" {
		addOption(SetToken(envToken))
	}

	if envOrganizationID != "" {
		addOption(SetOrganizationID(envOrganizationID))
	}

	if envEdgeURL != "" {
		addOption(SetEdgeURL(envEdgeURL))
	}

	if envEdgeRegion != "" {
		addOption(SetEdgeRegion(envEdgeRegion))
	}

	return c.Options(options...)
}

// Validate the configuration.
func (c Config) Validate() error {
	// Failsafe to protect against an empty baseURL.
	if c.baseURL == nil {
		c.baseURL = apiURL
	}

	if c.token == "" {
		return ErrMissingToken
	} else if !IsValidToken(c.token) {
		return ErrInvalidToken
	}

	// The organization ID is not required for API tokens.
	if c.organizationID == "" && IsPersonalToken(c.token) && c.baseURL.String() == apiURL.String() {
		return ErrMissingOrganizationID
	}

	return nil
}
