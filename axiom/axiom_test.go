package axiom_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/internal/config"
	"github.com/axiomhq/axiom-go/internal/test/testhelper"
)

const (
	apiToken      = "xaat-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
	personalToken = "xapt-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX" //nolint:gosec // Chill, it's just testing.
)

func TestValidateEnvironment(t *testing.T) {
	tests := []struct {
		name        string
		environment map[string]string
		err         error
	}{
		{
			name: "no environment",
			err:  config.ErrMissingAccessToken,
		},
		{
			name: "bad environment",
			environment: map[string]string{
				"AXIOM_ORG_ID": "mycompany-1234",
			},
			err: config.ErrMissingAccessToken,
		},
		{
			name: "good environment",
			environment: map[string]string{
				"AXIOM_TOKEN":  personalToken,
				"AXIOM_ORG_ID": "mycompany-1234",
			},
		},
		{
			name: "good environment with api token",
			environment: map[string]string{
				"AXIOM_TOKEN": apiToken,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.SafeClearEnv(t)

			for k, v := range tt.environment {
				os.Setenv(k, v)
			}

			err := axiom.ValidateEnvironment()
			assert.Equal(t, tt.err, err)
		})
	}
}
