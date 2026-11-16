package query

// An AggregationOp is the name of the [Aggregation] operation applied on a
// [Field]. Names without a constant decode as is.
type AggregationOp string

// Known [Aggregation] operations. The server can return others.
const (
	OpUnknown AggregationOp = "unknown"

	OpCount                     AggregationOp = "count"
	OpCountIf                   AggregationOp = "countif"
	OpDistinct                  AggregationOp = "distinct"
	OpDistinctIf                AggregationOp = "distinctif"
	OpSum                       AggregationOp = "sum"
	OpSumIf                     AggregationOp = "sumif"
	OpAvg                       AggregationOp = "avg"
	OpAvgIf                     AggregationOp = "avgif"
	OpMin                       AggregationOp = "min"
	OpMinIf                     AggregationOp = "minif"
	OpMax                       AggregationOp = "max"
	OpMaxIf                     AggregationOp = "maxif"
	OpTopk                      AggregationOp = "topk"
	OpTopkIf                    AggregationOp = "topkif"
	OpPercentiles               AggregationOp = "percentiles"
	OpPercentilesIf             AggregationOp = "percentilesif"
	OpHistogram                 AggregationOp = "histogram"
	OpHistogramIf               AggregationOp = "histogramif"
	OpStandardDeviation         AggregationOp = "stdev"
	OpStandardDeviationIf       AggregationOp = "stdevif"
	OpVariance                  AggregationOp = "variance"
	OpVarianceIf                AggregationOp = "varianceif"
	OpArgMin                    AggregationOp = "argmin"
	OpArgMax                    AggregationOp = "argmax"
	OpRate                      AggregationOp = "rate"
	OpMakeSet                   AggregationOp = "makeset"
	OpMakeSetIf                 AggregationOp = "makesetif"
	OpMakeList                  AggregationOp = "makelist"
	OpMakeListIf                AggregationOp = "makelistif"
	OpComputed                  AggregationOp = "computed"
	OpSpotlight                 AggregationOp = "spotlight"
	OpTakeAny                   AggregationOp = "takeany"
	OpTakeAnyIf                 AggregationOp = "takeanyif"
	OpPhrases                   AggregationOp = "phrases"
	OpPresence                  AggregationOp = "presence"
	OpAvgWeighted               AggregationOp = "avgw"
	OpCountWeighted             AggregationOp = "countw"
	OpSumWeighted               AggregationOp = "sumw"
	OpHistogramWeighted         AggregationOp = "histogramw"
	OpPercentilesWeighted       AggregationOp = "percentilesw"
	OpStandardDeviationWeighted AggregationOp = "stdevw"
	OpVarianceWeighted          AggregationOp = "variancew"
	OpTopkWeighted              AggregationOp = "topkw"
)

func (op AggregationOp) String() string { return string(op) }

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
