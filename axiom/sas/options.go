package sas

import (
	"errors"
	"net/url"
)

// The parameter names for the shared access signature query string.
const (
	queryOrgID        = "oi"
	queryDataset      = "dt"
	queryFilter       = "fl"
	queryMinStartTime = "mst"
	queryMaxEndTime   = "met"
	queryToken        = "tk"
)

// Options are the url query parameters used to authenticate a query request.
type Options struct {
	Params

	// Token is the signature created from the other fields in the options.
	Token string `url:"tk"`
}

// Decode decodes the given signature into a set of options.
func Decode(signature string) (Options, error) {
	q, err := url.ParseQuery(signature)
	if err != nil {
		return Options{}, err
	}

	options := Options{
		Params: Params{
			OrganizationID: q.Get(queryOrgID),
			Dataset:        q.Get(queryDataset),
			Filter:         q.Get(queryFilter),
			MinStartTime:   q.Get(queryMinStartTime),
			MaxEndTime:     q.Get(queryMaxEndTime),
		},
		Token: q.Get(queryToken),
	}

	// Validate that the params are valid and the token is present.
	if err := options.Params.Validate(); err != nil {
		return options, err
	} else if options.Token == "" {
		return options, errors.New("missing token")
	}

	return options, nil
}
