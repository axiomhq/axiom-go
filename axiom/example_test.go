package axiom_test

import (
	"context"
	"fmt"
	"log"

	"github.com/axiomhq/axiom-go/axiom"
)

func Example() {
	client, err := axiom.NewClient("https://my-axiom.example.com", "<ACCESS-TOKEN>", "<ORGANIZATION-ID>")
	if err != nil {
		log.Fatal(err)
	}

	version, err := client.Version.Get(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(version)
}
