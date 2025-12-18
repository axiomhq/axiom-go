package axiom

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/klauspost/compress/gzhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/axiomhq/axiom-go/axiom/ingest"
	"github.com/axiomhq/axiom-go/axiom/query"
	"github.com/axiomhq/axiom-go/axiom/querylegacy"
	"github.com/axiomhq/axiom-go/internal/config"
	"github.com/axiomhq/axiom-go/internal/version"
)

const (
	headerAuthorization  = "Authorization"
	headerOrganizationID = "X-Axiom-Org-Id"
	headerEventLabels    = "X-Axiom-Event-Labels"

	headerAccept      = "Accept"
	headerContentType = "Content-Type"
	headerUserAgent   = "User-Agent"

	headerTraceID = "X-Axiom-Trace-Id"

	defaultMediaType = "application/octet-stream"
	mediaTypeJSON    = "application/json"
	mediaTypeNDJSON  = "application/x-ndjson"

	otelTracerName = "github.com/axiomhq/axiom-go/axiom"
)

// service is the base service used by all Axiom API services.
type service struct {
	client   *Client
	basePath string
}

// DefaultHTTPClient returns the default [http.Client] used for making requests.
func DefaultHTTPClient() *http.Client {
	return &http.Client{
		Transport: DefaultHTTPTransport(),
		Timeout:   time.Minute * 5,
	}
}

// DefaultHTTPTransport returns the default [http.Client.Transport] used by
// [DefaultHTTPClient].
func DefaultHTTPTransport() http.RoundTripper {
	return otelhttp.NewTransport(gzhttp.Transport(&http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   time.Second * 30,
			KeepAlive: time.Second * 30,
		}).DialContext,
		IdleConnTimeout:       time.Minute,
		ResponseHeaderTimeout: time.Minute * 2,
		TLSHandshakeTimeout:   time.Second * 10,
		ExpectContinueTimeout: time.Second * 1,
		ForceAttemptHTTP2:     true,
	}))
}

// Client provides the Axiom HTTP API operations.
type Client struct {
	config config.Config

	httpClient *http.Client
	userAgent  string
	noEnv      bool
	noRetry    bool

	strictDecoding bool

	tracer trace.Tracer

	// Services for communicating with different parts of the Axiom API.
	Datasets      *DatasetsService
	Organizations *OrganizationsService
	Users         *UsersService
	Monitors      *MonitorsService
	Notifiers     *NotifiersService
	Annotations   *AnnotationsService
	Tokens        *TokensService
	VirtualFields *VirtualFieldsService
}

// NewClient returns a new Axiom API client. It automatically takes its
// configuration from the environment. To connect, export the following
// environment variables:
//
//   - AXIOM_TOKEN
//   - AXIOM_ORG_ID (only when using a personal token)
//
// The configuration can be set manually using options which are prefixed with
// "Set".
//
// The token must be an api or personal token which can be created on the
// settings or user profile page on Axiom.
func NewClient(options ...Option) (*Client, error) {
	client := &Client{
		config: config.Default(),

		httpClient: DefaultHTTPClient(),

		userAgent: "axiom-go",

		tracer: otel.Tracer(otelTracerName),
	}

	// Include module version in the user agent.
	if v := version.Get(); v != "" {
		client.userAgent += fmt.Sprintf("/%s", v)
	}

	client.Datasets = &DatasetsService{client: client, basePath: "/v2/datasets"}
	client.Organizations = &OrganizationsService{client: client, basePath: "/v2/orgs"}
	client.Users = &UsersService{client: client, basePath: "/v2/users"}
	client.Monitors = &MonitorsService{client: client, basePath: "/v2/monitors"}
	client.Notifiers = &NotifiersService{client: client, basePath: "/v2/notifiers"}
	client.Annotations = &AnnotationsService{client: client, basePath: "/v2/annotations"}
	client.Tokens = &TokensService{client: client, basePath: "/v2/tokens"}
	client.VirtualFields = &VirtualFieldsService{client: client, basePath: "/v2/vfields"}

	// Apply supplied options.
	if err := client.Options(options...); err != nil {
		return nil, err
	}

	// Make sure to populate remaining fields from the environment, if not
	// explicitly disabled.
	if !client.noEnv {
		if err := client.config.IncorporateEnvironment(); err != nil {
			return nil, err
		}
	}

	return client, client.config.Validate()
}

// Options applies options to the client.
func (c *Client) Options(options ...Option) error {
	for _, option := range options {
		if option == nil {
			continue
		} else if err := option(c); err != nil {
			return err
		}
	}
	return nil
}

// ValidateCredentials makes sure the client can properly authenticate against
// the configured Axiom API.
func (c *Client) ValidateCredentials(ctx context.Context) error {
	if config.IsPersonalToken(c.config.Token()) {
		_, err := c.Users.Current(ctx)
		return err
	}

	// FIXME(lukasmalkmus): Well, with the current API, we need to assume the
	// API token is valid as there is no endpoint to check it.
	// return ErrInvalidToken
	return nil
}

