package axiom_test

import (
	"context"
	"fmt"
	"log"

	"github.com/axiomhq/axiom-go/axiom"
)

func Example() {
	client, err := axiom.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	user, err := client.Users.Current(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Hello %s!\n", user.Name)
}
