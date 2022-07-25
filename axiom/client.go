package axiom

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/klauspost/compress/gzhttp"
)

// CloudURL is the url of the cloud hosted version of Axiom.
const CloudURL = "https://cloud.axiom.co"

const (
	headerAuthorization  = "Authorization"
	headerOrganizationID = "X-Axiom-Org-Id"

	headerAccept      = "Accept"
	headerContentType = "Content-Type"
	headerUserAgent   = "User-Agent"

	defaultMediaType = "application/octet-stream"
	mediaTypeJSON    = "application/json"
	mediaTypeNDJSON  = "application/x-ndjson"
)

var validOnlyAPITokenPaths = regexp.MustCompile(`^/api/v1/datasets/([^/]+/(ingest|query)|_apl)(\?.+)?$`)

// service is the base service used by all Axiom API services.
type service struct {
	client   *Client
	basePath string
}

// DefaultHTTPClient returns the default HTTP client used for making requests.
func DefaultHTTPClient() *http.Client {
	return &http.Client{
		Transport: gzhttp.Transport(&http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 5 * time.Second,
			ForceAttemptHTTP2:   true,
		}),
	}
}

// Client provides the Axiom HTTP API operations.
type Client struct {
	baseURL     *url.URL
	accessToken string
	orgID       string

	httpClient     *http.Client
	userAgent      string
	strictDecoding bool
	noEnv          bool
	noLimiting     bool

	// Rate limit for the client as determined by the most recent API call.
	limits   map[string]Limit
	limitsMu sync.Mutex

	// Services for communicating with different parts of the GitHub API.
	Datasets      *DatasetsService
	Organizations *OrganizationsService
	Users         *UsersService
}

// NewClient returns a new Axiom API client. It automatically takes its
// configuration from the environment.
//
// To connect to Axiom Cloud:
//
//   - AXIOM_TOKEN
//   - AXIOM_ORG_ID (only when using a personal token)
//
// To connect to an Axiom Selfhost:
//
//   - AXIOM_URL
//   - AXIOM_TOKEN
//
// The configuration can be set manually using `Option` functions prefixed with
// `Set`. Refer to `SetCloudConfig()` and `SetSelfhostConfig()`. Individual
// properties can be overwritten as well.
//
// The access token must be an api or personal token which can be created on
// the settings or user profile page on Axiom.
func NewClient(options ...Option) (*Client, error) {
	client := &Client{
		userAgent: "axiom-go",

		httpClient: DefaultHTTPClient(),

		limits: make(map[string]Limit),
	}

	// Try to get module version to include in the user agent.
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			if dep.Path == "github.com/axiomhq/axiom-go" {
				client.userAgent += fmt.Sprintf("/%s", dep.Version)
				break
			}
		}
	}

	client.Datasets = &DatasetsService{client, "/api/v1/datasets"}
	client.Organizations = &OrganizationsService{client, "/api/v1/orgs"}
	client.Users = &UsersService{client, "/api/v1/users"}

	// Apply supplied options.
	if err := client.Options(options...); err != nil {
		return nil, err
	}

	// Make sure to populate remaining fields from the environment or fail.
	return client, client.populateClientFromEnvironment()
}

// Options applies Options to the Client.
func (c *Client) Options(options ...Option) error {
	for _, option := range options {
		if err := option(c); err != nil {
			return err
		}
	}
	return nil
}

// ValidateCredentials makes sure the client can properly authenticate against
// the configured Axiom deployment.
func (c *Client) ValidateCredentials(ctx context.Context) error {
	if IsPersonalToken(c.accessToken) {
		_, err := c.Users.Current(ctx)
		return err
	}

	// FIXME(lukasmalkmus): Well, with the current API, we need to assume the
	// token is valid.
	// return ErrInvalidToken
	return nil
}

