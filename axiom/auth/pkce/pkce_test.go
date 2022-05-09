package pkce

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEncoding makes sure that the base64 encoding used is correct.
//
// See https://tools.ietf.org/html/rfc7636#appendix-B for a description of the
// test data used.
func TestEncoding(t *testing.T) {
	b := [32]byte{116, 24, 223, 180, 151, 153, 224, 37, 79, 250, 96, 125, 216, 173,
		187, 186, 22, 212, 37, 77, 105, 214, 191, 240, 91, 88, 5, 88, 83,
		132, 141, 121}

	assert.Equal(t, "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk", encoding.EncodeToString(b[:]))
}

// TestChallengeS256 makes sure that the Code Challenge is calculated correctly.
//
// See https://tools.ietf.org/html/rfc7636#appendix-B for a description of the
// test data used.
func TestChallengeS256(t *testing.T) {
	verifier := VerifierFromString("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk")

	challenge := verifier.Challenge(MethodS256)
	assert.Equal(t, "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM", challenge.String())
}

func TestPlain(t *testing.T) {
	verifier, err := New()
	require.NoError(t, err)
	assert.NotEmpty(t, verifier)

	challenge := verifier.Challenge(MethodPlain)
	assert.EqualValues(t, verifier, challenge)

	assert.True(t, challenge.Verify(verifier, MethodPlain))
}

func TestS256(t *testing.T) {
	verifier, err := New()
	require.NoError(t, err)
	assert.NotEmpty(t, verifier)

	challenge := verifier.Challenge(MethodS256)
	assert.NotEqualValues(t, verifier, challenge)

	assert.True(t, challenge.Verify(verifier, MethodS256))
}
