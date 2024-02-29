package sas

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/google/go-querystring/query"
)

// The parameter names for the shared access signature query string.
const (
	queryOrgID        = "oi"
	queryDataset      = "dt"
	queryFilter       = "fl"
	queryMinStartTime = "mst"
	queryMaxEndTime   = "met"
	queryExpiryTime   = "exp"
	queryToken        = "tk"
)

// Options are the url query parameters used to authenticate a query request.
type Options struct {
	Params

	// Token is the signature created from the other fields in the options.
	Token string `url:"tk"`
}

// Decode the given signature into a set of options.
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
			ExpiryTime:     q.Get(queryExpiryTime),
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

// Attach the options to the given request as a query string. Existing query
// parameters are retained unless they are overwritten by the key of one of the
// options.
func (o Options) Attach(req *http.Request) error {
	q, err := query.Values(o)
	if err != nil {
		return err
	}

	qc := req.URL.Query()
	for k := range q {
		qc.Set(k, q.Get(k))
	}
	req.URL.RawQuery = qc.Encode()

	return nil
}

// Encode the options into a url query string.
func (o Options) Encode() (string, error) {
	q, err := query.Values(o)
	if err != nil {
		return "", err
	}
	return q.Encode(), nil
}
