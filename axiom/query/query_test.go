package query_test

import (
	"fmt"

	"github.com/axiomhq/axiom-go/axiom/query"
)

func Example() {
	q, _ := query.Query("dataset").
		Distinct("status_code").
		Count().
		Build()

	// Output:
	// dataset
	// | distinct status_code
	// | count
	fmt.Println(q)
}
