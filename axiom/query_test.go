package axiom_test

import (
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/apl"
	"github.com/axiomhq/axiom-go/axiom/query"
)

var (
	_ axiom.Query = apl.Query("")
	_ axiom.Query = query.Query{}
)
