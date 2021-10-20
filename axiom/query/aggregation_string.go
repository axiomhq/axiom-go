// Code generated by "stringer -type=AggregationOp -linecomment -output=aggregation_string.go"; DO NOT EDIT.

package query

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[UnknownAggregationOp-0]
	_ = x[OpCount-1]
	_ = x[OpCountIf-2]
	_ = x[OpCountDistinct-3]
	_ = x[OpCountDistinctIf-4]
	_ = x[OpSum-5]
	_ = x[OpAvg-6]
	_ = x[OpMin-7]
	_ = x[OpMax-8]
	_ = x[OpTopk-9]
	_ = x[OpPercentiles-10]
	_ = x[OpHistogram-11]
	_ = x[OpVariance-12]
	_ = x[OpStandardDeviation-13]
}

const _AggregationOp_name = "countcountifdistinctdistinctifsumavgminmaxtopkpercentileshistogramvariancestdev"

var _AggregationOp_index = [...]uint8{0, 0, 5, 12, 20, 30, 33, 36, 39, 42, 46, 57, 66, 74, 79}

func (i AggregationOp) String() string {
	if i >= AggregationOp(len(_AggregationOp_index)-1) {
		return "AggregationOp(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AggregationOp_name[_AggregationOp_index[i]:_AggregationOp_index[i+1]]
}
