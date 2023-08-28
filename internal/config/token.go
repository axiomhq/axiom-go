package config

import "strings"

// IsAPIToken returns true if the given token is an API token. It does not
// validate the token itself.
func IsAPIToken(token string) bool {
	return strings.HasPrefix(token, "xaat-")
}

// IsPersonalToken returns true if the given token is a personal token. It does
// not validate the token itself.
func IsPersonalToken(token string) bool {
	return strings.HasPrefix(token, "xapt-")
}

// IsValidToken returns true if the given credential is a valid Axiom
// token. It does not validate the token itself.
func IsValidToken(credential string) bool {
	return IsAPIToken(credential) || IsPersonalToken(credential)
}
