// Package sas implements functionality for creating and verifying shared access
// signatures (SAS) and shared access tokens (SAT). A SAS grants querying
// capabilities to a dataset for a given time window and with a global filter
// applied on behalf of an organization. A SAS is an URL query string composed
// of a set of query parameters that make up the payload for a signature and the
// signature itself. That signature is a SAT. Shared access is only supported on
// Axiom Cloud.
//
// To create a SAS string, that can be attached to a query request, use the
// high-level `Create()` function. The returned string is an already url-encoded
// query string.
//
// To create a SAT string for a sat of values that make up a signature, use
// the low-level `CreateToken()` function. The returned string is an already
// base64 url-encoded string.
//
// To verify a SAS string against a signing key use the `Verify()` function.
//
// To verify a SAT string against a signing key and a set of values use the
// `VerifyToken()` function.
package sas
