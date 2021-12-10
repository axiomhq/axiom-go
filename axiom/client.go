package axiom

import (
	"bytes"
	"context"
	"encoding/json"
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

var validOnlyAPITokenPaths = regexp.MustCompile(`^/api/v1/datasets/([^/]+/(ingest|query)|_apl)(\?.+)?$`)

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

// Client provides the Axiom HTTP API operations.
type Client struct {
	baseURL     *url.URL
	accessToken string
	orgID       string

	httpClient     *http.Client
	userAgent      string
	strictDecoding bool
	noEnv          bool

	Dashboards    *DashboardsService
	Datasets      *DatasetsService
	Monitors      *MonitorsService
	Notifiers     *NotifiersService
	Organizations struct {
		Cloud    *CloudOrganizationsService
		Selfhost *OrganizationsService
	}
	StarredQueries *StarredQueriesService
	Teams          *TeamsService
	Tokens         struct {
		API      *APITokensService
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
	}

	client.Dashboards = &DashboardsService{client, "/api/v1/dashboards"}
	client.Datasets = &DatasetsService{client, "/api/v1/datasets"}
	client.Monitors = &MonitorsService{client, "/api/v1/monitors"}
	client.Notifiers = &NotifiersService{client, "/api/v1/notifiers"}
	client.Organizations.Cloud = &CloudOrganizationsService{OrganizationsService{client, "/api/v1/orgs"}}
	client.Organizations.Selfhost = &OrganizationsService{client, "/api/v1/orgs"}
	client.StarredQueries = &StarredQueriesService{client, "/api/v1/starred"}
	client.Teams = &TeamsService{client, "/api/v1/teams"}
	client.Tokens.API = &APITokensService{tokensService{client, "/api/v1/tokens/api"}}
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
	// is the Axiom Cloud url and the access token is not a personal token, the
	// organization ID is explicitly required and an error is returned, if it is
	// not set.
	cloudURLSetByOption := c.baseURL != nil && c.baseURL.String() == CloudURL
	cloudURLSetByEnvironment := deploymentURL == CloudURL
	cloudURLSet := cloudURLSetByOption || cloudURLSetByEnvironment
	isPersonalToken := IsPersonalToken(c.accessToken) || IsPersonalToken(accessToken)
	if c.orgID == "" {
		if (orgID == "" && cloudURLSet && isPersonalToken) || (c.noEnv && cloudURLSet) {
			return ErrMissingOrganizationID
		}
		addOption(SetOrgID(orgID))
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
		req.Header.Set("Content-Type", "application/json")
	} else if body != nil {
		req.Header.Set("Content-Type", "application/octet-stream")
	}

	// Set authorization header, if present.
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	// Set organization ID header when using a personal token.
	if IsPersonalToken(c.accessToken) && c.orgID != "" {
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
		// Handle a generic HTTP error if the response is not JSON formatted.
		if val := resp.Header.Get("Content-Type"); !strings.HasPrefix(val, "application/json") {
			return resp, Error{
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
		errResp := Error{Status: statusCode}
		if err = dec.Decode(&errResp); err != nil {
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
			return resp, fmt.Errorf("%v: %w", errResp, ErrRateLimitExceeded)
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
