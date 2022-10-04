package builder

// Count implements the [count operator].
//
// [count operator]: https://www.axiom.co/docs/apl/tabular-operators/count-operator
type Count struct {
	*Query
}

// build implements builder.
func (b *Count) build() (string, error) {
	return "count", nil
}
