package sas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptions_Decode(t *testing.T) {
	options, err := Decode("dt=logs&fl=customer+%3D%3D+%22vercel%22&met=now&mst=ago%281h%29&oi=axiom&tk=0M41vwyiTVtAqW_aw8ZaIgayOlxnSwtFoFbywuQ-VBc%3D")
	require.NoError(t, err)
	require.NotEmpty(t, options)

	assert.Equal(t, Options{
		Params: Params{
			OrganizationID: "axiom",
			Dataset:        "logs",
			Filter:         `customer == "vercel"`,
			MinStartTime:   "ago(1h)",
			MaxEndTime:     "now",
		},
		Token: "0M41vwyiTVtAqW_aw8ZaIgayOlxnSwtFoFbywuQ-VBc=",
	}, options)
}
