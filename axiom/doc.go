// Package axiom implements Go bindings for the Axiom API.
//
// Usage:
//
// 	import "github.com/axiomhq/axiom-go/axiom"
// 	import "github.com/axiomhq/axiom-go/axiom/query" // When constructing queries
//
// Construct a new Axiom client, then use the various services on the client to
// access different parts of the Axiom API. A valid deployment URL and an access
// token must be passed. The access token can be a personal one or an ingest
// token. The ingest token however, will just allow ingestion into the datasets
// targeted by the ingest token:
//
//	client, err := axiom.NewClient("https://my-axiom.example.com", "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX")
//
// 	Get the version of the configured deployment:
//	version, err := client.Version.Get(ctx)
//
// Some API methods have additional parameters that can be passed:
//
//	dashboards, err := client.Dashboards.List(ctx, axiom.ListOptions{
//		Limit: 5,
//	})
//
// NOTE: Every client method mapping to an API method takes a context.Context
// (https://godoc.org/context) as its first parameter to pass cancelation
// signals and deadlines to requests. In case there is no context available,
// then context.Background() can be used as a starting point.
//
// For more code samples, check out the https://github.com/axiomhq/axiom-go/tree/master/example directory.
package axiom
