// Package axiom implements Go bindings for the Axiom API.
//
// Usage:
//
//	import "github.com/axiomhq/axiom-go/axiom"
//	import "github.com/axiomhq/axiom-go/axiom/ingest" // When ingesting data
//	import "github.com/axiomhq/axiom-go/axiom/otel" // When using OpenTelemetry
//	import "github.com/axiomhq/axiom-go/axiom/query" // When constructing APL queries
//	import "github.com/axiomhq/axiom-go/axiom/querylegacy" // When constructing legacy queries
//	import "github.com/axiomhq/axiom-go/axiom/sas" // When using shared access
//
// Construct a new Axiom client, then use the various services on the client to
// access different parts of the Axiom API. The package automatically takes its
// configuration from the environment if not specified otherwise. Refer to
// [NewClient] for details. The token can be an api or personal token. The api
// token however, will just allow ingestion or querying into or from the
// datasets the token is valid for, depending on its assigned permissions.
//
// To construct a client:
//
//	client, err := axiom.NewClient()
//
// or with [Option] functions:
//
//	client, err := axiom.NewClient(
//	    axiom.SetToken("..."),
//	    axiom.SetOrganizationID("..."),
//	)
//
// Get the current authenticated user:
//
//	user, err := client.Users.Current(ctx)
//
// NOTE: Every client method mapping to an API method takes a [context.Context]
// as its first parameter to pass cancellation signals and deadlines to
// requests. In case there is no context available, then [context.Background]
// can be used as a starting point.
//
// For more code samples, check out the [examples].
//
// [examples]: https://github.com/axiomhq/axiom-go/tree/main/examples
package axiom
