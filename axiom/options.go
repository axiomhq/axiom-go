package axiom

import (
	"net/url"
	"reflect"

	"github.com/google/go-querystring/query"
)

// ListOptions specifies the optional parameters to various List methods that
// support result limits.
type ListOptions struct {
	// Limit of the of the result set.
	Limit uint `url:"limit,omitempty"`
	// Offset the result set has from its base.
	Offset uint `url:"offset,omitempty"`
}

// addOptions adds the parameters in opt as URL query parameters to s. opt must
// be a struct whose fields may contain "url" tags.
// https://github.com/google/go-github/blob/master/github/github.go#L232
func addOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}