// populateClientFromEnvironment populates the client with values from the
// environment. It omits properties that have already been set by user options.
func (c *Client) populateClientFromEnvironment() (err error) {
	var (
		deploymentURL = os.Getenv("AXIOM_URL")
		accessToken   = os.Getenv("AXIOM_TOKEN")
		orgID         = os.Getenv("AXIOM_ORG_ID")

		options   = make([]Option, 0)
		addOption = func(option Option) {
			options = append(options, option)
		}
	)

	// When the base url is not set, use `AXIOM_URL` or default to the Axiom
	// Cloud url.
	if c.baseURL == nil {
		if deploymentURL == "" || c.noEnv {
			deploymentURL = CloudURL
		}
		addOption(SetURL(deploymentURL))
	}

	// When the access token is not set, use `AXIOM_TOKEN`.
	if c.accessToken == "" {
		if accessToken == "" || c.noEnv {
			return ErrMissingAccessToken
		}
		addOption(SetAccessToken(accessToken))
	}

	// When the organization ID is not set, use `AXIOM_ORG_ID`. In case the url
	// is the Axiom Cloud url and the access token is a personal token, the
	// organization ID is explicitly required and an error is returned, if it is
	// not set.
	cloudURLSetByOption := c.baseURL != nil && c.baseURL.String() == CloudURL
	cloudURLSetByEnvironment := deploymentURL == CloudURL
	cloudURLSet := cloudURLSetByOption || cloudURLSetByEnvironment
	isPersonalToken := IsPersonalToken(c.accessToken) || IsPersonalToken(accessToken)
	if c.orgID == "" && isPersonalToken {
		if (orgID == "" && cloudURLSet) || (c.noEnv && cloudURLSet) {
			return ErrMissingOrganizationID
		} else if !c.noEnv {
			addOption(SetOrgID(orgID))
		}
	}
	return c.Options(options...)
}

// call creates a new API request and executes it. The response body is JSON
// decoded or directly written to v, depending on v being an io.Writer or not.
func (c *Client) call(ctx context.Context, method, path string, body, v any) error {
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return err
	} else if _, err = c.do(req, v); err != nil {
		return err
	}
	return nil
}

// newRequest creates an API request. If specified, the value pointed to by body
// will be included as the request body. If it is not an io.Reader, it will be
// included as a JSON encoded request body.
func (c *Client) newRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	rel, err := url.ParseRequestURI(path)
	if err != nil {
		return nil, err
	}
	u := c.baseURL.ResolveReference(rel)

	if IsAPIToken(c.accessToken) && !validOnlyAPITokenPaths.MatchString(u.Path) {
		return nil, ErrUnprivilegedToken
	}

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

	req, err := http.NewRequestWithContext(ctx, method, u.String(), r)
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
	if c.accessToken != "" {
		req.Header.Set(headerAuthorization, "Bearer "+c.accessToken)
	}

	// Set organization ID header when using a personal token.
	if IsPersonalToken(c.accessToken) && c.orgID != "" {
		req.Header.Set(headerOrganizationID, c.orgID)
	}

	// Set other headers.
	req.Header.Set(headerAccept, mediaTypeJSON)
	req.Header.Set(headerUserAgent, c.userAgent)

	return req, nil
}

