package query

import (
	"encoding/json"
	"fmt"
	"net/url"
)

//go:generate go run -mod=mod golang.org/x/tools/cmd/stringer -type=Kind -linecomment -output=kind_string.go

// Kind represents the role of a query.
type Kind uint8

// All available query kinds.
const (
	Analytics Kind = iota + 1 // analytics
	Stream                    // stream
	APL                       // apl
)

// MarshalJSON implements json.Marshaler. It is in place to marshal the Kind to
// its string representation because that's what the server expects.
func (k Kind) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

// UnmarshalJSON implements json.Unmarshaler. It is in place to unmarshal the
// Kind from the string representation the server returns.
func (k *Kind) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch s {
	case Analytics.String():
		*k = Analytics
	case Stream.String():
		*k = Stream
	case APL.String():
		*k = APL
	default:
		return fmt.Errorf("unknown query kind %q", s)
	}

	return nil
}

// EncodeValues implements `query.Encoder`. It is in place to encode the Kind
// into a string URL value because that's what the server expects.
func (k Kind) EncodeValues(key string, v *url.Values) error {
	v.Set(key, k.String())
	return nil
}
