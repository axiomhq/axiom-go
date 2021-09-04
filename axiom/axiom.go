package axiom

import (
	"context"
	"strings"
)

// IsIngestToken returns true if the given acces token is an ingest token. If
// false is returned, that does not imply that the token is a personal token.
func IsIngestToken(token string) bool {
	return strings.HasPrefix(token, "xait-")
}

// IsPersonalToken returns true if the given acces token is a personal token.
// If false is returned, that does not imply that the token is an ingest token.
func IsPersonalToken(token string) bool {
	return strings.HasPrefix(token, "xapt-")
}

// IsValidToken returns true if the given acces token is a valid Axiom token.
func IsValidToken(token string) bool {
	return IsIngestToken(token) || IsPersonalToken(token)
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
