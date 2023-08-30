// Package sas implements functionality for creating and verifying shared access
// signatures (SAS) and shared access tokens (SAT) as well as using them to
// query Axiom datasets. A SAS grants querying capabilities to a dataset for a
// given time range and with a global filter applied on behalf of an
// organization. A SAS is an URL query string composed of a set of query
// parameters that make up the payload for a signature and the cryptographic
// signature itself. That cryptographic signature is called SAT.
//
// Usage:
//
//	import "github.com/axiomhq/axiom-go/axiom/sas"
//
// To create a SAS string, that can be attached to a query request, use the
// high-level [Create] function. The returned string is an already url encoded
// query string.
//
// To create a SAT string for a set of values that make up a signature, use
// the low-level [CreateToken] function. The returned string is an already
// base64 encoded string.
package sas
