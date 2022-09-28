package axiom

import "net/http"

// Response wraps the default `http.Response` type. It never has an open body.
type Response struct {
	*http.Response

	Limit Limit
}

// newResponse creates a new `response` from the given `*http.Response`.
func newResponse(r *http.Response) *Response {
	return &Response{
		Response: r,

		Limit: parseLimit(r),
	}
}
