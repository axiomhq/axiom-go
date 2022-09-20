package config

import "strings"

// IsAPIToken returns true if the given access token is an API token.
func IsAPIToken(token string) bool {
	return strings.HasPrefix(token, "xaat-")
}

// IsPersonalToken returns true if the given access token is a personal token.
func IsPersonalToken(token string) bool {
	return strings.HasPrefix(token, "xapt-")
}

// IsValidToken returns true if the given access token is a valid Axiom access
// token.
func IsValidToken(token string) bool {
	return IsAPIToken(token) || IsPersonalToken(token)
}
