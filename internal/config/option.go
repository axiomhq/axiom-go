package config

import (
	"net/url"
	"strings"
)

// An Option modifies the configuration.
type Option func(config *Config) error

// SetURL specifies the base URL to use.
func SetURL(baseURL string) Option {
	return func(config *Config) (err error) {
		baseURL, err := url.ParseRequestURI(baseURL)
		if err != nil {
			return err
		}

		config.SetBaseURL(baseURL)

		return nil
	}
}

// SetToken specifies the token to use.
func SetToken(token string) Option {
	return func(config *Config) error {
		if !IsValidToken(token) {
			return ErrInvalidToken
		}

		config.SetToken(token)

		return nil
	}
}

// SetOrganizationID specifies the organization ID to use.
func SetOrganizationID(organizationID string) Option {
	return func(config *Config) error {
		config.SetOrganizationID(organizationID)
		return nil
	}
}

// SetEdgeURL specifies the edge URL to use for ingest and query operations.
// The URL should include the scheme (e.g., "https://eu-central-1.aws.edge.axiom.co").
// This takes precedence over [SetEdge] if both are set.
func SetEdgeURL(edgeURL string) Option {
	return func(config *Config) (err error) {
		parsedURL, err := url.ParseRequestURI(edgeURL)
		if err != nil {
			return err
		}

		config.SetEdgeURL(parsedURL)

		return nil
	}
}

// SetEdge specifies the regional edge domain to use for ingest and query
// operations. Specify the domain only (e.g., "eu-central-1.aws.edge.axiom.co").
// When set, edge URLs are built as "https://{edge}/v1/ingest/{dataset}".
func SetEdge(edge string) Option {
	return func(config *Config) error {
		config.SetEdge(edge)
		return nil
	}
}

// SetOtelEnabled enables or disables OpenTelemetry-based ingestion via the
// /v1/logs endpoint. The value should be "true" or "1" to enable.
func SetOtelEnabled(enabled string) Option {
	return func(config *Config) error {
		enabled = strings.ToLower(strings.TrimSpace(enabled))
		config.SetOtelEnabled(enabled == "true" || enabled == "1")
		return nil
	}
}
