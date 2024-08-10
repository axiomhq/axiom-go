//go:build !go1.23

package query

import (
	"context"

	"github.com/axiomhq/axiom-go/axiom/query/iter"
)

// Rows returns an iterator over the rows build from the columns the table.
func (t Table) Rows() iter.Iter[Row] {
	return Rows(t.Columns)
}

// Values returns an iterator over the values of the column.
func (c Column) Values() iter.Iter[any] {
	return iter.Slice(c, func(_ context.Context, v any) (any, error) {
		return v, nil
	})
}
