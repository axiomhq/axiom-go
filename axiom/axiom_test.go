package axiom_test

import (
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
