package sas

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptions_Decode(t *testing.T) {
	options, err := Decode("dt=logs&exp=now&fl=customer+%3D%3D+%22vercel%22&met=now&mst=ago%281h%29&oi=axiom&tk=zdLDQdmMUIz1glTnQUCVJpYdZSIRLIPAj-c-y8zqph0%3D")
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
		Token: "zdLDQdmMUIz1glTnQUCVJpYdZSIRLIPAj-c-y8zqph0=",
	}, options)
}

func TestOptions_Attach(t *testing.T) {
	options := Options{
		Params: Params{
			OrganizationID: "axiom",
			Dataset:        "logs",
			Filter:         `customer == "vercel"`,
			MinStartTime:   "ago(1h)",
			MaxEndTime:     "now",
			ExpiryTime:     "now",
		},
		Token: "zdLDQdmMUIz1glTnQUCVJpYdZSIRLIPAj-c-y8zqph0=",
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/datasets/_apl", nil)

	err := options.Attach(req)
	require.NoError(t, err)

	parsedOptions, err := Decode(req.URL.RawQuery)
	require.NoError(t, err)

	assert.Equal(t, options, parsedOptions)
}

func TestOptions_Encode(t *testing.T) {
	options := Options{
		Params: Params{
			OrganizationID: "axiom",
			Dataset:        "logs",
			Filter:         `customer == "vercel"`,
			MinStartTime:   "ago(1h)",
			MaxEndTime:     "now",
			ExpiryTime:     "now",
		},
		Token: "zdLDQdmMUIz1glTnQUCVJpYdZSIRLIPAj-c-y8zqph0=",
	}

	s, err := options.Encode()
	require.NoError(t, err)
	require.NotEmpty(t, s)

	assert.Equal(t, "dt=logs&exp=now&fl=customer+%3D%3D+%22vercel%22&met=now&mst=ago%281h%29&oi=axiom&tk=zdLDQdmMUIz1glTnQUCVJpYdZSIRLIPAj-c-y8zqph0%3D", s)
}
