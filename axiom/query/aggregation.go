package query

// An AggregationOp can be applied on queries to aggrgate based on different
// conditions.
type AggregationOp string

// All available query aggregation operations.
const (
	// Works with all types, field should be `*`.
	OpCount         AggregationOp = "count"
	OpCountDistinct AggregationOp = "distinct"

	// Only works for numbers.
	OpSum         AggregationOp = "sum"
	OpAvg         AggregationOp = "avg"
	OpMin         AggregationOp = "min"
	OpMax         AggregationOp = "max"
	OpTopk        AggregationOp = "topk"
	OpPercentiles AggregationOp = "percentiles"
	OpHistogram   AggregationOp = "histogram"
)

// Aggregation performed as part of a query.
type Aggregation struct {
	// Op is the operation of the aggregation.
	Op AggregationOp `json:"op"`
	// Field the aggregation operation is performed on.
	Field string `json:"field"`
	// Argument to the aggregation. Only valid for OpTopk and OpPercentiles
	// aggregations.
	Argument interface{} `json:"argument"`
}
