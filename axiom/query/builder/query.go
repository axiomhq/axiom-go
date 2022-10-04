package builder

import (
	"fmt"
	"strings"
)

// builder is the interface implemented by any query statement.
type builder interface {
	build() (string, error)
}

// Query implements a query builder for a named dataset.
type Query struct {
	dataset  Dataset
	options  []builder
	builders []builder
}

// New returns a new query builder for the given dataset.
func New(dataset string) *Query {
	return &Query{
		dataset:  Dataset(dataset),
		options:  make([]builder, 0),
		builders: make([]builder, 0),
	}
}

// Build the query.
func (q *Query) Build() (string, error) {
	var buf strings.Builder

	// TODO: validate dataset

	// Set query options first.
	for _, b := range q.options {
		s, err := b.build()
		if err != nil {
			return buf.String(), err
		}
		_, _ = fmt.Fprintln(&buf, s)
	}

	// Dataset marks the beginning of the query.
	_, _ = fmt.Fprintf(&buf, "%s", q.dataset)

	// Build each statement.
	for _, b := range q.builders {
		s, err := b.build()
		if err != nil {
			return buf.String(), err
		}
		_, _ = fmt.Fprintf(&buf, "\n| %s", s)
	}

	return buf.String(), nil
}

// ============================== Query Statements =============================

// Set adds a [set statement] to the query.
//
// The [set statement] is used to set a query option. Options enabled with the
// [set statement] only have effect for the duration of the query. It will
// affect how your query is processed and the results returned.
//
// [set query statement]: https://www.axiom.co/docs/apl/query-statement/set-statement
func (q *Query) Set(key, value string) *Set {
	b := &Set{
		Query: q,

		key:   key,
		value: value,
	}
	q.addOption(b)
	return b
}

// The following query statements (a.k.a. query options) are implemented using
// [query.Set], they don't need to call [Query.addOption]!

// SetStrictTypes adds a [stricttypes query option] to the query.
//
// The [stricttypes query option] is used to enable strict type checking for
// columns with mixed types.
//
// [stricttypes query option]: https://www.axiom.co/docs/apl/query-statement/set-statement#strict-types
func (q *Query) SetStrictTypes() *Set {
	return q.Set("stricttypes", "")
}

// ============================= Tabular Operators =============================

// Count adds a [count operator] to the query.
//
// Returns the number of events from the input dataset.
//
// [count operator]: https://www.axiom.co/docs/apl/tabular-operators/count-operator
func (q *Query) Count() *Count {
	b := &Count{
		Query: q,
	}
	q.addBuilder(b)
	return b
}

// Distinct adds a [distinct operator] to the query. Subsequent calls will
// squash the distinct operators into a single distinct operator.
//
// Produces a table with the distinct combination of the provided columns of the
// input table.
//
// [distinct operator]: https://www.axiom.co/docs/apl/tabular-operators/distinct-operator
func (q *Query) Distinct(cols ...Column) *Distinct {
	b := &Distinct{
		Query: q,

		cols: cols,
	}
	q.addBuilder(b)
	return b
}

// Extend adds an [extend operator] to the query. Subsequent calls will squash
// the extend operators into a single extend operator.
//
// Creates calculated columns and appends them to the result set.
//
// [extend operator]: https://www.axiom.co/docs/apl/tabular-operators/extend-operator
func (q *Query) Extend(alias string, expr Expr) *Extend {
	b := &Extend{
		Query: q,

		alias: []string{alias},
		exprs: []Expr{expr},
	}
	q.addBuilder(b)
	return b
}

// Limit adds a [limit operator] to the query. It is an alias for [Query.Take].
//
// Returns up to the specified number of rows.
//
// [limit operator]: https://www.axiom.co/docs/apl/tabular-operators/limit-operator
func (q *Query) Limit(v uint64) *Limit {
	b := &Limit{
		Query: q,

		alias: "limit",
		limit: v,
	}
	q.addBuilder(b)
	return b
}

// Order adds an [order operator] to the query. Subsequent calls will squash the
// order operators into a single order operator.
//
// Sorts the rows of the input dataset into order by column.
//
// [order operator]: https://www.axiom.co/docs/apl/tabular-operators/order-operator
func (q *Query) Order(col Column, ord SortOrder) *Order {
	b := &Order{
		Query: q,

		cols:       []Column{col},
		sortOrders: []SortOrder{ord},
	}
	q.addBuilder(b)
	return b
}

// OrderAsc adds an ascending [order operator] to the query. Subsequent calls
// will squash the order operators into a single order operator.
//
// Sorts the rows of the input dataset into ascending order by column.
//
// [order operator]: https://www.axiom.co/docs/apl/tabular-operators/order-operator
func (q *Query) OrderAsc(col Column) *Order {
	return q.Order(col, SortOrderAsc)
}

// OrderDesc adds an descending [order operator] to the query. Subsequent calls
// will squash the order operators into a single order operator.
//
// Sorts the rows of the input dataset into descending order by column.
//
// [order operator]: https://www.axiom.co/docs/apl/tabular-operators/order-operator
func (q *Query) OrderDesc(col Column) *Order {
	return q.Order(col, SortOrderDesc)
}

