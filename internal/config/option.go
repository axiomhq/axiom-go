package config

import "net/url"

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
