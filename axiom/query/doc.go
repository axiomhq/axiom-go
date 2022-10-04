// Package query provides the datatypes and functions for construction queries
// using the Axiom Processing Language (APL) and working with their results.
//
// Usage:
//
//	import "github.com/axiomhq/axiom-go/axiom/query"

// The base for every APL query is the [Query] type which return a new query
// builder:
//
//	q := query.Query("dataset").
//		// ...
//		Build()
package query
