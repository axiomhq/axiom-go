package axiom

import (
	"net/http"
	"net/url"
)

// An Option modifies the behaviour of the API client. If not otherwise
// specified by a specific option, they are safe to use even after API methods
// have been called. However, they are not safe to use while the client is
// performing an operation.
type Option func(c *Client) error

// SetURL sets the base URL used by the client. Can also be specified using the
// `AXIOM_URL` environment variable.
func SetURL(baseURL string) Option {
	return func(c *Client) (err error) {
		c.baseURL, err = url.ParseRequestURI(baseURL)
		return err
	}
}

// SetAccessToken specifies the access token to use. Can also be specified using
// the `AXIOM_TOKEN` environment variable.
func SetAccessToken(accessToken string) Option {
	return func(c *Client) error {
		if !IsValidToken(accessToken) {
			return ErrInvalidToken
		}
		c.accessToken = accessToken
		return nil
	}
}

// SetOrgID specifies the organization ID to use when connecting to Axiom Cloud.
// When a personal access token is used, this method can be used to switch
// between organizations without creating a new client instance. Can also be
// specified using the `AXIOM_ORG_ID` environment variable.
func SetOrgID(orgID string) Option {
	return func(c *Client) error {
		c.orgID = orgID
		return nil
	}
}

// SetCloudConfig specifies all properties needed in order to successfully
// connect to Axiom Cloud with a personal access token.
func SetCloudConfig(personalAccessToken, orgID string) Option {
	return func(c *Client) error {
		return c.Options(
			SetAccessToken(personalAccessToken),
			SetOrgID(orgID),
		)
	}
}

// SetCloudConfigWithAPIToken specifies all properties needed in order to successfully
// connect to Axiom Cloud with an API token.
func SetCloudConfigWithAPIToken(apiToken string) Option {
	return SetAccessToken(apiToken)
}

// SetSelfhostConfig specifies all properties needed in order to successfully
// connect to an Axiom Selfhost deployment.
func SetSelfhostConfig(deploymentURL, accessToken string) Option {
	return func(c *Client) error {
		return c.Options(
			SetURL(deploymentURL),
			SetAccessToken(accessToken),
		)
	}
}

// SetClient specifies a custom http client that should be used to make
// requests.
func SetClient(client *http.Client) Option {
	return func(c *Client) error {
		if client == nil {
			return nil
		}
		c.httpClient = client
		return nil
	}
}

// SetUserAgent sets the user agent used by the client.
func SetUserAgent(userAgent string) Option {
	return func(c *Client) error {
		c.userAgent = userAgent
		return nil
	}
}

// SetNoEnv prevents the client from deriving its configuration from the
// environment.
func SetNoEnv() Option {
	return func(c *Client) error {
		c.noEnv = true
		return nil
	}
}
