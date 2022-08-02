package axiom_test

import (
	"context"
	"fmt"
	"log"

	"github.com/axiomhq/axiom-go/axiom"
)

func Example() {
	// Export `AXIOM_TOKEN` and `AXIOM_ORG_ID` for Axiom Cloud.
	// Export `AXIOM_URL` and `AXIOM_TOKEN` for Axiom Selfhost.

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
