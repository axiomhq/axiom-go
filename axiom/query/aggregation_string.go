// Code generated by "stringer -type=AggregationOp -linecomment -output=aggregation_string.go"; DO NOT EDIT.

package query

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[OpUnknown-0]
	_ = x[OpCount-1]
	_ = x[OpCountIf-2]
	_ = x[OpDistinct-3]
	_ = x[OpDistinctIf-4]
	_ = x[OpSum-5]
	_ = x[OpSumIf-6]
	_ = x[OpAvg-7]
	_ = x[OpAvgIf-8]
	_ = x[OpMin-9]
	_ = x[OpMinIf-10]
	_ = x[OpMax-11]
	_ = x[OpMaxIf-12]
	_ = x[OpTopk-13]
	_ = x[OpPercentiles-14]
	_ = x[OpHistogram-15]
	_ = x[OpStandardDeviation-16]
	_ = x[OpStandardDeviationIf-17]
	_ = x[OpVariance-18]
	_ = x[OpVarianceIf-19]
	_ = x[OpArgMin-20]
	_ = x[OpArgMax-21]
	_ = x[OpRate-22]
	_ = x[OpPearson-23]
	_ = x[OpMakeSet-24]
	_ = x[OpMakeSetIf-25]
	_ = x[OpMakeList-26]
	_ = x[OpMakeListIf-27]
}

const _AggregationOp_name = "unknowncountcountifdistinctdistinctifsumsumifavgavgifminminifmaxmaxiftopkpercentileshistogramstdevstdevifvariancevarianceifargminargmaxratepearson_correlationmakesetmakesetifmakelistmakelistif"

var _AggregationOp_index = [...]uint8{0, 7, 12, 19, 27, 37, 40, 45, 48, 53, 56, 61, 64, 69, 73, 84, 93, 98, 105, 113, 123, 129, 135, 139, 158, 165, 174, 182, 192}

func (i AggregationOp) String() string {
	if i >= AggregationOp(len(_AggregationOp_index)-1) {
		return "AggregationOp(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AggregationOp_name[_AggregationOp_index[i]:_AggregationOp_index[i+1]]
}
