package builder

import (
	"fmt"
	"strings"
)

// Dataset represents a dataset.
type Dataset string

// String makes sure that datasets with special characters are quoted.
//
// It implements [fmt.Stringer].
func (d Dataset) String() string {
	s := string(d)
	if strings.ContainsAny(s, "-_.") {
		return fmt.Sprintf("['%s']", s)
	}
	return s
}
