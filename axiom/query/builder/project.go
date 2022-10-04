package builder

import (
	"fmt"
	"strings"
)

// TODO: Add expression support.

// Project implements the [project operator].
//
// [project operator]: https://www.axiom.co/docs/apl/tabular-operators/project-operator
type Project struct {
	*Query

	cols []Column
}

// Project adds a [project operator] to the query. Subsequent calls will squash
// the project operators into a single project operator.
//
// Selects the columns to insert, rename, include or drop and embeds new
// computed columns.
//
// [project operator]: https://www.axiom.co/docs/apl/tabular-operators/project-operator
func (b *Project) Project(cols ...Column) *Project {
	b.cols = append(b.cols, cols...)
	return b
}

// build implements builder.
func (b *Project) build() (string, error) {
	var buf strings.Builder

	// TODO: validate cols

	_, _ = fmt.Fprintf(&buf, "project %s", b.cols[0])

	for i := 1; i < len(b.cols); i++ {
		_, _ = fmt.Fprintf(&buf, ", %s", b.cols[i])
	}

	return buf.String(), nil
}