// Project adds a [project operator] to the query. Subsequent calls will squash
// the project operators into a single project operator.
//
// Selects the columns to insert, rename, include or drop and embeds new
// computed columns.
//
// [project operator]: https://www.axiom.co/docs/apl/tabular-operators/project-operator
func (q *Query) Project(cols ...Column) *Project {
	b := &Project{
		Query: q,

		cols: cols,
	}
	q.addBuilder(b)
	return b
}

// ProjectAway adds a [project-away operator] to the query. Subsequent calls
// will squash the project-away operators into a single project-away operator.
//
// Selects specified columns from the input to exclude from the output.
//
// [project-away operator]: https://www.axiom.co/docs/apl/tabular-operators/project-away-operator
func (q *Query) ProjectAway(cols ...Column) *ProjectAway {
	b := &ProjectAway{
		Query: q,

		cols: cols,
	}
	q.addBuilder(b)
	return b
}

// ProjectKeep adds a [project-keep operator] to the query. Subsequent calls
// will squash the project-keep operators into a single project-keep operator.
//
// Selects specified columns from the input to keep in the output.
//
// [project-keep operator]: https://www.axiom.co/docs/apl/tabular-operators/project-keep-operator
func (q *Query) ProjectKeep(cols ...Column) *ProjectKeep {
	b := &ProjectKeep{
		Query: q,

		cols: cols,
	}
	q.addBuilder(b)
	return b
}

// TODO(lukasmalkmus): Sort

// TODO(lukasmalkmus): Summarize

// Take adds a [take operator] to the query. It is an alias for [Query.Limit].
//
// Returns up to the specified number of rows.
//
// [take operator]: https://www.axiom.co/docs/apl/tabular-operators/take-operator
func (q *Query) Take(v uint64) *Limit {
	b := &Limit{
		Query: q,

		alias: "take",
		limit: v,
	}
	q.addBuilder(b)
	return b
}

// Top adds a [top operator] to the query.
//
// Returns the first N records grouped by the specified columns.
//
// [top operator]: https://www.axiom.co/docs/apl/tabular-operators/top-operator
func (q *Query) Top(n uint64, expr string) *Top {
	b := &Top{
		Query: q,

		n:    n,
		expr: expr,
	}
	q.addBuilder(b)
	return b
}

// Where adds a [where operator] to the query.
//
// Filters out a dataset to a branch of rows that meets a condition when
// executed.
//
// [where operator]: https://www.axiom.co/docs/apl/tabular-operators/where-operator
func (q *Query) Where(col Column, op FilterOp, expr string) *Where {
	b := &Where{
		Query: q,

		col:  col,
		op:   op,
		expr: expr,
		root: true,
	}
	q.addBuilder(b)
	return b
}

// WhereEq adds a [where operator] to the query that filters by "==". See
// [query.Where] for details.
//
// [where operator]: https://www.axiom.co/docs/apl/tabular-operators/where-operator
func (q *Query) WhereEq(col Column, expr string) *Where {
	return q.Where(col, FilterOpEqual, expr)
}

// WhereNEq adds a [where operator] to the query that filters by "!=". See
// [query.Where] for details.
//
// [where operator]: https://www.axiom.co/docs/apl/tabular-operators/where-operator
func (q *Query) WhereNEq(col Column, expr string) *Where {
	return q.Where(col, FilterOpNotEqual, expr)
}

// WhereGt adds a [where operator] to the query that filters by ">". See
// [query.Where] for details.
//
// [where operator]: https://www.axiom.co/docs/apl/tabular-operators/where-operator
func (q *Query) WhereGt(col Column, expr string) *Where {
	return q.Where(col, FilterOpGreaterThan, expr)
}

// WhereGtEq adds a [where operator] to the query that filters by ">=". See
// [query.Where] for details.
//
// [where operator]: https://www.axiom.co/docs/apl/tabular-operators/where-operator
func (q *Query) WhereGtEq(col Column, expr string) *Where {
	return q.Where(col, FilterOpGreaterThan, expr)
}

// WhereLt adds a [where operator] to the query that filters by "<". See
// [query.Where] for details.
//
// [where operator]: https://www.axiom.co/docs/apl/tabular-operators/where-operator
func (q *Query) WhereLt(col Column, expr string) *Where {
	return q.Where(col, FilterOpGreaterThan, expr)
}

// WhereLtEq adds a [where operator] to the query that filters by "<=". See
// [query.Where] for details.
//
// [where operator]: https://www.axiom.co/docs/apl/tabular-operators/where-operator
func (q *Query) WhereLtEq(col Column, expr string) *Where {
	return q.Where(col, FilterOpGreaterThan, expr)
}

// =============================== Helper methods ==============================

// addOption adds an option to the query builder.
func (q *Query) addOption(b builder) {
	q.options = append(q.options, b)
}

// addBuilder adds a builder to the query builder.
func (q *Query) addBuilder(b builder) {
	q.builders = append(q.builders, b)
}
