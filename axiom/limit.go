package axiom

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=limitType,LimitScope -linecomment -output=limit_string.go

const (
	headerRateScope     = "X-RateLimit-Scope"
	headerRateLimit     = "X-RateLimit-Limit"
	headerRateRemaining = "X-RateLimit-Remaining"
	headerRateReset     = "X-RateLimit-Reset"

	headerQueryLimit     = "X-QueryLimit-Limit"
	headerQueryRemaining = "X-QueryLimit-Remaining"
	headerQueryReset     = "X-QueryLimit-Reset"

	headerIngestLimit     = "X-IngestLimit-Limit"
	headerIngestRemaining = "X-IngestLimit-Remaining"
	headerIngestReset     = "X-IngestLimit-Reset"
)

type limitType uint8

const (
	limitRate   limitType = iota + 1 // rate
	limitQuery                       // query
	limitIngest                      // ingest
)

// LimitScope is the scope of a limit.
type LimitScope uint8

// All available limit scopes.
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
	// The maximum limit a client is limited to for a specified time window
	// which resets at the time indicated by `Reset`.
	Limit int
	// The remaining count towards the maximum limit.
	Remaining int
	// The time at which the current limit time window will reset.
	Reset time.Time

	limitType limitType
}

// String returns a string representation of the rate.
//
// It implements `Stringer`.
func (l Limit) String() string {
	return fmt.Sprintf("%d/%d remaining until %s", l.Remaining, l.Limit, l.Reset)
}

// parseLimit parses the limit related headers from a `*http.Response`.
func parseLimit(r *http.Response) Limit {
	var limit Limit
	if strings.HasSuffix(r.Request.URL.Path, "/ingest") {
		limit = parseLimitFromHeaders(r, "", headerIngestLimit, headerIngestRemaining, headerIngestReset)
		limit.limitType = limitIngest
	} else if strings.HasSuffix(r.Request.URL.Path, "/query") || strings.HasSuffix(r.Request.URL.Path, "/_apl") {
		limit = parseLimitFromHeaders(r, "", headerQueryLimit, headerQueryRemaining, headerQueryReset)
		limit.limitType = limitQuery
	} else {
		limit = parseLimitFromHeaders(r, headerRateScope, headerRateLimit, headerRateRemaining, headerRateReset)
		limit.limitType = limitRate
	}
	return limit
}

// parseLimitFromHeaders parses the named headers from a `*http.Response`.
func parseLimitFromHeaders(r *http.Response, headerScope, headerLimit, headerRemaining, headerReset string) Limit {
	var limit Limit
	if v := r.Header.Get(headerScope); v != "" {
		limit.Scope, _ = limitScopeFromString(v)
	}
	if v := r.Header.Get(headerLimit); v != "" {
		limit.Limit, _ = strconv.Atoi(v)
	}
	if v := r.Header.Get(headerRemaining); v != "" {
		limit.Remaining, _ = strconv.Atoi(v)
	}
	if v := r.Header.Get(headerReset); v != "" {
		if v, _ := strconv.ParseInt(v, 10, 64); v != 0 {
			limit.Reset = time.Unix(v, 0)
		}
	}
	return limit
}

// limitKey returns a unique key for the limit type and scope combination.
func limitKey(limitType limitType, limitScope LimitScope) string {
	return fmt.Sprintf("%s:%s", limitType, limitScope)
}
