package query

import (
	"encoding/json"
	"fmt"
	"strings"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=AggregationOp -linecomment -output=aggregation_string.go

// An AggregationOp describes the [Aggregation] operation applied on a [Field].
type AggregationOp uint8

// All available [Aggregation] operations.
const (
	OpUnknown AggregationOp = iota // unknown

	OpCount               // count
	OpCountIf             // countif
	OpDistinct            // distinct
	OpDistinctIf          // distinctif
	OpSum                 // sum
	OpSumIf               // sumif
	OpAvg                 // avg
	OpAvgIf               // avgif
	OpMin                 // min
	OpMinIf               // minif
	OpMax                 // max
	OpMaxIf               // maxif
	OpTopk                // topk
	OpPercentiles         // percentiles
	OpHistogram           // histogram
	OpStandardDeviation   // stdev
	OpStandardDeviationIf // stdevif
	OpVariance            // variance
	OpVarianceIf          // varianceif
	OpArgMin              // argmin
	OpArgMax              // argmax
	OpRate                // rate
	OpPearson             // pearson_correlation
	OpMakeSet             // makeset
	OpMakeSetIf           // makesetif
	OpMakeList            // makelist
	OpMakeListIf          // makelistif
)

func aggregationOpFromString(s string) (op AggregationOp, err error) {
	switch strings.ToLower(s) {
	case OpCount.String():
		op = OpCount
	case OpCountIf.String():
		op = OpCountIf
	case OpDistinct.String():
		op = OpDistinct
	case OpDistinctIf.String():
		op = OpDistinctIf
	case OpSum.String():
		op = OpSum
	case OpSumIf.String():
		op = OpSumIf
	case OpAvg.String():
		op = OpAvg
	case OpAvgIf.String():
		op = OpAvgIf
	case OpMin.String():
		op = OpMin
	case OpMinIf.String():
		op = OpMinIf
	case OpMax.String():
		op = OpMax
	case OpMaxIf.String():
		op = OpMaxIf
	case OpTopk.String():
		op = OpTopk
	case OpPercentiles.String():
		op = OpPercentiles
	case OpHistogram.String():
		op = OpHistogram
	case OpStandardDeviation.String():
		op = OpStandardDeviation
	case OpStandardDeviationIf.String():
		op = OpStandardDeviationIf
	case OpVariance.String():
		op = OpVariance
	case OpVarianceIf.String():
		op = OpVarianceIf
	case OpArgMin.String():
		op = OpArgMin
	case OpArgMax.String():
		op = OpArgMax
	case OpRate.String():
		op = OpRate
	case OpPearson.String():
		op = OpPearson
	case OpMakeSet.String():
		op = OpMakeSet
	case OpMakeSetIf.String():
		op = OpMakeSetIf
	case OpMakeList.String():
		op = OpMakeList
	case OpMakeListIf.String():
		op = OpMakeListIf
	default:
		return OpUnknown, fmt.Errorf("unknown aggregation operation: %s", s)
	}

	return op, nil
}

// UnmarshalJSON implements [json.Unmarshaler]. It is in place to unmarshal the
// AggregationOp from the string representation the server returns.
func (op *AggregationOp) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}

	*op, err = aggregationOpFromString(s)

	return err
}

// Aggregation that is applied to a [Field] in a [Table].
type Aggregation struct {
	// Op is the aggregation operation. If the aggregation is aliased, the alias
	// is stored in the parent [Field.Name].
	Op AggregationOp `json:"name"`
	// Fields specifies the names of the fields this aggregation is computed on.
	// E.g. ["players"] for "topk(players, 10)".
	Fields []string `json:"fields"`
	// Args are the non-field arguments of the aggregation, if any. E.g. "10"
	// for "topk(players, 10)".
	Args []any `json:"args"`
}
