package builder_test

import (
	"fmt"

	. "github.com/axiomhq/axiom-go/axiom/query/builder"
)

func Example() {
	q, _ := New("dataset").
		SetStrictTypes().
		Extend("foo", "bar").
		Extend("fuz", "baz").
		WhereEq("fuz", "baz").
		AndNEq("foo", "bar").
		OrGt("1", "0").
		Project("foo", "bar").
		Project("baz").
		Distinct("foo").
		Distinct("bar", "baz").
		OrderAsc("foo").
		OrderDesc("bar").
		Top(10, "bar").
		Limit(1000).
		Take(100).
		Count().
		Build()

	// Output:
	// set stricttypes;
	// dataset
	// | extend foo = bar, fuz = baz
	// | where fuz == baz and foo != bar or 1 > 0
	// | project foo, bar, baz
	// | distinct foo, bar, baz
	// | order by foo asc, bar desc
	// | top 10 by bar
	// | limit 1000
	// | take 100
	// | count
	fmt.Println(q)
}
