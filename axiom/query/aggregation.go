package query

import (
	"encoding/json"
	"fmt"
	"strings"
)

//go:generate go run -mod=mod golang.org/x/tools/cmd/stringer -type=AggregationOp -linecomment -output=aggregation_string.go

// An AggregationOp can be applied on queries to aggrgate based on different
// conditions.
type AggregationOp uint8

// All available query aggregation operations.
const (
	UnknownAggregationOp AggregationOp = iota //

	// Works with all types, field should be `*`.
	OpCount           // count
	OpCountIf         // countif
	OpCountDistinct   // distinct
	OpCountDistinctIf // distinctif

	// Only works for numbers.
	OpSum               // sum
	OpAvg               // avg
	OpMin               // min
	OpMax               // max
	OpTopk              // topk
	OpPercentiles       // percentiles
	OpHistogram         // histogram
	OpVariance          // variance
	OpStandardDeviation // stdev
)

// MarshalJSON implements json.Marshaler. It is in place to marshal the
// AggregationOp to its string representation because that's what the server
// expects.
func (op AggregationOp) MarshalJSON() ([]byte, error) {
	return json.Marshal(op.String())
}

// UnmarshalJSON implements json.Unmarshaler. It is in place to unmarshal the
// AggregationOp from the string representation the server returns.
func (op *AggregationOp) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch strings.ToLower(s) {
	case OpCount.String():
		*op = OpCount
	case OpCountIf.String():
		*op = OpCountIf
	case OpCountDistinct.String():
		*op = OpCountDistinct
	case OpCountDistinctIf.String():
		*op = OpCountDistinctIf
	case OpSum.String():
		*op = OpSum
	case OpAvg.String():
		*op = OpAvg
	case OpMin.String():
		*op = OpMin
	case OpMax.String():
		*op = OpMax
	case OpTopk.String():
		*op = OpTopk
	case OpPercentiles.String():
		*op = OpPercentiles
	case OpHistogram.String():
		*op = OpHistogram
	case OpVariance.String():
		*op = OpVariance
	case OpStandardDeviation.String():
		*op = OpStandardDeviation
	default:
		return fmt.Errorf("unknown aggregation operation %q", s)
	}

	return nil
}

// Aggregation performed as part of a query.
type Aggregation struct {
	// Alias for the aggregation.
	Alias string `json:"alias"`
	// Op is the operation of the aggregation.
	Op AggregationOp `json:"op"`
	// Field the aggregation operation is performed on.
	Field string `json:"field"`
	// Argument to the aggregation. Only valid for `OpCountDistinctIf`,
	// `OpTopk`, `OpPercentiles` and `OpHistogram`
	// aggregations.
	Argument interface{} `json:"argument"`
}
