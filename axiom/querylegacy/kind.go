package querylegacy

import (
	"encoding/json"
	"fmt"
	"net/url"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=Kind -linecomment -output=kind_string.go

// Kind represents the role of a query.
type Kind uint8

// All available query kinds.
const (
	emptyKind Kind = iota //

	Analytics // analytics
	Stream    // stream
	APL       // apl
)

func kindFromString(s string) (k Kind, err error) {
	switch s {
	case emptyKind.String():
		k = emptyKind
	case Analytics.String():
		k = Analytics
	case Stream.String():
		k = Stream
	case APL.String():
		k = APL
	default:
		err = fmt.Errorf("unknown query kind %q", s)
	}

	return k, err
}

// MarshalJSON implements [json.Marshaler]. It is in place to marshal the kind to
// its string representation because that's what the server expects.
func (k Kind) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

// UnmarshalJSON implements [json.Unmarshaler]. It is in place to unmarshal the
// kind from the string representation the server returns.
func (k *Kind) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}

	*k, err = kindFromString(s)

	return err
}

// EncodeValues implements [github.com/google/go-querystring/query.Encoder]. It
// is in place to encode the kind into a string URL value because that's what
// the server expects.
func (k Kind) EncodeValues(key string, v *url.Values) error {
	v.Set(key, k.String())
	return nil
}
