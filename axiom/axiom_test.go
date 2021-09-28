package axiom_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axiomhq/axiom-go/axiom"
)

//nolint:gosec // Chill bro, those are just for testing.
const (
	ingestTokenStr      = "xait-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
	personalTokenStr    = "xapt-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
	unspecifiedTokenStr = "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
)

func TestIsIngestToken(t *testing.T) {
	assert.True(t, axiom.IsIngestToken(ingestTokenStr))
	assert.False(t, axiom.IsIngestToken(personalTokenStr))
	assert.False(t, axiom.IsIngestToken(unspecifiedTokenStr))
}

func TestIsPersonalToken(t *testing.T) {
	assert.True(t, axiom.IsPersonalToken(personalTokenStr))
	assert.False(t, axiom.IsPersonalToken(ingestTokenStr))
	assert.False(t, axiom.IsPersonalToken(unspecifiedTokenStr))
}

func TestIsValidToken(t *testing.T) {
	assert.True(t, axiom.IsValidToken(ingestTokenStr))
	assert.True(t, axiom.IsValidToken(personalTokenStr))
	assert.False(t, axiom.IsValidToken(unspecifiedTokenStr))
}

func TestValidateEnvironment(t *testing.T) {
	defer os.Clearenv()

	tests := []struct {
		name        string
		environment map[string]string
		err         error
	}{
		{
			name: "no environment",
			err:  axiom.ErrMissingOrganizationID,
		},
		{
			name: "bad environment",
			environment: map[string]string{
				"AXIOM_ORG_ID": "mycompany-1234",
			},
			err: axiom.ErrMissingAccessToken,
		},
		{
			name: "cloud environment",
			environment: map[string]string{
				"AXIOM_TOKEN":  personalTokenStr,
				"AXIOM_ORG_ID": "mycompany-1234",
			},
		},
		{
			name: "selfhost environment",
			environment: map[string]string{
				"AXIOM_URL":   "https://axiom.internal.mycompany.org",
				"AXIOM_TOKEN": personalTokenStr,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()

			for k, v := range tt.environment {
				os.Setenv(k, v)
			}

			err := axiom.ValidateEnvironment()
			assert.Equal(t, tt.err, err)
		})
	}
}
