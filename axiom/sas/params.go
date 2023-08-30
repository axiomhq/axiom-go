package sas

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var tokenCodec = base64.URLEncoding

// Params represents the parameters for creating a shared access token or a
// shared access signatur for a query request.
type Params struct {
	// OrganizationID is the ID of the organization the token and signature is
	// valid for.
	OrganizationID string `url:"oi"`
	// Dataset name the token and signature is valid for.
	Dataset string `url:"dt"`
	// Filter is the top-level query filter to apply to all query requests the
	// token and signature is valid for. Must be a valid APL filter expression.
	Filter string `url:"fl"`
	// MinStartTime is the earliest query start time the token and signature is
	// valid for. Can be a timestamp or an APL expression.
	MinStartTime string `url:"mst"`
	// MaxEndTime is the latest query end time the token and signature is valid
	// for. Can be a timestamp or an APL expression.
	MaxEndTime string `url:"met"`
}

// Validate makes sure that all query parameters are provided.
func (p Params) Validate() error {
	if p.OrganizationID == "" {
		return errors.New("organization ID is required")
	} else if p.Dataset == "" {
		return errors.New("dataset is required")
	} else if p.Filter == "" {
		return errors.New("filter is required")
	} else if p.MinStartTime == "" {
		return errors.New("minimum start time is required")
	} else if p.MaxEndTime == "" {
		return errors.New("maximum end time is required")
	}
	return nil
}

// sign the parameters with the given key and returns a base64 encoded token.
func (p Params) sign(key string) (string, error) {
	k, err := uuid.Parse(key)
	if err != nil {
		return "", fmt.Errorf("invalid key: %s", err)
	}

	var (
		pl = buildSignaturePayload(p)
		h  = hmac.New(sha256.New, k[:])
	)
	if _, err = h.Write([]byte(pl)); err != nil {
		return "", fmt.Errorf("computing hmac: %s", err)
	}

	token := h.Sum(nil)

	return tokenCodec.EncodeToString(token), nil
}

// buildSignaturePayload builds the payload for a shared access token. The
// format is a simple, newline decoded string composed of the following values
// in that order: organization ID, dataset name, filter, minimum start time,
// maximum end time.
func buildSignaturePayload(params Params) string {
	return strings.Join([]string{
		params.OrganizationID,
		params.Dataset,
		params.Filter,
		params.MinStartTime,
		params.MaxEndTime,
	}, "\n")
}
