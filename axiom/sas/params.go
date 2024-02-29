package sas

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
)

var (
	keyCodec   = base64.StdEncoding
	tokenCodec = base64.RawURLEncoding
)

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
	// ExpiryTime is the time after the token and signature are not valid
	// anymore. Can be a timestamp or an APL expression.
	ExpiryTime string `url:"exp"`
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
	} else if p.ExpiryTime == "" {
		return errors.New("expiry time is required")
	}
	return nil
}

// sign the parameters with the given key and return a base64 encoded token.
func (p Params) sign(key string) (string, error) {
	k, err := keyCodec.DecodeString(key)
	if err != nil {
		return "", fmt.Errorf("invalid key: %s", err)
	}

	h := hmac.New(sha256.New, k)
	if err = buildSignature(h, p); err != nil {
		return "", fmt.Errorf("computing hmac: %s", err)
	}

	return tokenCodec.EncodeToString(h.Sum(nil)), nil
}

// buildSignature builds the payload for a shared access token and writes it to
// the given [hash.Hash] in order to compute the signature. The format for the
// payload is a simple, newline decoded string composed of the following values
// in that order: organization ID, dataset name, filter, minimum start time,
// maximum end time, expiry time.
func buildSignature(h hash.Hash, params Params) error {
	paramList := []string{
		params.OrganizationID,
		params.Dataset,
		params.Filter,
		params.MinStartTime,
		params.MaxEndTime,
		params.ExpiryTime,
	}
	for idx, param := range paramList {
		if _, err := h.Write([]byte(param)); err != nil {
			return err
		} else if idx < len(paramList)-1 { // Skip newline for last param.
			if _, err = h.Write([]byte{'\n'}); err != nil {
				return err
			}
		}
	}
	return nil
}
