package builder

import (
	"fmt"
	"strings"
)

// Set implements the [set query statement].
//
// [set query statement]: https://www.axiom.co/docs/apl/query-statement/set-statement
type Set struct {
	*Query

	key   string
	value string
}

// build implements builder.
func (b *Set) build() (string, error) {
	var buf strings.Builder

	// TODO: validate key (and value)

	_, _ = fmt.Fprintf(&buf, "set %s", b.key)
	if b.value != "" {
		_, _ = fmt.Fprintf(&buf, "=%s", b.value)
	}
	fmt.Fprint(&buf, ";")

	return buf.String(), nil
}
