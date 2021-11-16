package axiom

import (
	"context"
	"strings"
)

// IsAPIToken returns true if the given access token is an API token.
func IsAPIToken(token string) bool {
	return strings.HasPrefix(token, "xaat-")
}

// IsIngestToken returns true if the given access token is an ingest token.
func IsIngestToken(token string) bool {
	return strings.HasPrefix(token, "xait-")
}

// IsPersonalToken returns true if the given access token is a personal token.
func IsPersonalToken(token string) bool {
	return strings.HasPrefix(token, "xapt-")
}

// IsValidToken returns true if the given access token is a valid Axiom access
// token.
func IsValidToken(token string) bool {
	return IsAPIToken(token) || IsIngestToken(token) || IsPersonalToken(token)
}

// ValidateEnvironment returns nil if the environment variables, needed to
// configure a new Client, are present and syntactically valid. Otherwise, it
// returns an appropriate error.
func ValidateEnvironment() error {
	var client Client
	return client.populateClientFromEnvironment()
}

// ValidateCredentials returns nil if the environment variables that configure a
// Client are valid. Otherwise, it returns an appropriate error. This function
// establishes a connection to the configured Axiom deployment.
func ValidateCredentials(ctx context.Context) error {
	client, err := NewClient()
	if err != nil {
		return err
	}
	return client.ValidateCredentials(ctx)
}
