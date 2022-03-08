package apl

import (
	"net/url"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=Format -linecomment -output=format_string.go

// Format represents the format of an APL query.
type Format uint8

// All available query formats.
const (
	Legacy Format = iota // legacy
)

// EncodeValues implements `query.Encoder`. It is in place to encode the Format
// into a string URL value because that's what the server expects.
func (f Format) EncodeValues(key string, v *url.Values) error {
	v.Set(key, f.String())
	return nil
}
