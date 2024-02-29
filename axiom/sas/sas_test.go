package sas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testKeyStr = "aeyGXNKLbqpPhBHqjHnVr4FS+eJ1d3LsheK1M8k6054="

func TestCreate(t *testing.T) {
	options, err := Create(testKeyStr, Params{
		OrganizationID: "axiom",
		Dataset:        "logs",
		Filter:         `customer == "vercel"`,
		MinStartTime:   "ago(1h)",
		MaxEndTime:     "now",
		ExpiryTime:     "now",
	})
	require.NoError(t, err)
	require.NotEmpty(t, options)

	assert.Equal(t, Options{
		Params: Params{
			OrganizationID: "axiom",
			Dataset:        "logs",
			Filter:         `customer == "vercel"`,
			MinStartTime:   "ago(1h)",
			MaxEndTime:     "now",
			ExpiryTime:     "now",
		},
		Token: "zdLDQdmMUIz1glTnQUCVJpYdZSIRLIPAj-c-y8zqph0",
	}, options)
}
