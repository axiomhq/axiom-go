package otel

import (
	"time"

	"github.com/axiomhq/axiom-go/internal/config"
)

const (
	defaultMetricAPIEndpoint = "/v1/metrics"
	defaultTraceAPIEndpoint  = "/v1/traces"
)

// Config is the configuration for OpenTelemetry components initialized by this
// helper package. This type is exported for convenience but an [Option] is
// naturally applied by one or more "Set"-prefixed functions.
type Config struct {
	config.Config

	// APIEndpoint is the endpoint to use for an exporter.
	APIEndpoint string
	// Timeout is the timeout for an exporters underlying [http.Client].
	Timeout time.Duration
	// NoEnv disables the use of "AXIOM_*" environment variables.
	NoEnv bool
}

func defaultMetricConfig() Config {
	return Config{
		Config:      config.Default(),
		APIEndpoint: defaultMetricAPIEndpoint,
	}
}

func defaultTraceConfig() Config {
	return Config{
		Config:      config.Default(),
		APIEndpoint: defaultTraceAPIEndpoint,
	}
}

// An Option modifies the behaviour of OpenTelemetry exporters. Nonetheless,
// the official "OTEL_*" environment variables are preferred over the options or
// "AXIOM_*" environment variables.
type Option func(c *Config) error

// SetURL sets the base URL used by the client.
//
// Can also be specified using the "AXIOM_URL" environment variable.
func SetURL(baseURL string) Option {
	return func(c *Config) error { return c.Options(config.SetURL(baseURL)) }
}

// SetToken specifies the authentication token used by the client.
//
// Can also be specified using the "AXIOM_TOKEN" environment variable.
func SetToken(token string) Option {
	return func(c *Config) error { return c.Options(config.SetToken(token)) }
}

// SetOrganizationID specifies the organization ID used by the client.
//
// Can also be specified using the "AXIOM_ORG_ID" environment variable.
func SetOrganizationID(organizationID string) Option {
	return func(c *Config) error { return c.Options(config.SetOrganizationID(organizationID)) }
}

// SetAPIEndpoint specifies the api endpoint used by the client.
func SetAPIEndpoint(path string) Option {
	return func(c *Config) error {
		c.APIEndpoint = path
		return nil
	}
}

// SetTimeout specifies the http timeout used by the client.
func SetTimeout(timeout time.Duration) Option {
	return func(c *Config) error {
		c.Timeout = timeout
		return nil
	}
}

// SetNoEnv prevents the client from deriving its configuration from the
// environment (by auto reading "AXIOM_*" environment variables).
func SetNoEnv() Option {
	return func(c *Config) error {
		c.NoEnv = true
		return nil
	}
}

func populateAndValidateConfig(base *Config, options ...Option) error {
	// Apply supplied options.
	for _, option := range options {
		if option == nil {
			continue
		} else if err := option(base); err != nil {
			return err
		}
	}

	// Make sure to populate remaining fields from the environment, if not
	// explicitly disabled.
	if !base.NoEnv {
		if err := base.IncorporateEnvironment(); err != nil {
			return err
		}
	}

	return base.Validate()
}
