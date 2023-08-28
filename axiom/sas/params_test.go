package sas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParams_Validate(t *testing.T) {
	var params Params

	err := params.Validate()
	assert.EqualError(t, err, "organization ID is required")

	params.OrganizationID = "axiom"

	err = params.Validate()
	assert.EqualError(t, err, "dataset is required")

	params.Dataset = "logs"

	err = params.Validate()
	assert.EqualError(t, err, "filter is required")

	params.Filter = `customer == "vercel"`

	err = params.Validate()
	assert.EqualError(t, err, "minimum start time is required")

	params.MinStartTime = "ago(1h)"

	err = params.Validate()
	assert.EqualError(t, err, "maximum end time is required")

	params.MaxEndTime = "now"

	err = params.Validate()
	assert.EqualError(t, err, "expiry time is required")

	params.ExpiryTime = "now"

	err = params.Validate()
	assert.NoError(t, err)
}

func TestParams_sign(t *testing.T) {
	params := Params{
		OrganizationID: "axiom",
		Dataset:        "logs",
		Filter:         `customer == "vercel"`,
		MinStartTime:   "ago(1h)",
		MaxEndTime:     "now",
		ExpiryTime:     "now",
	}

	token, err := params.sign(testKeyStr)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	assert.Equal(t, "zdLDQdmMUIz1glTnQUCVJpYdZSIRLIPAj-c-y8zqph0", token)
}
