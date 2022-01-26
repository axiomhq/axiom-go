package axiom

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"time"

	urlquery "github.com/google/go-querystring/query"
	"github.com/google/uuid"

	"github.com/axiomhq/axiom-go/axiom/query"
)

// SharedAccessOptions represents the options for creating a shared access token
// or shared access signature.
type SharedAccessOptions struct {
	// OrganizationID is the ID of the organization the token and signature is
	// valid for.
	OrganizationID string `url:"oi,omitempty"`
	// Dataset name the token and signature is valid for.
	Dataset string `url:"dt,omitempty"`
	// Filter is the top-level query filter to apply to all query requests
	// the token and signature is valid for.
	Filter *query.Filter `url:"fl,omitempty"`
	// MinStartTime is the earliest query start time the token and signature is
	// valid for.
	MinStartTime time.Time `url:"mst,omitempty"`
	// MaxEndTime is the latest query end time the token and signature is valid
	// for.
	MaxEndTime time.Time `url:"met,omitempty"`
}

// CreateSharedAccessSignature creates a shared access signature using the given
// key and valid for the given options. The returned string is a query string
// that can be attached to a URL.
//
// Shared access is only supported on Axiom Cloud.
func CreateSharedAccessSignature(keyStr string, options SharedAccessOptions) (string, error) {
	token, err := CreateSharedAccessToken(keyStr, options)
	if err != nil {
		return "", err
	}

	q, err := urlquery.Values(struct {
		SharedAccessOptions
		Token string `url:"tk,omitempty"`
	}{
		SharedAccessOptions: options,
		Token:               token,
	})
	if err != nil {
		return "", err
	}

	return q.Encode(), nil
}

// CreateSharedAccessToken creates a shared access token signed with the given
// key and valid for the given options.
//
// This function is only useful if the intention is to create the shared
// signature without the help of `CreateSharedAccessSignature`.
//
// Shared access is only supported on Axiom Cloud.
func CreateSharedAccessToken(keyStr string, options SharedAccessOptions) (string, error) {
	key, err := uuid.Parse(keyStr)
	if err != nil {
		return "", err
	}

	q, err := urlquery.Values(options)
	if err != nil {
		return "", err
	}
	qs := q.Encode()

	h := hmac.New(sha256.New, key[:])
	if _, err = h.Write([]byte(qs)); err != nil {
		return "", err
	}
	token := h.Sum(nil)

	return base64.URLEncoding.EncodeToString(token), nil
}
