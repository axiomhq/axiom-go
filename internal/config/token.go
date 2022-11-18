package config

import "strings"

// IsAPIToken returns true if the given token is an api token.
func IsAPIToken(token string) bool {
	return strings.HasPrefix(token, "xaat-")
}

// IsPersonalToken returns true if the given token is a personal token.
func IsPersonalToken(token string) bool {
	return strings.HasPrefix(token, "xapt-")
}

// IsValidToken returns true if the given token is a valid Axiom token.
func IsValidToken(token string) bool {
	return IsAPIToken(token) || IsPersonalToken(token)
}
