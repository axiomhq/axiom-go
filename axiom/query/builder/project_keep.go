package builder

import (
	"fmt"
	"strings"
)

// ProjectKeep implements the [project-keep operator].
//
// [project-keep operator]: https://www.axiom.co/docs/apl/tabular-operators/project-keep-operator
type ProjectKeep struct {
	*Query

	cols []Column
}

// ProjectKeep adds a [project-keep operator] to the query. Subsequent calls
// will squash the project-keep operators into a single project-keep operator.
//
// Selects specified columns from the input to keep in the output.
//
// [project-keep operator]: https://www.axiom.co/docs/apl/tabular-operators/project-keep-operator
func (b *ProjectKeep) ProjectKeep(cols ...Column) *ProjectKeep {
	b.cols = append(b.cols, cols...)
	return b
}

// build implements builder.
func (b *ProjectKeep) build() (string, error) {
	var buf strings.Builder

	// TODO: validate cols

	_, _ = fmt.Fprintf(&buf, "project-keep %s", b.cols[0])

	for i := 1; i < len(b.cols); i++ {
		_, _ = fmt.Fprintf(&buf, ", %s", b.cols[i])
	}

	return buf.String(), nil
}
