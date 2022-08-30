// Package PKCE implements Proof Key for Code Exchange by OAuth Public Clients.
//
// See also: https://datatracker.ietf.org/doc/html/rfc7636.
package pkce

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"fmt"

	"golang.org/x/oauth2"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=Method -linecomment -output=pkce_string.go

var encoding = base64.URLEncoding.WithPadding(base64.NoPadding)

// Method used to create the PKCE Code Challenge.
type Method uint8

// Available methods for creating the PKCE Code Challenge.
//
// If the client is capable of using MethodS256, it MUST use MethodS256, as
// MethodS256 is Mandatory To Implement (MTI) on the server. Clients are
// permitted to use MethodPlain only if they cannot support MethodS256 for some
// technical reason and know via out-of-band configuration that the server
// supports MethodPlain.
//
// The plain transformation is for compatibility with existing deployments and
// for constrained environments that can't use the S256 transformation.
//
// See also: https://datatracker.ietf.org/doc/html/rfc7636#section-4.2.
const (
	MethodPlain Method = iota + 1 // plain
	MethodS256                    // S256
)

// MethodFromString returns a `Method` from its string representation.
func MethodFromString(s string) (Method, error) {
	switch s {
	case MethodPlain.String():
		return MethodPlain, nil
	case MethodS256.String():
		return MethodS256, nil
	}
	return 0, fmt.Errorf("invalid method %q", s)
}

// AuthCodeOption returns an option compatible with
// `golang.org/x/oauth2.Config.Exchange()`.
func (m Method) AuthCodeOption() oauth2.AuthCodeOption {
	return oauth2.SetAuthURLParam("code_challenge_method", m.String())
}

// Verifier is a 43-octet URL safe PKCE Code Verifier.
//
// Use `String()` to get a string representation of `Verifier` or
// `AuthCodeOption()` to get an option compatible with
// `golang.org/x/oauth2.Config.AuthCodeURL()`.
type Verifier [43]byte

// New creates a new PKCE Code Verifier.
func New() (v Verifier, err error) {
	b := make([]byte, 32)
	if err := binary.Read(rand.Reader, binary.LittleEndian, b); err != nil {
		return v, err
	}

	encoding.Encode(v[:], b)

	return v, nil
}

// VerifierFromString returns a `Verifier` from its string representation.
func VerifierFromString(s string) (v Verifier) {
	copy(v[:], []byte(s))
	return
}

// Challenge creates the PKCE Code Challenge for the PKCE Code Verifier using
// the given `Method`.
//
// If the client is capable of using MethodS256, it MUST use MethodS256, as
// MethodS256 is Mandatory To Implement (MTI) on the server. Clients are
// permitted to use MethodPlain only if they cannot support MethodS256 for some
// technical reason and know via out-of-band configuration that the server
// supports MethodPlain.
//
// The plain transformation is for compatibility with existing deployments and
// for constrained environments that can't use the S256 transformation.
//
// See also: https://datatracker.ietf.org/doc/html/rfc7636#section-4.2.
func (v Verifier) Challenge(method Method) (c Challenge) {
	switch method {
	case MethodPlain:
		copy(c[:], v[:])
	case MethodS256:
		sum := sha256.Sum256(v[:])
		encoding.Encode(c[:], sum[:])
	default:
		panic(fmt.Errorf("unknown code challenge method %q", method))
	}
	return
}

// AuthCodeOption returns an option compatible with
// `golang.org/x/oauth2.Config.AuthCodeURL()`.
func (v Verifier) AuthCodeOption() oauth2.AuthCodeOption {
	return oauth2.SetAuthURLParam("code_verifier", v.String())
}

// String returns the string representation of the PKCE Code Verifier.
//
// It implements `fmt.Stringer`.
func (v Verifier) String() string {
	return string(v[:])
}

// Challenge is a 43-octet URL safe PKCE Code Challenge.
//
// Use `String()` to get a string representation of `Challenge` or
// `AuthCodeOption()` to get an option compatible with
// `golang.org/x/oauth2.Config.Exchange()`.
type Challenge [43]byte

// ChallengeFromString returns a `Challenge` from its string representation.
func ChallengeFromString(s string) (c Challenge) {
	copy(c[:], []byte(s))
	return
}

// AuthCodeOption returns an option compatible with
// `golang.org/x/oauth2.Config.Exchange()`.
func (c Challenge) AuthCodeOption() oauth2.AuthCodeOption {
	return oauth2.SetAuthURLParam("code_challenge", c.String())
}

// Verify the PKCE Code Challenge using the given PKCE Code Verifier and method.
func (c Challenge) Verify(verifier Verifier, method Method) bool {
	challenge := verifier.Challenge(method)
	return subtle.ConstantTimeCompare(c[:], challenge[:]) == 1
}

// String returns the string representation of the PKCE Code Challenge.
//
// It implements `fmt.Stringer`.
func (c Challenge) String() string {
	return string(c[:])
}
