package sas

import (
	"errors"
	"fmt"

	"github.com/google/go-querystring/query"
)

// Create creates a shared access signature using the given signing key and
// valid for the given parameters. The returned string is a query string that
// can be attached to a URL.
func Create(keyStr string, params Params) (string, error) {
	token, err := CreateToken(keyStr, params)
	if err != nil {
		return "", err
	}

	q, err := query.Values(Options{
		Params: params,
		Token:  token,
	})
	if err != nil {
		return "", err
	}

	// Although officially there is no limit specified by RFC 2616, many
	// security protocols and recommendations state that maxQueryStrings on a
	// server should be set to a maximum character limit of 1024. While the
	// entire URL, including the querystring, should be set to a max of 2048
	// characters.
	sig := q.Encode()
	if len(sig) > 1023 { // 1024 - 1 for '?'
		return "", errors.New("signature too long")
	}

	return sig, nil
}

// CreateToken creates a shared access token signed with the given key and valid
// for the given parameters.
//
// This function is only useful if the intention is to create the shared access
// signature manually and without the help of [Create].
func CreateToken(key string, params Params) (string, error) {
	if err := params.Validate(); err != nil {
		return "", fmt.Errorf("invalid parameters: %s", err)
	}
	return params.sign(key)
}
