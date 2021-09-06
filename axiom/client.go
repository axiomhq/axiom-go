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
	"strings"
	"time"
)

// CloudURL is the url of the cloud hosted version of Axiom.
const CloudURL = "https://cloud.axiom.co"

var (
	// ErrInvalidToken is returned when the access token is invalid.
	ErrInvalidToken = errors.New("invalid access token")

	// ErrMissingAccessToken is raised when an access token is not provided. Set
	// it manually using the SetAccessToken option or export `AXIOM_TOKEN`.
	ErrMissingAccessToken = errors.New("missing access token")

	// ErrMissingOrganizationID is raised when an organization ID is not
	// provided. Set it manually using the SetOrgID option or export
	// `AXIOM_ORG_ID`.
	ErrMissingOrganizationID = errors.New("missing organization id")

	// ErrUnauthenticated is raised when the access token is not valid.
	ErrUnauthenticated = errors.New("invalid authentication credentials")

	// ErrUnprivilegedToken is raised when a client tries to call a non-ingest
	// endpoint with an ingest-only token configured.
	ErrUnprivilegedToken = errors.New("using ingest token for non-ingest operation")
)

var validIngestTokenPathRe = regexp.MustCompile("^/api/v1/(datasets/.+/ingest|tokens/ingest/validate)$")

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

