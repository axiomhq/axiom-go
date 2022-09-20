package config

import (
	"net/url"
	"os"
)

// Config is the configuration for Axiom related functionality. It should never
// be created manually but always via the `Default` function.
type Config struct {
	// baseURL of the Axiom instance. Defaults to `CloudURL`.
	baseURL *url.URL
	// accessToken is the authentication token that will be set as 'Bearer' on
	// the 'Authorization' header. It must be a personal or an API token.
	accessToken string
	// organizationID is the Axiom organization ID that will be set on the
	// 'X-Axiom-Org-Id' header. Not required for API tokens or Axiom Selfhost.
	organizationID string
}

// Default returns a default configuration with the base URL set.
func Default() Config {
	return Config{
		baseURL: cloudURL,
	}
}

// BaseURL returns the base URL.
func (c Config) BaseURL() *url.URL {
	return c.baseURL
}

// AccessToken returns the access token.
func (c Config) AccessToken() string {
	return c.accessToken
}

// OrganizationID returns the organization ID.
func (c Config) OrganizationID() string {
	return c.organizationID
}

// SetBaseURL sets the base URL.
func (c *Config) SetBaseURL(baseURL *url.URL) {
	c.baseURL = baseURL
}

// SetAccessToken sets the access token.
func (c *Config) SetAccessToken(accessToken string) {
	c.accessToken = accessToken
}

// SetOrganizationID sets the organization ID.
func (c *Config) SetOrganizationID(organizationID string) {
	c.organizationID = organizationID
}

// Options applies options to the configuration.
func (c *Config) Options(options ...Option) error {
	for _, option := range options {
		if err := option(c); err != nil {
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
		envAccessToken    = os.Getenv("AXIOM_TOKEN")
		envOrganizationID = os.Getenv("AXIOM_ORG_ID")

		options   = make([]Option, 0, 3)
		addOption = func(option Option) { options = append(options, option) }
	)

	if envURL != "" {
		addOption(SetURL(envURL))
	}

	if envAccessToken != "" {
		addOption(SetAccessToken(envAccessToken))
	}

	if envOrganizationID != "" {
		addOption(SetOrganizationID(envOrganizationID))
	}

	return c.Options(options...)
}

// Validate the configuration.
func (c Config) Validate() error {
	// Failsafe to protect against an empty baseURL.
	if c.baseURL == nil {
		c.baseURL = cloudURL
	}

	if c.accessToken == "" {
		return ErrMissingAccessToken
	} else if !IsValidToken(c.accessToken) {
		return ErrInvalidToken
	}

	// The organization ID is not required for API tokens or Axiom Selfhost.
	if c.organizationID == "" && IsPersonalToken(c.accessToken) && c.baseURL.String() == cloudURL.String() {
		return ErrMissingOrganizationID
	}

	return nil
}
