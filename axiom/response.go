package axiom

import "net/http"

// Response wraps the default http response type. It never has an open body.
type Response struct {
	*http.Response

	TraceID string

	Limit Limit
}

// newResponse creates a new response from the given http response.
func newResponse(r *http.Response) *Response {
	return &Response{
		Response: r,

		TraceID: r.Header.Get(headerTraceID),

		Limit: parseLimit(r),
	}
}
