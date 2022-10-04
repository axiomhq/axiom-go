package builder

import (
	"fmt"
	"strings"
)

// Column represents a column in a dataset.
type Column string

// String makes sure that columns with special characters are quoted.
//
// It implements [fmt.Stringer].
func (c Column) String() string {
	s := string(c)
	if strings.Contains(s, ".") {
		return fmt.Sprintf("['%s']", s)
	}
	return s
}
