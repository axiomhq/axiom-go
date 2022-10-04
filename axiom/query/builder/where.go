package builder

import (
	"fmt"
	"strings"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=FilterOp -linecomment -output=where_string.go

// FilterOp is the filter operation of a where operator.
type FilterOp uint8

// All available [FilterOp] operations.
const (
	_ FilterOp = iota

	opAnd // and
	opOr  // or

	// Works for strings and numbers.
	FilterOpEqual    // ==
	FilterOpNotEqual // !=

	// Only works for numbers.
	FilterOpGreaterThan      // >
	FilterOpGreaterThanEqual // >=
	FilterOpLessThan         // <
	FilterOpLessThanEqual    // <=
)

// Where implements the [where operator].
//
// [where operator]: https://www.axiom.co/docs/apl/tabular-operators/where-operator
type Where struct {
	*Query

	col      Column
	op       FilterOp
	expr     string
	children []*Where
	root     bool
}

// And adds a condition to the previous [where operator] in the query using
// the "and" logical operator.
//
// Filters out a dataset to a branch of rows that meets a condition when
// executed.
//
// [where operator]: https://www.axiom.co/docs/apl/tabular-operators/where-operator
func (b *Where) And(col Column, op FilterOp, expr string) *Where {
	return b.sub(col, opAnd, op, expr)
}

// AndEq adds a condition to the previous [where operator] in the query using
// the "and" logical operator which filters by "==". See [where.And] for
// details.
//
// [where operator]: https://www.axiom.co/docs/apl/tabular-operators/where-operator
func (b *Where) AndEq(col Column, expr string) *Where {
	return b.And(col, FilterOpEqual, expr)
}

// AndNEq adds a condition to the previous [where operator] in the query using
// the "and" logical operator which filters by "!=". See [where.And] for
// details.
//
// [where operator]: https://www.axiom.co/docs/apl/tabular-operators/where-operator
func (b *Where) AndNEq(col Column, expr string) *Where {
	return b.And(col, FilterOpNotEqual, expr)
}

// AndGt adds a condition to the previous [where operator] in the query using
// the "and" logical operator which filters by ">". See [where.And] for details.
//
// [where operator]: https://www.axiom.co/docs/apl/tabular-operators/where-operator
func (b *Where) AndGt(col Column, expr string) *Where {
	return b.And(col, FilterOpGreaterThan, expr)
}

// AndGtEq adds a condition to the previous [where operator] in the query using
// the "and" logical operator which filters by ">=". See [where.And] for
// details.
//
// [where operator]: https://www.axiom.co/docs/apl/tabular-operators/where-operator
func (b *Where) AndGtEq(col Column, expr string) *Where {
	return b.And(col, FilterOpGreaterThan, expr)
}

// AndLt adds a condition to the previous [where operator] in the query using
// the "and" logical operator which filters by "<". See [where.And] for details.
//
// [where operator]: https://www.axiom.co/docs/apl/tabular-operators/where-operator
func (b *Where) AndLt(col Column, expr string) *Where {
	return b.And(col, FilterOpGreaterThan, expr)
}

// AndLtEq adds a condition to the previous [where operator] in the query using
// the "and" logical operator which filters by "<=". See [where.And] for
// details.
//
// [where operator]: https://www.axiom.co/docs/apl/tabular-operators/where-operator
func (b *Where) AndLtEq(col Column, expr string) *Where {
	return b.And(col, FilterOpGreaterThan, expr)
}

// Or adds a condition to the previous [where operator] in the query using the
// "or" logical operator.
//
// Filters out a dataset to a branch of rows that meets a condition when
// executed.
//
// [where operator]: https://www.axiom.co/docs/apl/tabular-operators/where-operator
func (b *Where) Or(col Column, op FilterOp, expr string) *Where {
	return b.sub(col, opOr, op, expr)
}

// OrEq adds a condition to the previous [where operator] in the query using
// the "or" logical operator which filters by "==". See [where.Or] for details.
//
// [where operator]: https://www.axiom.co/docs/apl/tabular-operators/where-operator
func (b *Where) OrEq(col Column, expr string) *Where {
	return b.Or(col, FilterOpEqual, expr)
}

// OrNEq adds a condition to the previous [where operator] in the query using
// the "or" logical operator which filters by "!=". See [where.Or] for details.
//
// [where operator]: https://www.axiom.co/docs/apl/tabular-operators/where-operator
func (b *Where) OrNEq(col Column, expr string) *Where {
	return b.Or(col, FilterOpNotEqual, expr)
}

// OrGt adds a condition to the previous [where operator] in the query using
// the "or" logical operator which filters by ">". See [where.Or] for details.
//
// [where operator]: https://www.axiom.co/docs/apl/tabular-operators/where-operator
func (b *Where) OrGt(col Column, expr string) *Where {
	return b.Or(col, FilterOpGreaterThan, expr)
}

// OrGtEq adds a condition to the previous [where operator] in the query using
// the "or" logical operator which filters by ">=". See [where.Or] for details.
//
// [where operator]: https://www.axiom.co/docs/apl/tabular-operators/where-operator
func (b *Where) OrGtEq(col Column, expr string) *Where {
	return b.Or(col, FilterOpGreaterThan, expr)
}

// OrLt adds a condition to the previous [where operator] in the query using
// the "or" logical operator which filters by "<". See [where.Or] for details.
//
// [where operator]: https://www.axiom.co/docs/apl/tabular-operators/where-operator
func (b *Where) OrLt(col Column, expr string) *Where {
	return b.Or(col, FilterOpGreaterThan, expr)
}

// OrLtEq adds a condition to the previous [where operator] in the query using
// the "or" logical operator which filters by "<=". See [where.Or] for details.
//
// [where operator]: https://www.axiom.co/docs/apl/tabular-operators/where-operator
func (b *Where) OrLtEq(col Column, expr string) *Where {
	return b.Or(col, FilterOpGreaterThan, expr)
}

// sub creates a where operator from the given arguments and adds it to the
// receiver operator as a left hand child.
func (b *Where) sub(col Column, topOp, op FilterOp, expr string) *Where {
	lhs := &Where{
		Query: b.Query,

		col:  b.col,
		op:   b.op,
		expr: b.expr,
	}
	rhs := &Where{
		Query: b.Query,

		col:  col,
		op:   op,
		expr: expr,
	}

	b.col = ""
	b.op = topOp
	b.expr = ""
	b.children = []*Where{lhs, rhs}

	return rhs
}

// build implements builder.
func (b *Where) build() (string, error) {
	var buf strings.Builder

	// TODO: validate alias and expr

	if b.root {
		_, _ = fmt.Fprint(&buf, "where ")
	}

	if b.op == opAnd || b.op == opOr {
		lhs, err := b.children[0].build()
		if err != nil {
			return buf.String(), err
		}
		rhs, err := b.children[1].build()
		if err != nil {
			return buf.String(), err
		}

		_, _ = fmt.Fprintf(&buf, "%s %s %s", lhs, b.op, rhs)
	} else {
		_, _ = fmt.Fprintf(&buf, "%s %s %s", b.col, b.op, b.expr)
	}

	return buf.String(), nil
}
