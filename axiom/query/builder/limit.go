package builder

import (
	"fmt"
)

// Limit implements the [limit operator].
//
// [limit operator]: https://www.axiom.co/docs/apl/tabular-operators/limit-operator
type Limit struct {
	*Query

	alias string
	limit uint64
}

// build implements builder.
func (b *Limit) build() (string, error) {
	return fmt.Sprintf("%s %d", b.alias, b.limit), nil
}