// Call creates a new API request and executes it. The response body is JSON
// decoded or directly written to v, depending on v being an [io.Writer] or not.
func (c *Client) Call(ctx context.Context, method, path string, body, v any) error {
	if req, err := c.NewRequest(ctx, method, path, body); err != nil {
		return err
	} else if _, err = c.Do(req, v); err != nil {
		return err
	}
	return nil
}

// NewRequest creates an API request. If specified, the value pointed to by body
// will be included as the request body. If it is not an [io.Reader], it will be
// included as a JSON encoded request body.
func (c *Client) NewRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	rel, err := url.ParseRequestURI(path)
	if err != nil {
		return nil, err
	}
	endpoint := c.config.BaseURL().ResolveReference(rel)

	var (
		r        io.Reader
		isReader bool
	)
	if body != nil {
		if r, isReader = body.(io.Reader); !isReader {
			buf := new(bytes.Buffer)
			if err = json.NewEncoder(buf).Encode(body); err != nil {
				return nil, err
			}
			r = buf
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), r)
	if err != nil {
		return nil, err
	}

	// Set Content-Type.
	if body != nil && !isReader {
		req.Header.Set(headerContentType, mediaTypeJSON)
	} else if body != nil {
		req.Header.Set(headerContentType, defaultMediaType)
	}

	// Set authorization header, if present.
	if c.config.Token() != "" {
		req.Header.Set(headerAuthorization, "Bearer "+c.config.Token())
	}

	// Set organization ID header when using a personal token.
	if config.IsPersonalToken(c.config.Token()) && c.config.OrganizationID() != "" {
		req.Header.Set(headerOrganizationID, c.config.OrganizationID())
	}

	// Set other headers.
	req.Header.Set(headerAccept, mediaTypeJSON)
	req.Header.Set(headerUserAgent, c.userAgent)

	return req, nil
}

// Do sends an API request and returns the API response. The response body is
// JSON decoded or directly written to v, depending on v being an [io.Writer] or
// not.
func (c *Client) Do(req *http.Request, v any) (*Response, error) {
	var (
		resp *Response
		err  error
	)
	if req.GetBody != nil && !c.noRetry {
		bck := backoff.NewExponentialBackOff()
		bck.InitialInterval = time.Millisecond * 200
		bck.MaxElapsedTime = time.Second * 10
		bck.Multiplier = 2.0

		err = backoff.Retry(func() error {
			var httpResp *http.Response
			//nolint:bodyclose // The response body is closed later down below.
			httpResp, err = c.httpClient.Do(req)
			switch {
			case errors.Is(err, context.Canceled):
				return backoff.Permanent(err)
			case err != nil:
				return err
			}
			resp = newResponse(httpResp)

			// We should only retry in the case the status code is >= 500,
			// anything below isn't worth retrying.
			if code := resp.StatusCode; code >= 500 {
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()

				// Reset the requests body, so it can be re-read.
				if req.Body, err = req.GetBody(); err != nil {
					return backoff.Permanent(err)
				}

				return fmt.Errorf("got status code %d", code)
			}

			return nil
		}, bck)
	} else {
		var httpResp *http.Response
		//nolint:bodyclose // The response body is closed later down below.
		if httpResp, err = c.httpClient.Do(req); err != nil {
			return nil, err
		}
		resp = newResponse(httpResp)
	}

	defer func() {
		if resp != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}
	}()

	if err != nil {
		return resp, err
	}

	span := trace.SpanFromContext(req.Context())
	if span.IsRecording() && resp.TraceID() != "" {
		span.SetAttributes(attribute.String("axiom_trace_id", resp.TraceID()))
	}

	if statusCode := resp.StatusCode; statusCode >= http.StatusBadRequest {
		httpErr := HTTPError{
			Status:  statusCode,
			Message: http.StatusText(statusCode),
			TraceID: resp.TraceID(),
		}

		if span.IsRecording() {
			span.SetAttributes(semconv.HTTPResponseStatusCode(statusCode))
		}

		// Handle a generic HTTP error if the response is not JSON formatted.
		if ct, _, _ := mime.ParseMediaType(resp.Header.Get(headerContentType)); ct != mediaTypeJSON {
			return resp, httpErr
		}

		// For error handling, we want to have access to the raw request body
		// to inspect it further
		var (
			buf bytes.Buffer
			dec = json.NewDecoder(io.TeeReader(resp.Body, &buf))
		)

		// Handle a properly JSON formatted Axiom API error response.
		if err = dec.Decode(&httpErr); err != nil {
			return resp, fmt.Errorf("error decoding %d error response: %w", statusCode, err)
		}

		// In case something went wrong, include the raw response and hope for
		// the best.
		if httpErr.Message == "" {
			s := strings.ReplaceAll(buf.String(), "\n", " ")
			httpErr.Message = s
		}

		// Handle special error types.
		switch statusCode {
		case http.StatusTooManyRequests, httpStatusLimitExceeded:
			return resp, LimitError{
				HTTPError: httpErr,

				Limit: resp.Limit,
			}
		}

		return resp, httpErr
	}

	if span.IsRecording() {
		span.SetAttributes(semconv.HTTPResponseBodySize(int(resp.ContentLength)))
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
			return resp, err
		}

		if ct, _, _ := mime.ParseMediaType(resp.Header.Get(headerContentType)); ct != mediaTypeJSON {
			return resp, fmt.Errorf("cannot decode response with unsupported content type %q", ct)
		}

		dec := json.NewDecoder(resp.Body)
		if c.strictDecoding {
			dec.DisallowUnknownFields()
		}
		return resp, dec.Decode(v)
	}

	return resp, nil
}