// response wraps the default http.Response type. It never has an open body.
type response struct {
	*http.Response
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

// SetAccessToken specifies the access token to use. Can also be specified using
// the `AXIOM_TOKEN` environment variable.
func SetAccessToken(accessToken string) Option {
	return func(c *Client) error {
		if !IsValidToken(accessToken) {
			return ErrInvalidToken
		}
		c.accessToken = accessToken
		return nil
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

// SetCloudConfig specifies all properties needed in order to successfully
// connect to Axiom Cloud.
func SetCloudConfig(accessToken, orgID string) Option {
	return func(c *Client) error {
		return c.Options(
			SetAccessToken(accessToken),
			SetOrgID(orgID),
		)
	}
}

// SetOrgID specifies the organization ID to use when connecting to Axiom Cloud.
// When a personal access token is used, this method can be used to switch
// between organizations without creating a new client instance. Can also be
// specified using the `AXIOM_ORG_ID` environment variable.
func SetOrgID(orgID string) Option {
	return func(c *Client) error {
		c.orgID = orgID
		return nil
	}
}

// SetSelfhostConfig specifies all properties needed in order to successfully
// connect to an Axiom Selfhost deployment.
func SetSelfhostConfig(deploymentURL, accessToken string) Option {
	return func(c *Client) error {
		return c.Options(
			SetURL(deploymentURL),
			SetAccessToken(accessToken),
		)
	}
}

// SetURL sets the base URL used by the client. Can also be specified using the
// `AXIOM_URL` environment variable.
func SetURL(baseURL string) Option {
	return func(c *Client) (err error) {
		c.baseURL, err = url.ParseRequestURI(baseURL)
		return err
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
	baseURL     *url.URL
	accessToken string
	orgID       string

	httpClient     *http.Client
	userAgent      string
	strictDecoding bool

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

// NewClient returns a new Axiom API client. It automatically takes its
// configuration from the environment.
//
// To connect to Axiom Cloud:
//
//   - AXIOM_TOKEN
//   - AXIOM_ORG_ID
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
// The access token must be a personal or ingest token which can be created on
// the user profile or settings page on Axiom.
func NewClient(options ...Option) (*Client, error) {
	client := &Client{
		userAgent: "axiom-go",

		httpClient: DefaultHTTPClient(),
	}

	client.Dashboards = &DashboardsService{client, "/api/v1/dashboards"}
	client.Datasets = &DatasetsService{client, "/api/v1/datasets"}
	client.Monitors = &MonitorsService{client, "/api/v1/monitors"}
	client.Notifiers = &NotifiersService{client, "/api/v1/notifiers"}
	client.Organizations = &OrganizationsService{client, "/api/v1/orgs"}
	client.StarredQueries = &StarredQueriesService{client, "/api/v1/starred"}
	client.Teams = &TeamsService{client, "/api/v1/teams"}
	client.Tokens.Ingest = &IngestTokensService{tokensService{client, "/api/v1/tokens/ingest"}}
	client.Tokens.Personal = &PersonalTokensService{tokensService{client, "/api/v1/tokens/personal"}}
	client.Users = &UsersService{client, "/api/v1/users"}
	client.Version = &VersionService{client, "/api/v1/version"}
	client.VirtualFields = &VirtualFieldsService{client, "/api/v1/vfields"}

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
	if IsIngestToken(c.accessToken) {
		return c.Tokens.Ingest.Validate(ctx)
	} else if IsPersonalToken(c.accessToken) {
		_, err := c.Users.Current(ctx)
		return err
	}
	return ErrInvalidToken
}

// populateClientFromEnvironment populates the client with values from the
// environment. It omits properties that have already been set by user options.
func (c *Client) populateClientFromEnvironment() (err error) {
	var (
		deploymentURL  = os.Getenv("AXIOM_URL")
		organizationID = os.Getenv("AXIOM_ORG_ID")
		accessToken    = os.Getenv("AXIOM_TOKEN")

		options   = make([]Option, 0)
		addOption = func(option Option) {
			options = append(options, option)
		}
	)

	// When the base url is not set, use `AXIOM_URL` or default to the Axiom
	// Cloud url.
	if c.baseURL == nil {
		if deploymentURL == "" {
			deploymentURL = CloudURL
		}
		addOption(SetURL(deploymentURL))
	}

	// When the base url is set to the Axiom Cloud url but no organization ID is
	// set, use `AXIOM_ORG_ID`.
	cloudURLSetByOption := c.baseURL != nil && c.baseURL.String() == CloudURL
	cloudURLSetByEnvironment := deploymentURL == CloudURL
	if (cloudURLSetByOption || cloudURLSetByEnvironment) && c.orgID == "" {
		if organizationID == "" {
			return ErrMissingOrganizationID
		}
		addOption(SetOrgID(organizationID))
	}

	// When the access token is not set, use `AXIOM_TOKEN`.
	if c.accessToken == "" {
		if accessToken == "" {
			return ErrMissingAccessToken
		}
		addOption(SetAccessToken(accessToken))
	}

	return c.Options(options...)
}

// call creates a new API request and executes it. The response body is JSON
// decoded or directly written to v, depending on v being an io.Writer or not.
func (c *Client) call(ctx context.Context, method, path string, body, v interface{}) error {
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
func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	rel, err := url.ParseRequestURI(path)
	if err != nil {
		return nil, err
	}
	u := c.baseURL.ResolveReference(rel)

	if IsIngestToken(c.accessToken) && !validIngestTokenPathRe.MatchString(u.Path) {
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
		req.Header.Set("Content-Type", "application/json")
	} else if body != nil {
		req.Header.Set("Content-Type", "application/octet-stream")
	}

	// Set authorization header, if present.
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	// Set organization ID header, if present.
	if c.orgID != "" {
		req.Header.Set("X-Axiom-Org-Id", c.orgID)
	}

	// Set other headers.
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	return req, nil
}

// do sends an API request and returns the API response. The response body is
// JSON decoded or directly written to v, depending on v being an io.Writer or
// not.
func (c *Client) do(req *http.Request, v interface{}) (*response, error) {
	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	resp := &response{httpResp}

	if statusCode := resp.StatusCode; statusCode >= 400 {
		// Handle special errors.
		if statusCode == http.StatusForbidden {
			return resp, ErrUnauthenticated
		}

		// Handle a generic HTTP error if the response is not JSON formatted.
		if val := resp.Header.Get("Content-Type"); !strings.HasPrefix(val, "application/json") {
			return resp, Error{
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
			return resp, fmt.Errorf("error decoding %d error response: %w", statusCode, err)
		}

		// In case something went wrong, include the raw response and hope for
		// the best.
		if errResp.Message == "" && errResp.Error() == "" {
			s := strings.ReplaceAll(buf.String(), "\n", " ")
			errResp.Message = s
		}

		return resp, errResp
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
			return resp, err
		}

		dec := json.NewDecoder(resp.Body)
		if c.strictDecoding {
			dec.DisallowUnknownFields()
		}
		return resp, dec.Decode(v)
	}

	return resp, nil
}
