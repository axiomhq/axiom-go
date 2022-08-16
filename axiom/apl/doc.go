// Package apl provides the datatypes and functions for construction queries
// using the Axiom Processing Language (APL) and working with their results.
// They usually extend the functionality of existing types from the `query`
// package.
//
// The base for every APL query is the `Query` type which wraps an APL query
// string:
//
//	q := apl.Query("...")
package apl
