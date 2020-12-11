package query

// A FilterOp can be applied on queries to filter based on different conditions.
type FilterOp string

// All available query filter operations.
const (
	OpAnd FilterOp = "and"
	OpOr  FilterOp = "or"
	OpNot FilterOp = "not"

	// Works for strings and numbers.
	OpEqual     FilterOp = "=="
	OpNotEqual  FilterOp = "!="
	OpExists    FilterOp = "exists"
	OpNotExists FilterOp = "not-exists"

	// Only works for numbers.
	OpGreaterThan      FilterOp = ">"
	OpGreaterThanEqual FilterOp = ">="
	OpLessThan         FilterOp = "<"
	OpLessThanEqual    FilterOp = "<="

	// Only works for strings.
	OpStartsWith    FilterOp = "starts-with"
	OpNotStartsWith FilterOp = "not-starts-with"
	OpEndsWith      FilterOp = "ends-with"
	OpNotEndsWith   FilterOp = "not-ends-with"
	OpRegexp        FilterOp = "regexp"
	OpNotRegexp     FilterOp = "not-regexp"

	// Works for strings and arrays.
	OpContains    FilterOp = "contains"
	OpNotContains FilterOp = "not-contains"
)

// Filter applied as part of a query.
type Filter struct {
	// Op is the operation of the filter.
	Op FilterOp `json:"op"`
	// Field the filter operation is performed on.
	Field string `json:"field"`
	// Value to perform the filter operation against.
	Value interface{} `json:"value"`
	// CaseInsensitive specifies if the filter is case insensitive or not. Only
	// valid for OpStartsWith, OpNotStartsWith, OpEndsWith, OpNotEndsWith,
	// OpContains and OpNotContains.
	CaseInsensitive bool `json:"caseInsensitive"`
	// Children specifies child filters for the filter. Only valid for OpAnd,
	// OpOr and OpNot.
	Children []Filter `json:"children"`
}
