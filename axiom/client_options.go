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

// SetNoRetry prevents the [Client] from auto-retrying failed HTTP requests
// under certain circumstances.
func SetNoRetry() Option {
	return func(c *Client) error {
		c.noRetry = true
		return nil
	}
}

// SetNoTracing prevents the [Client] from acquiring a tracer. It doesn't
// affect the default HTTP client transport used by the [Client], which uses
// [otelhttp.NewTransport] to create a new trace for each outgoing HTTP request.
// To prevent that behavior, users must provide their own HTTP client via
// [SetClient].
func SetNoTracing() Option {
	return func(c *Client) error {
		c.tracer = noop.Tracer{}
		return nil
	}
}

// SetEdgeURL specifies the edge URL used by the [Client] for ingest and query
// operations. The URL should include the scheme (e.g., "https://eu-central-1.aws.edge.axiom.co").
// When set, ingest requests are sent to "{edgeURL}/v1/ingest/{dataset}" and query
// requests are sent to "{edgeURL}/v1/query/_apl".
//
// Can also be specified using the "AXIOM_EDGE_URL" environment variable.
func SetEdgeURL(edgeURL string) Option {
	return func(c *Client) error { return c.config.Options(config.SetEdgeURL(edgeURL)) }
}
