package query

import (
	"encoding/json"
	"fmt"
	"strings"
)

// A FieldType describes the type of a [Field].
type FieldType uint16

// All available [Field] types. Conforms to DB Types
const (
	TypeInvalid  FieldType = 0         // invalid
	TypeUnknown  FieldType = 1 << iota // unknown
	TypeInteger                        // integer
	TypeString                         // string
	TypeBool                           // boolean
	TypeDateTime                       // datetime
	TypeFloat                          // float
	TypeTimespan                       // timespan
	TypeMap                            // map
	TypeArray                          // array

	maxFieldType
)

func fieldTypeFromString(s string) (ft FieldType, err error) {
	types := strings.Split(s, "|")

	// FIXME(lukasmalkmus): It looks like there are more/different type aliases
	// then documented: https://axiom.co/docs/apl/data-types/scalar-data-types.
	for _, t := range types {
		switch strings.ToLower(t) {
		case TypeBool.String(), "boolean":
			ft |= TypeBool
		case TypeDateTime.String(), "date":
			ft |= TypeDateTime
		case TypeInteger.String(), "int":
			ft |= TypeInteger
		case TypeFloat.String(), "double", "float", "float64": // "float" and "float64" are not documented.
			ft |= TypeFloat
		case TypeString.String():
			ft |= TypeString
		case TypeTimespan.String(), "time":
			ft |= TypeTimespan
		case TypeArray.String():
			ft |= TypeArray
		case TypeMap.String():
			ft |= TypeMap
		case TypeUnknown.String():
			ft |= TypeUnknown
		default:
			return TypeInvalid, fmt.Errorf("invalid field type: %s", t)
		}
	}

	return ft, nil
}

// String returns a string representation of the field type.
//
// It implements [fmt.Stringer].
func (ft FieldType) String() string {
	if ft >= maxFieldType {
		return fmt.Sprintf("<invalid field type: %d (%08b)>", ft, ft)
	}

	//nolint:exhaustive // maxFieldType is not a valid field type and already
	// handled above.
	switch ft {
	case TypeBool:
		return "boolean"
	case TypeDateTime:
		return "datetime"
	case TypeInteger:
		return "integer"
	case TypeFloat:
		return "float"
	case TypeString:
		return "string"
	case TypeTimespan:
		return "timespan"
	case TypeArray:
		return "array"
	case TypeMap:
		return "map"
	case TypeUnknown:
		return "unknown"
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
