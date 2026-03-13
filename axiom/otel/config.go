package otel

import (
	"time"

	"github.com/axiomhq/axiom-go/internal/config"
)

type exporterConfig struct {
	config.Config

	// APIEndpoint is the endpoint to use for the exporter.
	APIEndpoint string
	// Timeout is the timeout for the exporters underlying [http.Client].
	Timeout time.Duration
	// NoEnv disables the use of "AXIOM_*" environment variables.
	NoEnv bool
}

func defaultExporterConfig(apiEndpoint string) exporterConfig {
	return exporterConfig{
		Config:      config.Default(),
		APIEndpoint: apiEndpoint,
	}
}

// An Option modifies the behaviour of OpenTelemetry exporters. Nonetheless,
// the official "OTEL_*" environment variables are preferred over the options or
// "AXIOM_*" environment variables.
type Option func(c *exporterConfig) error

// SetURL sets the base URL used by the client.
//
// Can also be specified using the "AXIOM_URL" environment variable.
func SetURL(baseURL string) Option {
	return func(c *exporterConfig) error { return c.Options(config.SetURL(baseURL)) }
}

// SetToken specifies the authentication token used by the client.
//
// Can also be specified using the "AXIOM_TOKEN" environment variable.
func SetToken(token string) Option {
	return func(c *exporterConfig) error { return c.Options(config.SetToken(token)) }
}

// SetOrganizationID specifies the organization ID used by the client.
//
// Can also be specified using the "AXIOM_ORG_ID" environment variable.
func SetOrganizationID(organizationID string) Option {
	return func(c *exporterConfig) error { return c.Options(config.SetOrganizationID(organizationID)) }
}

// SetAPIEndpoint specifies the api endpoint used by the client.
func SetAPIEndpoint(path string) Option {
	return func(c *exporterConfig) error {
		c.APIEndpoint = path
		return nil
	}
}

// SetTimeout specifies the http timeout used by the client.
func SetTimeout(timeout time.Duration) Option {
	return func(c *exporterConfig) error {
		c.Timeout = timeout
		return nil
	}
}

// SetNoEnv prevents the client from deriving its configuration from the
// environment (by auto reading "AXIOM_*" environment variables).
func SetNoEnv() Option {
	return func(c *exporterConfig) error {
		c.NoEnv = true
		return nil
	}
}
