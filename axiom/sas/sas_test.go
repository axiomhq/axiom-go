package sas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testKeyStr = "aba84eee-3935-4b51-8aae-2c41b8693016"

func TestCreate(t *testing.T) {
	signature, err := Create(testKeyStr, Params{
		OrganizationID: "axiom",
		Dataset:        "logs",
		Filter:         `customer == "vercel"`,
		MinStartTime:   "ago(1h)",
		MaxEndTime:     "now",
	})
	require.NoError(t, err)
	require.NotEmpty(t, signature)

	assert.Equal(t, "dt=logs&fl=customer+%3D%3D+%22vercel%22&met=now&mst=ago%281h%29&oi=axiom&tk=0M41vwyiTVtAqW_aw8ZaIgayOlxnSwtFoFbywuQ-VBc%3D", signature)
}

func TestCreateToken(t *testing.T) {
	token, err := CreateToken(testKeyStr, Params{
		OrganizationID: "axiom",
		Dataset:        "logs",
		Filter:         `customer == "vercel"`,
		MinStartTime:   "ago(1h)",
		MaxEndTime:     "now",
	})
	require.NoError(t, err)
	require.NotEmpty(t, token)

	assert.Equal(t, "0M41vwyiTVtAqW_aw8ZaIgayOlxnSwtFoFbywuQ-VBc=", token)
}
