package axiom

import "net/http"

// response wraps the default `http.Response` type. It never has an open body.
type response struct {
	*http.Response

	Limit Limit
}

// newResponse creates a new `response` from the given `*http.Response`.
func newResponse(r *http.Response) *response {
	return &response{
		Response: r,

		Limit: parseLimit(r),
	}
}
