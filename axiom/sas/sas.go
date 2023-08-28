package sas

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"

	"github.com/google/uuid"
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

var tokenCodec = base64.URLEncoding

// Create creates a shared access signature using the given signing key and
// valid for the given options. The returned string is a query string that can
// be attached to a URL.
func Create(keyStr string, options Options) (string, error) {
	token, err := CreateToken(keyStr, options)
	if err != nil {
		return "", err
	}

	q, err := options.urlValues()
	if err != nil {
		return "", err
	}
	q.Set("tk", token)

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
// for the given options.
//
// This function is only useful if the intention is to create the shared access
// signature manually and without the help of [Create].
func CreateToken(keyStr string, options Options) (string, error) {
	if err := options.validate(); err != nil {
		return "", err
	}

	q, err := options.urlValues()
	if err != nil {
		return "", err
	}

	token, err := createRawToken(keyStr, q)
	if err != nil {
		return "", err
	}

	return tokenCodec.EncodeToString(token), nil
}

// Verify the given shared access signature string using the given signing key.
//
// An error is returend if the signature can't be processed. It will always
// return "false" in that case.
//
// If no error is returned it returns "true" if the signature is valid.
// Otherwise it returns "false".
func Verify(keyStr, signature string) (bool, Options, error) {
	q, err := url.ParseQuery(signature)
	if err != nil {
		return false, Options{}, err
	}

	options, err := optionsFromURLValues(q)
	if err != nil {
		return false, options, err
	}

	givenToken := q.Get(queryToken)
	if givenToken == "" {
		return false, options, errors.New("missing token")
	}
	q.Del(queryToken)

	givenTokenBytes, err := tokenCodec.DecodeString(givenToken)
	if err != nil {
		return false, options, err
	}

	computedToken, err := createRawToken(keyStr, q)
	if err != nil {
		return false, options, err
	}

	return hmac.Equal(givenTokenBytes, computedToken), options, nil
}

// VerifyToken the the validity of given shared access token for the given
// options using the given signing key.
//
// An error is returend if the signature can't be processed. It will always
// return "false" in that case.
//
// If no error is returned it returns "true" if the signature is valid.
// Otherwise it returns "false".
func VerifyToken(keyStr, token string, options Options) (bool, error) {
	q, err := options.urlValues()
	if err != nil {
		return false, err
	}

	givenTokenBytes, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return false, err
	}

	computedToken, err := createRawToken(keyStr, q)
	if err != nil {
		return false, err
	}

	return hmac.Equal(givenTokenBytes, computedToken), nil
}

// createRawToken creates a shared access token signed with the given key and
// valid for the given query parameter values. The result is not base64
// url-encoded.
func createRawToken(keyStr string, q url.Values) ([]byte, error) {
	key, err := uuid.Parse(keyStr)
	if err != nil {
		return nil, err
	}

	var (
		pl = buildSignaturePayload(q)
		h  = hmac.New(sha256.New, key[:])
	)
	if _, err = h.Write([]byte(pl)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// buildSignaturePayload builds the payload for a shared access token. The
// format is a simple, newline decoded string composed of the following values
// in that order: organization ID, dataset name, filter (JSON), minimum start
// time, maximum end time. The filter must be a encoded as a JSON string with
// child filters JSON encoded if included or left out, if empty.
func buildSignaturePayload(q url.Values) string {
	return strings.Join([]string{
		q.Get(queryOrgID),        // 1. Organization ID
		q.Get(queryDataset),      // 2. Dataset name
		q.Get(queryFilter),       // 3. Filter
		q.Get(queryMinStartTime), // 4. Minimum start time
		q.Get(queryMaxEndTime),   // 5. Maximum end time
	}, "\n")
}
