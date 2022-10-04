package builder

import (
	"fmt"
	"strings"
)

// ProjectAway implements the [project-away operator].
//
// [project-away operator]: https://www.axiom.co/docs/apl/tabular-operators/project-away-operator
type ProjectAway struct {
	*Query

	cols []Column
}

// ProjectAway adds a [project-away operator] to the query. Subsequent calls
// will squash the project-away operators into a single project-away operator.
//
// Selects specified columns from the input to exclude from the output.
//
// [project-away operator]: https://www.axiom.co/docs/apl/tabular-operators/project-away-operator
func (b *ProjectAway) ProjectAway(cols ...Column) *ProjectAway {
	b.cols = append(b.cols, cols...)
	return b
}

// build implements builder.
func (b *ProjectAway) build() (string, error) {
	var buf strings.Builder

	// TODO: validate cols

	_, _ = fmt.Fprintf(&buf, "project-away %s", b.cols[0])

	for i := 1; i < len(b.cols); i++ {
		_, _ = fmt.Fprintf(&buf, ", %s", b.cols[i])
	}

	return buf.String(), nil
}