// do sends an API request and returns the API response. The response body is
// JSON decoded or directly written to v, depending on v being an io.Writer or
// not.
// If the rate limit is exceeded and reset time is in the future, it returns
// `*LimitError` immediately without making an API call.
func (c *Client) do(req *http.Request, v any) (*response, error) {
	// If we've hit the rate limit, don't make further requests before
	// the reset time.
	if err := c.checkLimit(req); err != nil {
		return &response{
			Response: err.response,

			Limit: err.Limit,
		}, err
	}

	bck := backoff.NewExponentialBackOff()
	bck.InitialInterval = 200 * time.Millisecond
	bck.Multiplier = 2.0
	bck.MaxElapsedTime = 10 * time.Second

	var resp *response
	err := backoff.Retry(func() error {
		httpResp, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}

		resp = newResponse(httpResp)

		// We should only retry in the case the status code is >= 500, anything below isn't worth retrying.
		if code := resp.StatusCode; code >= 500 {
			return fmt.Errorf("got status code %d", code)
		}

		return nil
	}, bck)

	defer func() {
		if resp != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}
	}()

	if err != nil {
		return resp, err
	}

	c.limitsMu.Lock()
	c.limits[resp.Limit.limitType.String()] = resp.Limit
	c.limitsMu.Unlock()

	if statusCode := resp.StatusCode; statusCode >= 400 {
		// Handle a generic HTTP error if the response is not JSON formatted.
		if val := resp.Header.Get(headerContentType); !strings.HasPrefix(val, mediaTypeJSON) {
			return resp, &Error{
				Status:  statusCode,
				Message: http.StatusText(statusCode),
			}
		}

		// For error handling, we want to have access to the raw request body
		// to inspect it further
		var (
			buf bytes.Buffer
			dec = json.NewDecoder(io.TeeReader(resp.Body, &buf))
		)

		// Handle a properly JSON formatted Axiom API error response.
		errResp := &Error{Status: statusCode}
		if err := dec.Decode(&errResp); err != nil {
			return resp, fmt.Errorf("error decoding %d error response: %w", statusCode, err)
		}

		// In case something went wrong, include the raw response and hope for
		// the best.
		if errResp.Message == "" {
			s := strings.ReplaceAll(buf.String(), "\n", " ")
			errResp.Message = s
			return resp, errResp
		}

		// In case everything went fine till this point, handle special errors
		// and wrap them with our errors so user can check for them using
		// `errors.Is()`.
		switch statusCode {
		case http.StatusUnauthorized:
			return resp, fmt.Errorf("%v: %w", errResp, ErrUnauthenticated)
		case http.StatusForbidden:
			return resp, fmt.Errorf("%v: %w", errResp, ErrUnauthorized)
		case http.StatusNotFound:
			return resp, fmt.Errorf("%v: %w", errResp, ErrNotFound)
		case http.StatusConflict:
			return resp, fmt.Errorf("%v: %w", errResp, ErrExists)
		case http.StatusTooManyRequests:
			return resp, &LimitError{
				Limit:   resp.Limit,
				Message: errResp.Message,

				response: resp.Response,
			}
		}

		return resp, errResp
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err := io.Copy(w, resp.Body)
			return resp, err
		}

		if val := resp.Header.Get(headerContentType); strings.HasPrefix(val, mediaTypeJSON) {
			dec := json.NewDecoder(resp.Body)
			if c.strictDecoding {
				dec.DisallowUnknownFields()
			}
			return resp, dec.Decode(v)
		}

		return resp, errors.New("cannot decode response with unknown content type")
	}

	return resp, nil
}

// checkLimit checks if *LimitError can be immediately returned from
// `Client.do`, and if so, returns it so that `Client.do` can skip making an API
// call unnecessarily.
func (c *Client) checkLimit(req *http.Request) *LimitError {
	if !c.noLimiting {
		return nil
	}

	var lt limitType
	if strings.HasSuffix(req.URL.Path, "/ingest") {
		lt = limitIngest
	} else if strings.HasSuffix(req.URL.Path, "/query") || strings.HasSuffix(req.URL.Path, "/_apl") {
		lt = limitQuery
	} else {
		return nil
	}

	c.limitsMu.Lock()
	defer c.limitsMu.Unlock()

	limit, ok := c.limits[lt.String()]
	if !ok {
		return nil
	}

	if !limit.Reset.IsZero() && limit.Remaining == 0 && time.Now().Before(limit.Reset) {
		// Create a fake response.
		resp := &http.Response{
			Status:     http.StatusText(http.StatusTooManyRequests),
			StatusCode: http.StatusTooManyRequests,
			Request:    req,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("")),
		}
		return &LimitError{
			Limit:   limit,
			Message: fmt.Sprintf("%s limit exceeded, not making remote request", lt.String()),

			response: resp,
		}
	}

	return nil
}
