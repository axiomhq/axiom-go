package builder

import (
	"fmt"
)

// Top implements the [top operator].
//
// [top operator]: https://www.axiom.co/docs/apl/tabular-operators/top-operator
type Top struct {
	*Query

	n    uint64
	expr string
}

// build implements builder.
func (b *Top) build() (string, error) {
	return fmt.Sprintf("top %d by %s", b.n, b.expr), nil
}
