package query

import (
	"context"

	"github.com/axiomhq/axiom-go/axiom/query/iter"
)

// Row represents a single row of a tabular query [Result].
type Row []any

// Values returns an iterator over the values of the row.
func (r Row) Values() iter.Iter[any] {
	return iter.Slice(r, func(_ context.Context, v any) (any, error) {
		return v, nil
	})
}

// Rows returns an iterator over the rows build from the columns of a tabular
// query [Result].
func Rows(columns []Column) iter.Iter[Row] {
	// Return an empty iterator if there are no columns or column values.
	if len(columns) == 0 || len(columns[0]) == 0 {
		return func(context.Context) (Row, error) {
			return nil, iter.Done
		}
	}

	return iter.Range(0, len(columns[0]), func(_ context.Context, idx int) (Row, error) {
		if idx >= len(columns[0]) {
			return nil, iter.Done
		}

		row := make(Row, len(columns))
		for columnIdx, column := range columns {
			row[columnIdx] = column[idx]
		}

		return row, nil
	})
}
