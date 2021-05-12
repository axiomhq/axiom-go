package axiom

import "strings"

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
