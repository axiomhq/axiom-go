package query

import (
	"encoding/json"
	"fmt"
	"strings"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=FilterOp -linecomment -output=filter_string.go

// A FilterOp can be applied on queries to filter based on different conditions.
type FilterOp uint8

// All available query filter operations.
const (
	emptyFilterOp FilterOp = iota //

	OpAnd // and
	OpOr  // or
	OpNot // not

	// Works for strings and numbers.
	OpEqual     // ==
	OpNotEqual  // !=
	OpExists    // exists
	OpNotExists // not-exists

	// Only works for numbers.
	OpGreaterThan      // >
	OpGreaterThanEqual // >=
	OpLessThan         // <
	OpLessThanEqual    // <=

	// Only works for strings.
	OpStartsWith    // starts-with
	OpNotStartsWith // not-starts-with
	OpEndsWith      // ends-with
	OpNotEndsWith   // not-ends-with
	OpRegexp        // regexp
	OpNotRegexp     // not-regexp

	// Works for strings and arrays.
	OpContains    // contains
	OpNotContains // not-contains
)

func filterOpFromString(s string) (op FilterOp, err error) {
	switch strings.ToLower(s) {
	case emptyFilterOp.String():
		op = emptyFilterOp
	case OpAnd.String():
		op = OpAnd
	case OpOr.String():
		op = OpOr
	case OpNot.String():
		op = OpNot
	case OpEqual.String():
		op = OpEqual
	case OpNotEqual.String():
		op = OpNotEqual
	case OpExists.String():
		op = OpExists
	case OpNotExists.String():
		op = OpNotExists
	case OpGreaterThan.String():
		op = OpGreaterThan
	case OpGreaterThanEqual.String():
		op = OpGreaterThanEqual
	case OpLessThan.String():
		op = OpLessThan
	case OpLessThanEqual.String():
		op = OpLessThanEqual
	case OpStartsWith.String():
		op = OpStartsWith
	case OpNotStartsWith.String():
		op = OpNotStartsWith
	case OpEndsWith.String():
		op = OpEndsWith
	case OpNotEndsWith.String():
		op = OpNotEndsWith
	case OpRegexp.String():
		op = OpRegexp
	case OpNotRegexp.String():
		op = OpNotRegexp
	case OpContains.String():
		op = OpContains
	case OpNotContains.String():
		op = OpNotContains
	default:
		err = fmt.Errorf("unknown filter operation %q", s)
	}

	return op, err
}

// MarshalJSON implements `json.Marshaler`. It is in place to marshal the
// FilterOp to its string representation because that's what the server expects.
func (op FilterOp) MarshalJSON() ([]byte, error) {
	return json.Marshal(op.String())
}

// UnmarshalJSON implements `json.Unmarshaler`. It is in place to unmarshal the
// FilterOp from the string representation the server returns.
func (op *FilterOp) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}

	*op, err = filterOpFromString(s)

	return err
}

// Filter applied as part of a query.
type Filter struct {
	// Op is the operation of the filter.
	Op FilterOp `json:"op"`
	// Field the filter operation is performed on.
	Field string `json:"field"`
	// Value to perform the filter operation against.
	Value interface{} `json:"value"`
	// CaseSensitive specifies if the filter is case sensitive or not. Only
	// valid for OpStartsWith, OpNotStartsWith, OpEndsWith, OpNotEndsWith,
	// OpContains and OpNotContains.
	CaseSensitive bool `json:"caseSensitive"`
	// Children specifies child filters for the filter. Only valid for OpAnd,
	// OpOr and OpNot.
	Children []Filter `json:"children"`
}
