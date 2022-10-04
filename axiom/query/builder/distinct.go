package builder

import (
	"fmt"
	"strings"
)

// Distinct implements the [distinct operator].
//
// [distinct operator]: https://www.axiom.co/docs/apl/tabular-operators/distinct-operator
type Distinct struct {
	*Query

	cols []Column
}

// Distinct adds a [distinct operator] to the query. Subsequent calls will
// squash the distinct operators into a single distinct operator.
//
// Produces a table with the distinct combination of the provided columns of the
// input table.
//
// [distinct operator]: https://www.axiom.co/docs/apl/tabular-operators/distinct-operator
func (b *Distinct) Distinct(cols ...Column) *Distinct {
	b.cols = append(b.cols, cols...)
	return b
}

// build implements builder.
func (b *Distinct) build() (string, error) {
	var buf strings.Builder

	// TODO: validate cols

	_, _ = fmt.Fprintf(&buf, "distinct %s", b.cols[0])

	for i := 1; i < len(b.cols); i++ {
		_, _ = fmt.Fprintf(&buf, ", %s", b.cols[i])
	}

	return buf.String(), nil
}