// Ingest data into the dataset identified by its id.
//
// The timestamp of the events will be set by the server to the current server
// time if the "_time" field is not set. The server can be instructed to use a
// different field as the timestamp by setting the [ingest.SetTimestampField]
// option. If not explicitly specified by [ingest.SetTimestampFormat], the
// timestamp format is auto detected.
//
// Restrictions for field names (JSON object keys) can be reviewed in
// [our documentation].
//
// The reader is streamed to the server until EOF is reached on a single
// connection. Keep that in mind when dealing with slow readers.
//
// This function is an alias to [DatasetsService.Ingest].
//
// [our documentation]: https://www.axiom.co/docs/usage/field-restrictions
func (c *Client) Ingest(ctx context.Context, id string, r io.Reader, typ ContentType, enc ContentEncoding, options ...ingest.Option) (*ingest.Status, error) {
	return c.Datasets.Ingest(ctx, id, r, typ, enc, options...)
}

// IngestEvents ingests events into the dataset identified by its id.
//
// The timestamp of the events will be set by the server to the current server
// time if the "_time" field is not set. The server can be instructed to use a
// different field as the timestamp by setting the [ingest.SetTimestampField]
// option. If not explicitly specified by [ingest.SetTimestampFormat], the
// timestamp format is auto detected.
//
// Restrictions for field names (JSON object keys) can be reviewed in
// [our documentation].
//
// For ingesting large amounts of data, consider using the [Client.Ingest] or
// [Client.IngestChannel] method.
//
// This function is an alias to [DatasetsService.IngestEvents].
//
// [our documentation]: https://www.axiom.co/docs/usage/field-restrictions
func (c *Client) IngestEvents(ctx context.Context, id string, events []Event, options ...ingest.Option) (*ingest.Status, error) {
	return c.Datasets.IngestEvents(ctx, id, events, options...)
}

// IngestChannel ingests events from a channel into the dataset identified by
// its id.
//
// The timestamp of the events will be set by the server to the current server
// time if the "_time" field is not set. The server can be instructed to use a
// different field as the timestamp by setting the [ingest.SetTimestampField]
// option. If not explicitly specified by [ingest.SetTimestampFormat], the
// timestamp format is auto detected.
//
// Restrictions for field names (JSON object keys) can be reviewed in
// [our documentation].
//
// Events are ingested in batches. A batch is either 10000 events for unbuffered
// channels or the capacity of the channel for buffered channels. The maximum
// batch size is 10000. A batch is sent to the server as soon as it is full,
// after one second or when the channel is closed.
//
// The method returns with an error when the context is marked as done or an
// error occurs when sending the events to the server. A partial ingestion is
// possible and the returned ingest status is valid to use. When the context is
// marked as done, no attempt is made to send the buffered events to the server.
//
// The method returns without an error if the channel is closed and the buffered
// events are successfully sent to the server.
//
// This function is an alias to [DatasetsService.IngestChannel].
//
// [our documentation]: https://www.axiom.co/docs/usage/field-restrictions
func (c *Client) IngestChannel(ctx context.Context, id string, events <-chan Event, options ...ingest.Option) (*ingest.Status, error) {
	return c.Datasets.IngestChannel(ctx, id, events, options...)
}

// Query executes the given query specified using the Axiom Processing
// Language (APL).
//
// To learn more about APL, please refer to [our documentation].
//
// This function is an alias to [DatasetsService.Query].
//
// [our documentation]: https://www.axiom.co/docs/apl/introduction
func (c *Client) Query(ctx context.Context, apl string, options ...query.Option) (*query.Result, error) {
	return c.Datasets.Query(ctx, apl, options...)
}

// QueryLegacy executes the given legacy query on the dataset identified by its
// id.
//
// This function is an alias to [DatasetsService.Query].
//
// Deprecated: Legacy queries will be replaced by queries specified using the
// Axiom Processing Language (APL) and the legacy query API will be removed in
// the future. Use [Client.Query] instead.
func (c *Client) QueryLegacy(ctx context.Context, id string, q querylegacy.Query, opts querylegacy.Options) (*querylegacy.Result, error) {
	return c.Datasets.QueryLegacy(ctx, id, q, opts)
}

func (c *Client) trace(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return c.tracer.Start(ctx, name, opts...)
}

func spanError(span trace.Span, err error) error {
	if err == nil || !span.IsRecording() {
		return err
	}

	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err)

	return err
}
