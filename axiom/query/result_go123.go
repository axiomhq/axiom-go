//go:build go1.23

// FIXME(lukasmalkmus): Once Go 1.24 is released, remove this file and move the
// Rows and Values methods to result.go.

package query

import "iter"

// Rows returns an iterator over the rows build from the columns the table.
func (t Table) Rows() iter.Seq[Row] {
	return Rows(t.Columns)
}

// Values returns an iterator over the values of the column.
func (c Column) Values() iter.Seq[any] {
	return func(yield func(any) bool) {
		for _, v := range c {
			if !yield(v) {
				return
			}
		}
	}
}
