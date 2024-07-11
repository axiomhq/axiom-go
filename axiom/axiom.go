package axiom

import (
	"context"

	"github.com/axiomhq/axiom-go/internal/config"
)

// ValidateEnvironment returns nil if the environment variables, needed to
// configure a new [Client], are present and syntactically valid. Otherwise, it
// returns an appropriate error.
func ValidateEnvironment() error {
	var cfg config.Config
	if err := cfg.IncorporateEnvironment(); err != nil {
		return err
	}
	return cfg.Validate()
}

// ValidateCredentials returns nil if the environment variables that configure a
// [Client] are valid. Otherwise, it returns an appropriate error. This function
// establishes a connection to the configured Axiom API.
func ValidateCredentials(ctx context.Context) error {
	client, err := NewClient()
	if err != nil {
		return err
	}
	return client.ValidateCredentials(ctx)
}
