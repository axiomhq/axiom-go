package builder

import (
	"fmt"
	"strings"
)

// Extend implements the [extend operator].
//
// [extend operator]: https://www.axiom.co/docs/apl/tabular-operators/extend-operator
type Extend struct {
	*Query

	alias []string
	exprs []Expr
}

// Extend adds an [extend operator] to the query. Subsequent calls will squash
// the extend operators into a single extend operator.
//
// Creates calculated columns and appends them to the result set.
//
// [extend operator]: https://www.axiom.co/docs/apl/tabular-operators/extend-operator
func (b *Extend) Extend(alias string, expr Expr) *Extend {
	b.alias = append(b.alias, alias)
	b.exprs = append(b.exprs, expr)
	return b
}

// build implements builder.
func (b *Extend) build() (string, error) {
	var buf strings.Builder

	// TODO: validate alias and expr

	_, _ = fmt.Fprintf(&buf, "extend %s = %s", b.alias[0], b.exprs[0])

	for i := 1; i < len(b.alias); i++ {
		_, _ = fmt.Fprintf(&buf, ", %s = %s", b.alias[i], b.exprs[i])
	}

	return buf.String(), nil
}
