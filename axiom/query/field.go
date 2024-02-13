package query

import (
	"encoding/json"
	"fmt"
	"strings"
)

// A FieldType describes the type of a [Field].
type FieldType uint16

// All available [Field] types.
const (
	TypeInvalid    FieldType = 0         // invalid
	TypeBool       FieldType = 1 << iota // bool
	TypeDateTime                         // datetime
	TypeInt                              // int
	TypeLong                             // long
	TypeReal                             // real
	TypeString                           // string
	TypeTimespan                         // timespan
	TypeArray                            // array
	TypeDictionary                       // dictionary
	maxFieldType
)

func fieldTypeFromString(s string) (ft FieldType, err error) {
	types := strings.Split(s, "|")

	// FIXME(lukasmalkmus): Correct type aliases.
	for _, t := range types {
		switch strings.ToLower(t) {
		case TypeBool.String():
			ft |= TypeBool
		case TypeDateTime.String():
			ft |= TypeDateTime
		case TypeInt.String(), "integer":
			ft |= TypeInt
		case TypeLong.String():
			ft |= TypeLong
		case TypeReal.String(), "float64":
			ft |= TypeReal
		case TypeString.String():
			ft |= TypeString
		case TypeTimespan.String():
			ft |= TypeTimespan
		case TypeArray.String():
			ft |= TypeArray
		case TypeDictionary.String():
			ft |= TypeDictionary
		default:
			return TypeInvalid, fmt.Errorf("unknown field type: %s", t)
		}
	}

	return ft, nil
}

// String returns a string representation of the field type.
//
// It implements [fmt.Stringer].
func (ft FieldType) String() string {
	if ft >= maxFieldType {
		return fmt.Sprintf("<unknown field type: %d (%08b)>", ft, ft)
	}

	//nolint:exhaustive // maxFieldType is not a valid field type and already
	// handled above.
	switch ft {
	case TypeBool:
		return "bool"
	case TypeDateTime:
		return "datetime"
	case TypeInt:
		return "int"
	case TypeLong:
		return "long"
	case TypeReal:
		return "real"
	case TypeString:
		return "string"
	case TypeTimespan:
		return "timespan"
	case TypeArray:
		return "array"
	case TypeDictionary:
		return "dictionary"
	}

	var res []string
	for fieldType := TypeBool; fieldType < maxFieldType; fieldType <<= 1 {
		if ft&fieldType != 0 {
			res = append(res, fieldType.String())
		}
	}
	return strings.Join(res, "|")
}

// UnmarshalJSON implements [json.Unmarshaler]. It is in place to unmarshal the
// FieldType from the string representation the server returns.
func (ft *FieldType) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}

	*ft, err = fieldTypeFromString(s)

	return err
}

// Field in a [Table].
type Field struct {
	// Name of the field.
	Name string `json:"name"`
	// Type of the field. Can also be composite types.
	Type FieldType `json:"type"`
	// Aggregation is the aggregation applied to the field.
	Aggregation *Aggregation `json:"agg"`
}
