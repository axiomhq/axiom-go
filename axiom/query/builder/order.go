package builder

import (
	"fmt"
	"strings"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=SortOrder -linecomment -output=order_string.go

// SortOrder is the sort order of a column.
type SortOrder uint8

// All available sort [Order] directions.
const (
	_ SortOrder = iota

	SortOrderAsc  // asc
	SortOrderDesc // desc
)

// Order implements the [order operator].
//
// [order operator]: https://www.axiom.co/docs/apl/tabular-operators/order-operator
type Order struct {
	*Query

	cols       []Column
	sortOrders []SortOrder
}

// Order adds an [order operator] to the query. Subsequent calls will squash the
// order operators into a single order operator.
//
// Sorts the rows of the input dataset into order by column.
//
// [order operator]: https://www.axiom.co/docs/apl/tabular-operators/order-operator
func (b *Order) Order(col Column, ord SortOrder) *Order {
	b.cols = append(b.cols, col)
	b.sortOrders = append(b.sortOrders, ord)
	return b
}

// OrderAsc adds an ascending [order operator] to the query. Subsequent calls
// will squash the order operators into a single order operator.
//
// Sorts the rows of the input dataset into ascending order by column.
//
// [order operator]: https://www.axiom.co/docs/apl/tabular-operators/order-operator
func (b *Order) OrderAsc(col Column) *Order {
	return b.Order(col, SortOrderAsc)
}

// OrderDesc adds an descending [order operator] to the query. Subsequent calls
// will squash the order operators into a single order operator.
//
// Sorts the rows of the input dataset into descending order by column.
//
// [order operator]: https://www.axiom.co/docs/apl/tabular-operators/order-operator
func (b *Order) OrderDesc(col Column) *Order {
	return b.Order(col, SortOrderDesc)
}

// build implements builder.
func (b *Order) build() (string, error) {
	var buf strings.Builder

	// TODO: validate cols

	_, _ = fmt.Fprintf(&buf, "order by %s %s", b.cols[0], b.sortOrders[0])

	for i := 1; i < len(b.cols); i++ {
		_, _ = fmt.Fprintf(&buf, ", %s %s", b.cols[i], b.sortOrders[i])
	}

	return buf.String(), nil
}
