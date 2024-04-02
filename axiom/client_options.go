package axiom

import (
	"net/http"

	"go.opentelemetry.io/otel/trace/noop"

	"github.com/axiomhq/axiom-go/internal/config"
)

// An Option modifies the behaviour of the client. If not otherwise specified
// by a specific option, they are safe to use even after methods have been
// called. However, they are not safe to use while the client is performing an
// operation.
type Option func(c *Client) error

// SetURL specifies the base URL used by the [Client].
//
// Can also be specified using the "AXIOM_URL" environment variable.
func SetURL(baseURL string) Option {
	return func(c *Client) error { return c.config.Options(config.SetURL(baseURL)) }
}

// SetToken specifies the token used by the [Client].
//
// Can also be specified using the "AXIOM_TOKEN" environment variable.
func SetToken(accessToken string) Option {
	return func(c *Client) error { return c.config.Options(config.SetToken(accessToken)) }
}

// SetOrganizationID specifies the organization ID used by the [Client].
//
// When a personal token is used, this method can be used to switch between
// organizations by passing it to the [Client.Options] method.
//
// Can also be specified using the "AXIOM_ORG_ID" environment variable.
func SetOrganizationID(organizationID string) Option {
	return func(c *Client) error { return c.config.Options(config.SetOrganizationID(organizationID)) }
}

// SetPersonalTokenConfig specifies all properties needed in order to
// successfully connect to Axiom with a personal token.
func SetPersonalTokenConfig(personalToken, organizationID string) Option {
	return func(c *Client) error {
		return c.Options(
			SetToken(personalToken),
			SetOrganizationID(organizationID),
		)
	}
}

// SetAPITokenConfig specifies all properties needed in order to successfully
// connect to Axiom with an API token.
func SetAPITokenConfig(apiToken string) Option {
	return SetToken(apiToken)
}

// SetClient specifies the custom http client used by the [Client] to make
// requests.
func SetClient(httpClient *http.Client) Option {
	return func(c *Client) error {
		if httpClient == nil {
			return nil
		}
		c.httpClient = httpClient
		return nil
	}
}

// SetUserAgent specifies the user agent used by the [Client].
func SetUserAgent(userAgent string) Option {
	return func(c *Client) error {
		c.userAgent = userAgent
		return nil
	}
}

// SetNoEnv prevents the [Client] from deriving its configuration from the
// environment (by auto reading "AXIOM_*" environment variables).
func SetNoEnv() Option {
	return func(c *Client) error {
		c.noEnv = true
		return nil
	}
}

// SetNoTracing prevents the [Client] from acquiring a tracer from the global
// tracer provider, even if one is configured.
func SetNoTracing() Option {
	return func(c *Client) error {
		c.tracer = noop.Tracer{}
		return nil
	}
}
