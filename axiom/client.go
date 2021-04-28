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
	"strings"
	"time"
)

// CloudURL is the url of the cloud hosted version of Axiom.
const CloudURL = "https://cloud.axiom.co"

// ErrUnauthenticated is raised when the access token isn't valid.
var ErrUnauthenticated = errors.New("invalid authentication credentials")

// Error is the generic error response returned on non 2xx HTTP status codes.
// Either one of the two fields is populated. However, calling the Error()
// method is preferred.
type Error struct {
	ErrorMessage string `json:"error"`
	Message      string `json:"message"`

	statusCode int
}

// Error implements the error interface.
func (e Error) Error() string {
	if e.ErrorMessage != "" {
		return fmt.Sprintf("API error %d: %s", e.statusCode, e.ErrorMessage)
	}
	return fmt.Sprintf("API error %d: %s", e.statusCode, e.Message)
}

// service is the base service used by all Axiom API services.
//nolint:structcheck // https://github.com/golangci/golangci-lint/issues/1517
type service struct {
	client   *Client
	basePath string
}

// DefaultHTTPClient returns the default HTTP client used for making requests.
func DefaultHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 5 * time.Second,
			ForceAttemptHTTP2:   true,
		},
	}
}

// An Option modifies the behaviour of the API client. If not otherwise
// specified by a specific option, they are safe to use even after API methods
// have been called. However, they are not safe to use while the client is
// performing an operation.
type Option func(c *Client) error

// SetBaseURL sets the base URL used by the client. It overwrittes the one set
// by the call to NewClient() or NewCloudClient().
func SetBaseURL(baseURL string) Option {
	return func(c *Client) (err error) {
		c.baseURL, err = url.ParseRequestURI(baseURL)
		return err
	}
}

// SetClient specifies a custom http client that should be used to make
// requests.
func SetClient(client *http.Client) Option {
	return func(c *Client) error {
		if client == nil {
			return nil
		}
		c.httpClient = client
		return nil
	}
}

// SetOrgID specifies the organization ID to use. When a personal access token
// is used, this method can be used to switch between organizations without
// creating a new client instance.
func SetOrgID(orgID string) Option {
	return func(c *Client) error {
		c.orgID = orgID
		return nil
	}
}

// SetUserAgent sets the user agent used by the client.
func SetUserAgent(userAgent string) Option {
	return func(c *Client) error {
		c.userAgent = userAgent
		return nil
	}
}

// Client provides the Axiom HTTP API operations.
type Client struct {
	baseURL        *url.URL
	userAgent      string
	accessToken    string
	orgID          string
	strictDecoding bool

	httpClient *http.Client

	Dashboards     *DashboardsService
	Datasets       *DatasetsService
	Monitors       *MonitorsService
	Notifiers      *NotifiersService
	Organizations  *OrganizationsService
	StarredQueries *StarredQueriesService
	Teams          *TeamsService
	Tokens         struct {
		Ingest   *IngestTokensService
		Personal *PersonalTokensService
	}
	Users         *UsersService
	Version       *VersionService
	VirtualFields *VirtualFieldsService
}

// NewClient returns a new Axiom API client. The access token must be a personal
// or ingest token which can be created on the user profile or settings page of
// a deployment.
func NewClient(baseURL, accessToken string, options ...Option) (*Client, error) {
	u, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return nil, err
	}

	client := &Client{
		baseURL:     u,
		userAgent:   "axiom-go",
		accessToken: accessToken,

		httpClient: DefaultHTTPClient(),
	}

	client.Dashboards = &DashboardsService{client, "/api/v1/dashboards"}
	client.Datasets = &DatasetsService{client, "/api/v1/datasets"}
	client.Monitors = &MonitorsService{client, "/api/v1/monitors"}
	client.Notifiers = &NotifiersService{client, "/api/v1/notifiers"}
	client.Organizations = &OrganizationsService{client, "/api/v1/orgs"}
	client.StarredQueries = &StarredQueriesService{client, "/api/v1/starred"}
	client.Teams = &TeamsService{client, "/api/v1/teams"}
	client.Tokens.Ingest = &IngestTokensService{client, "/api/v1/tokens/ingest"}
	client.Tokens.Personal = &PersonalTokensService{client, "/api/v1/tokens/personal"}
	client.Users = &UsersService{client, "/api/v1/users"}
	client.Version = &VersionService{client, "/api/v1/version"}
	client.VirtualFields = &VirtualFieldsService{client, "/api/v1/vfields"}

	// Apply supplied options.
	if err := client.Options(options...); err != nil {
		return nil, err
	}

	return client, nil
}

// NewCloudClient is like NewClient but assumes the official Axiom Cloud URL as
// base URL and accepts an organization ID. When using an ingest token, the
// organization ID must match the organization the token was issued for.
func NewCloudClient(accessToken, orgID string, options ...Option) (*Client, error) {
	options = append(options, SetOrgID(orgID))
	return NewClient(CloudURL, accessToken, options...)
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

// call creates a new API request and executes it. The response body is JSON
// decoded or directly written to v, depending on v being an io.Writer or not.
func (c *Client) call(ctx context.Context, method, path string, body, v interface{}) error {
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return err
	}
	return c.do(req, v)
}

// newRequest creates an API request. If specified, the value pointed to by
// body will be included as the request body. If it isn't an io.Reader, it will
// be included as a JSON encoded request body.
func (c *Client) newRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Request, error) {
	rel, err := url.ParseRequestURI(endpoint)
	if err != nil {
		return nil, err
	}
	u := c.baseURL.ResolveReference(rel)

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
		req.Header.Set("content-type", "application/json")
	} else if body != nil {
		req.Header.Set("content-type", "application/octet-stream")
	}

	// Set authorization header, if present.
	if c.accessToken != "" {
		req.Header.Set("authorization", "Bearer "+c.accessToken)
	}

	// Set organization id header, if present.
	if c.orgID != "" {
		req.Header.Set("x-axiom-org-id", c.orgID)
	}

	// Set other headers.
	req.Header.Set("accept", "application/json")
	req.Header.Set("user-agent", c.userAgent)

	return req, nil
}

// do sends an API request and returns the API response. The response body is
// JSON decoded or directly written to v, depending on v being an io.Writer or
// not.
func (c *Client) do(req *http.Request, v interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if statusCode := resp.StatusCode; statusCode >= 400 {
		// Handle special errors.
		if statusCode == http.StatusForbidden {
			return ErrUnauthenticated
		}

		// Handle a generic HTTP error if the response is not JSON formatted.
		if val := resp.Header.Get("content-type"); !strings.HasPrefix(val, "application/json") {
			return Error{
				Message:    http.StatusText(statusCode),
				statusCode: statusCode,
			}
		}

		// For error handling, we want to have access to the raw request body
		// to inspect it further
		var (
			buf bytes.Buffer
			dec = json.NewDecoder(io.TeeReader(resp.Body, &buf))
		)

		// Handle a properly JSON formatted Axiom API error response.
		errResp := Error{statusCode: statusCode}
		if err = dec.Decode(&errResp); err != nil {
			return fmt.Errorf("error decoding %d error response: %w", statusCode, err)
		}

		// In case something went wrong, include the raw response and hope for
		// the best.
		if errResp.Message == "" && errResp.Error() == "" {
			s := strings.ReplaceAll(buf.String(), "\n", " ")
			errResp.Message = s
		}

		return errResp
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
			return err
		}

		dec := json.NewDecoder(resp.Body)
		if c.strictDecoding {
			dec.DisallowUnknownFields()
		}
		return dec.Decode(v)
	}

	return nil
}
