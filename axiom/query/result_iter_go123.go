//go:build go1.23

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
