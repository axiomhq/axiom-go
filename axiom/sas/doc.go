// Package sas implements functionality for creating and verifying shared access
// signatures (SAS) and shared access tokens (SAT) as well as using them to
// query Axiom datasets. A SAS grants querying capabilities to a dataset for a
// given time range and with a global filter applied on behalf of an
// organization. A SAS is a URL query string composed of a set of query
// parameters that make up the payload for a signature and the cryptographic
// signature itself, the SAT.
//
// Usage:
//
//	import "github.com/axiomhq/axiom-go/axiom/sas"
//
// To create a SAS, that can be attached to a query request, use the
// high-level [Create] function. The returned [Options] can be attached to a
// [http.Request] or encoded to a query string by calling [Options.Encode].
package sas
