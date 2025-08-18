package querylegacy

import (
	"encoding/json"
	"strings"
)

//go:generate go tool stringer -type=AggregationOp -linecomment -output=aggregation_string.go

// An AggregationOp can be applied on queries to aggregate based on different
// conditions.
type AggregationOp uint8

// All available query aggregation operations.
const (
	OpUnknown AggregationOp = iota // unknown

	// Works with all types, field should be "*".
	OpCount    // count
	OpDistinct // distinct
	OpMakeSet  // makeset
	OpMakeList // makelist

	// Only works for numbers.
	OpSum               // sum
	OpAvg               // avg
	OpMin               // min
	OpMax               // max
	OpTopk              // topk
	OpPercentiles       // percentiles
	OpHistogram         // histogram
	OpStandardDeviation // stdev
	OpVariance          // variance
	OpArgMin            // argmin
	OpArgMax            // argmax
)

func aggregationOpFromString(s string) (op AggregationOp) {
	switch strings.ToLower(s) {
	case OpCount.String():
		op = OpCount
	case OpDistinct.String():
		op = OpDistinct
	case OpMakeSet.String():
		op = OpMakeSet
	case OpMakeList.String():
		op = OpMakeList
	case OpSum.String():
		op = OpSum
	case OpAvg.String():
		op = OpAvg
	case OpMin.String():
		op = OpMin
	case OpMax.String():
		op = OpMax
	case OpTopk.String():
		op = OpTopk
	case OpPercentiles.String():
		op = OpPercentiles
	case OpHistogram.String():
		op = OpHistogram
	case OpStandardDeviation.String():
		op = OpStandardDeviation
	case OpVariance.String():
		op = OpVariance
	case OpArgMin.String():
		op = OpArgMin
	case OpArgMax.String():
		op = OpArgMax
	default:
		op = OpUnknown
	}

	return op
}

// MarshalJSON implements [json.Marshaler]. It is in place to marshal the
// AggregationOp to its string representation because that's what the server
// expects.
func (op AggregationOp) MarshalJSON() ([]byte, error) {
	return json.Marshal(op.String())
}

// UnmarshalJSON implements [json.Unmarshaler]. It is in place to unmarshal the
// AggregationOp from the string representation the server returns.
func (op *AggregationOp) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}

	*op = aggregationOpFromString(s)

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
	// Argument to the aggregation. Only valid for [OpDistinctIf], [OpTopk],
	// [OpPercentiles] and [OpHistogram] aggregations.
	Argument any `json:"argument"`
}
