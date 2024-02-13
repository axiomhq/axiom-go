package query_test

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/axiomhq/axiom-go/axiom/query"
	"github.com/axiomhq/axiom-go/axiom/query/iter"
)

func ExampleRows() {
	columns := []query.Column{
		[]any{
			"2020-11-19T11:06:31.569475746Z",
			"2020-11-19T11:06:31.569479846Z",
		},
		[]any{
			"Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)",
			"Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)",
		},
		[]any{
			"93.180.71.3",
			"93.180.71.3",
		},
		[]any{
			"GET /downloads/product_1 HTTP/1.1",
			"GET /downloads/product_1 HTTP/1.1",
		},
		[]any{
			304,
			304,
		},
	}

	rows := query.Rows(columns)
	buf := new(strings.Builder)

	for {
		if row, err := rows.Next(context.Background()); err == iter.Done {
			break
		} else if err != nil {
			log.Fatal(err)
		} else {
			fmt.Fprintln(buf, row)
		}
	}

	// Output:
	// [2020-11-19T11:06:31.569475746Z Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21) 93.180.71.3 GET /downloads/product_1 HTTP/1.1 304]
	// [2020-11-19T11:06:31.569479846Z Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21) 93.180.71.3 GET /downloads/product_1 HTTP/1.1 304]
	fmt.Print(buf.String())
}
