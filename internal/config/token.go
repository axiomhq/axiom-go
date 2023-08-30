package config

import (
	"strings"

	"github.com/axiomhq/axiom-go/axiom/sas"
)

// IsAPIToken returns true if the given token is an api token. It does not
// validate the token itself.
func IsAPIToken(token string) bool {
	return strings.HasPrefix(token, "xaat-")
}

// IsPersonalToken returns true if the given token is a personal token. It does
// not validate the token itself.
func IsPersonalToken(token string) bool {
	return strings.HasPrefix(token, "xapt-")
}

// IsSharedAccessSignature returns true if the given signature is a shared
// access signature. It does not validate the signature itself.
func IsSharedAccessSignature(signature string) bool {
	_, err := sas.Decode(signature)
	return err == nil
}

// IsValidCredential returns true if the given credential is a valid Axiom
// token or shared access signature. It does not validate the token or signature
// itself.
func IsValidCredential(credential string) bool {
	return IsAPIToken(credential) ||
		IsPersonalToken(credential) ||
		IsSharedAccessSignature(credential)
}
