package config

import "net/url"

// An Option modifies the configuration. It is not safe for concurrent use.
type Option func(config *Config) error

// SetURL sets the base URL to use.
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

// SetAccessToken specifies the access token to use.
func SetAccessToken(accessToken string) Option {
	return func(config *Config) error {
		if !IsValidToken(accessToken) {
			return ErrInvalidToken
		}

		config.SetAccessToken(accessToken)

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
