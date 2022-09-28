package axiom

import (
	"net/url"
	"reflect"

	"github.com/google/go-querystring/query"
)

// AddOptions adds the parameters in opt as URL query parameters to s. opt must
// be a struct whose fields may contain "url" tags.
//
// Ref: https://github.com/google/go-github/blob/master/github/github.go#L232.
func AddOptions(s string, opt any) (string, error) {
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
