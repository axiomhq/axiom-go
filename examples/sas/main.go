// The purpose of this example is to show how to create a shared access
// signature (SAS) for a dataset and use it to query that dataset via an
// ordinary HTTP request.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/query"
	"github.com/axiomhq/axiom-go/axiom/sas"
)

func main() {
	// Export "AXIOM_DATASET" in addition to the required environment variables.

	dataset := os.Getenv("AXIOM_DATASET")
	if dataset == "" {
		log.Fatal("AXIOM_DATASET is required")
	}

	signingKey := os.Getenv("AXIOM_SIGNING_KEY")
	if dataset == "" {
		log.Fatal("AXIOM_SIGNING_KEY is required")
	}

	// 1. Initialize the Axiom API client.
	client, err := axiom.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// 2. Ingest some events with different values for the "team_id" field.
	events := []axiom.Event{
		{"team_id": "a", "value": 1},
		{"team_id": "a", "value": 2},
		{"team_id": "b", "value": 4},
		{"team_id": "b", "value": 5},
	}
	ingestRes, err := client.IngestEvents(context.Background(), dataset, events)
	if err != nil {
		log.Fatal(err)
	} else if fails := len(ingestRes.Failures); fails > 0 {
		log.Fatalf("Ingestion of %d events failed", fails)
	}

	// 3. Create a shared access signature that limits query access to events
	// by the "team_id" field to only those with the value "a". The queries time
	// range is limited to the last 5 minutes.
	options, err := sas.Create(signingKey, sas.Params{
		OrganizationID: os.Getenv("AXIOM_ORG_ID"),
		Dataset:        dataset,
		Filter:         `team_id == "a"`,
		MinStartTime:   "ago(5m)",
		MaxEndTime:     "now",
		ExpiryTime:     "now",
	})
	if err != nil {
		log.Fatal(err)
	}

	// ❗From here on, assume the code is executed by a non-Axiom user that is
	// delegated query access via the SAS handed to him on behalf of the
	// organization.

	// 4. Construct the Axiom API URL for the APL query endpoint.
	u := os.Getenv("AXIOM_URL")
	if u == "" {
		u = "https://api.axiom.co"
	}
	queryURL, err := url.JoinPath(u, "/v1/datasets/_apl")
	if err != nil {
		log.Fatal(err)
	}
	queryURL += "?format=legacy" // Currently, must be set to "legacy".

	// 5. Construct the APL query request.
	r := fmt.Sprintf(`{
		"apl": "['%s'] | count",
		"startTime": "ago(1m)",
		"endTime": "now"
	}`, dataset)
	req, err := http.NewRequest(http.MethodPost, queryURL, strings.NewReader(r))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 6. Attach the SAS to the request.
	if err = options.Attach(req); err != nil {
		log.Fatal(err)
	}

	// 7. Execute the request.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// 8. Check the response status code.
	if code := resp.StatusCode; code != http.StatusOK {
		log.Fatalf("unexpected status code: %d (%s)", code, http.StatusText(code))
	}

	// 9. Decode the response.
	var res query.Result
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		log.Fatal(err)
	}

	// 10. Print the count, which should be "3".
	fmt.Println(res.Buckets.Totals[0].Aggregations[0].Value)
}
