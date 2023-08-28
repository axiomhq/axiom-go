package config

import (
	"net/url"
	"os"
)

// Config is the configuration for Axiom related functionality. It should never
// be created manually but always via the [Default] function.
type Config struct {
	// baseURL of the Axiom instance. Defaults to [APIURL].
	baseURL *url.URL
	// token is the authentication token that will be set as 'Bearer' on the
	// 'Authorization' header. It must be an api or a personal token.
	token string
	// organizationID is the Axiom organization ID that will be set on the
	// 'X-Axiom-Org-Id' header. Not required for API tokens.
	organizationID string
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

		options   = make([]Option, 0, 3)
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
