package axiom_test

import (
	"github.com/axiomhq/axiom-go/axiom"
)

type is interface{ Is(error) bool }

var (
	_ error = (*axiom.HTTPError)(nil)
	_ error = (*axiom.LimitError)(nil)

	_ is = (*axiom.HTTPError)(nil)
	_ is = (*axiom.LimitError)(nil)
)
