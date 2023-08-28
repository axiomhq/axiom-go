package sas

import "fmt"

// Create creates the options that compose a shared access signature signed with
// the given key and valid for the given parameters. The options can be encoded
// into a query string by calling [Options.Encode] and attached to a URL by
// calling [Options.Attach].
func Create(key string, params Params) (Options, error) {
	if err := params.Validate(); err != nil {
		return Options{}, fmt.Errorf("invalid parameters: %s", err)
	}

	token, err := params.sign(key)
	if err != nil {
		return Options{}, err
	}

	return Options{
		Params: params,
		Token:  token,
	}, nil
}
