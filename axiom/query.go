package axiom

// Query represents a query to be executed.
type Query interface {
	// Query is implemented by a query type to indicate that it can be used as a
	// query on Axiom. While possible, user implementations of `Query` are not
	// supported. Only supported types are `*apl.Query` and `*query.Query`.
	Query()
}
