// Code generated by "stringer -type=AggregationOp -linecomment -output=aggregation_string.go"; DO NOT EDIT.

package querylegacy

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[OpUnknown-0]
	_ = x[OpCount-1]
	_ = x[OpDistinct-2]
	_ = x[OpMakeSet-3]
	_ = x[OpMakeList-4]
	_ = x[OpSum-5]
	_ = x[OpAvg-6]
	_ = x[OpMin-7]
	_ = x[OpMax-8]
	_ = x[OpTopk-9]
	_ = x[OpPercentiles-10]
	_ = x[OpHistogram-11]
	_ = x[OpStandardDeviation-12]
	_ = x[OpVariance-13]
	_ = x[OpArgMin-14]
	_ = x[OpArgMax-15]
}

const _AggregationOp_name = "unknowncountdistinctmakesetmakelistsumavgminmaxtopkpercentileshistogramstdevvarianceargminargmax"

var _AggregationOp_index = [...]uint8{0, 7, 12, 20, 27, 35, 38, 41, 44, 47, 51, 62, 71, 76, 84, 90, 96}

func (i AggregationOp) String() string {
	if i >= AggregationOp(len(_AggregationOp_index)-1) {
		return "AggregationOp(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AggregationOp_name[_AggregationOp_index[i]:_AggregationOp_index[i+1]]
}
