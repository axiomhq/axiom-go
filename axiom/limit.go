package axiom

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=limitType,LimitScope -linecomment -output=limit_string.go

const (
	headerIngestLimit     = "X-IngestLimit-Limit"
	headerIngestRemaining = "X-IngestLimit-Remaining"
	headerIngestReset     = "X-IngestLimit-Reset"

	headerQueryLimit     = "X-QueryLimit-Limit"
	headerQueryRemaining = "X-QueryLimit-Remaining"
	headerQueryReset     = "X-QueryLimit-Reset"

	headerRateScope     = "X-RateLimit-Scope"
	headerRateLimit     = "X-RateLimit-Limit"
	headerRateRemaining = "X-RateLimit-Remaining"
	headerRateReset     = "X-RateLimit-Reset"

	// httpStatusLimitExceeded is a non-standard http status code returned by
	// Axiom to indicate that the query and/or ingest limit has been reached.
	httpStatusLimitExceeded = 430
)

type limitType uint8

const (
	limitIngest limitType = iota + 1 // ingest
	limitQuery                       // query
	limitRate                        // rate
)

// LimitScope is the scope of a [Limit].
type LimitScope uint8

// All available [Limit] scopes.
const (
	LimitScopeUnknown      LimitScope = iota // unknown
	LimitScopeUser                           // user
	LimitScopeOrganization                   // organization
	LimitScopeAnonymous                      // anonymous
)

func limitScopeFromString(s string) (ls LimitScope, err error) {
	switch s {
	case LimitScopeUnknown.String():
		ls = LimitScopeUnknown
	case LimitScopeUser.String():
		ls = LimitScopeUser
	case LimitScopeOrganization.String():
		ls = LimitScopeOrganization
	case LimitScopeAnonymous.String():
		ls = LimitScopeAnonymous
	default:
		err = fmt.Errorf("unknown limit scope %q", s)
	}
	return ls, err
}

// Limit represents a limit for the current client.
//
// Axiom API Reference: https://www.axiom.co/docs/restapi/api-limits
type Limit struct {
	// Scope a limit is enforced for. Only present on rate limited requests.
	Scope LimitScope
	// The maximum limit a client is limited to for a specified time range
	// which resets at the time indicated by [Limit.Reset].
	Limit uint64
	// The remaining count towards the maximum limit.
	Remaining uint64
	// The time at which the current limit time range will reset.
	Reset time.Time

	limitType limitType
}

// String returns a string representation of the limit.
//
// It implements [fmt.Stringer].
func (l Limit) String() string {
	return fmt.Sprintf("%d/%d %s limit remaining until %s", l.Remaining, l.Limit, l.limitType, l.Reset)
}

// parseLimit parses the limit related headers from a http response.
func parseLimit(r *http.Response) Limit {
	var limit Limit
	if hasHeaders(r, headerIngestLimit, headerIngestRemaining, headerIngestReset) {
		limit = parseLimitFromHeaders(r, "", headerIngestLimit, headerIngestRemaining, headerIngestReset)
		limit.limitType = limitIngest
	} else if hasHeaders(r, headerQueryLimit, headerQueryRemaining, headerQueryReset) {
		limit = parseLimitFromHeaders(r, "", headerQueryLimit, headerQueryRemaining, headerQueryReset)
		limit.limitType = limitQuery
	} else if hasHeaders(r, headerRateScope, headerRateLimit, headerRateRemaining, headerRateReset) {
		limit = parseLimitFromHeaders(r, headerRateScope, headerRateLimit, headerRateRemaining, headerRateReset)
		limit.limitType = limitRate
	}
	return limit
}

// parseLimitFromHeaders parses the named headers from a  http response.
func parseLimitFromHeaders(r *http.Response, headerScope, headerLimit, headerRemaining, headerReset string) Limit {
	var limit Limit
	if v := r.Header.Get(headerScope); v != "" {
		limit.Scope, _ = limitScopeFromString(v)
	}
	if v := r.Header.Get(headerLimit); v != "" {
		limit.Limit, _ = strconv.ParseUint(v, 10, 64)
	}
	if v := r.Header.Get(headerRemaining); v != "" {
		limit.Remaining, _ = strconv.ParseUint(v, 10, 64)
	}
	if v := r.Header.Get(headerReset); v != "" {
		if v, _ := strconv.ParseInt(v, 10, 64); v != 0 {
			limit.Reset = time.Unix(v, 0)
		}
	}
	return limit
}

// hasHeaders returns true if the response has all the given headers populated.
func hasHeaders(r *http.Response, headers ...string) bool {
	for _, header := range headers {
		if r.Header.Get(header) == "" {
			return false
		}
	}
	return true
}
