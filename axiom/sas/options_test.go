package sas

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/axiom/query"
	"github.com/axiomhq/axiom-go/internal/test/testhelper"
)

func TestOptions_optionsFromURLValues(t *testing.T) {
	exp := getOptions(t)

	q := make(url.Values)
	q.Add("oi", "axiom")
	q.Add("dt", "logs")
	q.Add("fl", `{"op":"==","fd":"customer","vl":"vercel","cs":true}`)
	q.Add("mst", "2022-01-01T00:00:00Z")
	q.Add("met", "2023-01-01T00:00:00Z")

	options, err := optionsFromURLValues(q)
	require.NoError(t, err)
	require.NotEmpty(t, options)

	assert.Equal(t, exp, options)
}

func TestOptions_urlValues(t *testing.T) {
	expFilter := `{"op":"==","fd":"customer","vl":"vercel","cs":true,"ch":[{"op":"==","fd":"project","vl":"project-123","cs":false}]}`

	options := Options{
		OrganizationID: "axiom",
		Dataset:        "logs",
		Filter: query.Filter{
			Op:            query.OpEqual,
			Field:         "customer",
			Value:         "vercel",
			CaseSensitive: true,
			Children: []query.Filter{
				{
					Op:            query.OpEqual,
					Field:         "project",
					Value:         "project-123",
					CaseSensitive: false,
				},
			},
		},
		MinStartTime: testhelper.MustTimeParse(t, time.RFC3339, "2022-01-01T00:00:00Z"),
		MaxEndTime:   testhelper.MustTimeParse(t, time.RFC3339, "2023-01-01T00:00:00Z"),
	}

	q, err := options.urlValues()
	require.NoError(t, err)

	if assert.Len(t, q, 5) {
		assert.Equal(t, "axiom", q.Get("oi"))
		assert.Equal(t, "logs", q.Get("dt"))
		assert.Equal(t, expFilter, q.Get("fl"))
		assert.Equal(t, "2022-01-01T00:00:00Z", q.Get("mst"))
		assert.Equal(t, "2023-01-01T00:00:00Z", q.Get("met"))
	}
}

func TestOptions_validate(t *testing.T) {
	var options Options

	err := options.validate()
	assert.EqualError(t, err, "organization ID is required")

	options.OrganizationID = "axiom"

	err = options.validate()
	assert.EqualError(t, err, "dataset is required")

	options.Dataset = "logs"

	err = options.validate()
	assert.EqualError(t, err, "filter is required")

	options.Filter = query.Filter{
		Op:            query.OpEqual,
		Field:         "customer",
		Value:         "vercel",
		CaseSensitive: true,
	}

	err = options.validate()
	assert.EqualError(t, err, "minimum start time is required")

	options.MinStartTime = testhelper.MustTimeParse(t, time.RFC3339, "2022-01-01T00:00:00Z")

	err = options.validate()
	assert.EqualError(t, err, "maximum end time is required")

	options.MaxEndTime = testhelper.MustTimeParse(t, time.RFC3339, "2023-01-01T00:00:00Z")

	err = options.validate()
	assert.NoError(t, err)
}
