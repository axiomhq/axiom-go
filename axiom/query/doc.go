// Package query provides the datatypes and functions for construction queries
// using the Axiom Processing Language (APL) and working with their results.
//
// Usage:
//
//	import "github.com/axiomhq/axiom-go/axiom/query"
//
// # Tabular Result Format
//
// Query results are returned in a tabular format. Each query [Result] contains
// one or more [Table]s. Each [Table] contains a list of [Field]s and a list of
// [Column]s. All [Column]s are equally sized and there are as much [Column]s as
// there are [Field]s.
//
// In case you want to work with events that are usually composed of multiple
// fields, you will find the values separated by [Column]. To aid with working
// with events in the tabular result format, the [Table] type provides the
// [Table.Rows] method that returns an [iter.Iter] over the [Row]s of the
// [Table]. Under the hood, each call to [iter.Iter.Next] composes a [Row] from
// the [Column]s of the [Table]. Alternatively, you can compose an [iter.Iter]
// over the [Row]s yourself using the [Rows] function. This allows for passing
// in a subset of the [Column]s of the [Table] to work with:
//
//	// Only build rows from the first two columns of the table. Returns an
//	// iterator for over the rows.
//	rows := query.Rows(result.Tables[0].Columns[0:2])
//
// Keep in mind that it is preferable to alter the APL query to only return the
// fields you are interested in instead of working with a subset of the columns
// after the query has been executed.
package query
