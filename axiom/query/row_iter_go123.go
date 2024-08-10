//go:build go1.23

package query

import (
	"iter"
)

// Values returns an iterator over the values of the row.
func (r Row) Values() iter.Seq[any] {
	return func(yield func(any) bool) {
		for _, v := range r {
			if !yield(v) {
				return
			}
		}
	}
}

// Rows returns an iterator over the rows build from the columns of a tabular
// query [Result].
func Rows(columns []Column) iter.Seq[Row] {
	// Return an empty iterator if there are no columns or column values.
	if len(columns) == 0 || len(columns[0]) == 0 {
		return func(func(Row) bool) {}
	}

	return func(yield func(Row) bool) {
		for i := range columns[0] {
			row := make(Row, len(columns))
			for j, column := range columns {
				row[j] = column[i]
			}
			if !yield(row) {
				return
			}
		}
	}
}
