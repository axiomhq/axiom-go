package axiom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=Plan -linecomment -output=orgs_string.go

// Plan represents the plan of a deployment or organization.
type Plan uint8

// All available deployment or organization plans.
const (
	emptyPlan Plan = iota //

	Free       // free
	Pro        // pro
	Enterprise // enterprise
	Comped     // comped
)

func planFromString(s string) (plan Plan, err error) {
	switch s {
	case emptyPlan.String():
		plan = emptyPlan
	case Free.String():
		plan = Free
	case Pro.String():
		plan = Pro
	case Enterprise.String():
		plan = Enterprise
	case Comped.String():
		plan = Comped
	default:
		err = fmt.Errorf("unknown plan %q", s)
	}

	return plan, err
}

// MarshalJSON implements `json.Marshaler`. It is in place to marshal the Plan
// to its string representation because that's what the server expects.
func (p Plan) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// UnmarshalJSON implements `json.Unmarshaler`. It is in place to unmarshal the
// Plan from the string representation the server returns.
func (p *Plan) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}

	*p, err = planFromString(s)

	return err
}

// License of a deployment or organization.
type License struct {
	// ID of the license.
	ID string `json:"id"`
	// Issuer of the license.
	Issuer string `json:"issuer"`
	// IssuedTo describes who the license was issued to.
	IssuedTo string `json:"issuedTo"`
	// IssuedAt is the time the license was issued at.
	IssuedAt time.Time `json:"issuedAt"`
	// ValidFrom is the time the license is valid from.
	ValidFrom time.Time `json:"validFrom"`
	// ExpiresAt is the time the license expires.
	ExpiresAt time.Time `json:"expiresAt"`
	// Tier of the license.
	Tier Plan `json:"tier"`
	// DailyIngestGB is the daily amount of data in gigabytes that can be
	// ingested as part of the license.
	DailyIngestGB int `json:"dailyIngestGb"`
	// MaxUsers is the maximum amount of teams allowed.
	MaxUsers int `json:"maxUsers"`
	// MaxTeams is the maximum amount of user allowed.
	MaxTeams int `json:"maxTeams"`
	// MaxDatasets is the maximum amount of datasets allowed.
	MaxDatasets int `json:"maxDatasets"`
	// MaxQueryWindow is the maximum query window allowed.
	MaxQueryWindow time.Duration `json:"maxQueryWindowSeconds"`
	// MaxAuditWindow is the maximum audit window allowed.
	MaxAuditWindow time.Duration `json:"maxAuditWindowSeconds"`
	// WithRBAC specifies it RBAC is enabled.
	WithRBAC bool `json:"withRBAC"`
	// WithAuths specifies the supported authentication modes.
	WithAuths []string `json:"withAuths"`
	// Error is the last error (if any) on token refresh.
	Error string `json:"error"`
}

// MarshalJSON implements `json.Marshaler`. It is in place to marshal the
// MaxQueryWindow and MaxAuditWindow to seconds because that's what the server
// expects.
func (l License) MarshalJSON() ([]byte, error) {
	type localLicense License

	// Set to the value in seconds.
	l.MaxQueryWindow = time.Duration(l.MaxQueryWindow.Seconds())
	l.MaxAuditWindow = time.Duration(l.MaxAuditWindow.Seconds())

	return json.Marshal(localLicense(l))
}

// UnmarshalJSON implements `json.Unmarshaler`. It is in place to unmarshal the
// MaxQueryWindow and MaxAuditWindow into a proper time.Duration value because
// the server returns it in seconds.
func (l *License) UnmarshalJSON(b []byte) error {
	type localLicense *License

	if err := json.Unmarshal(b, localLicense(l)); err != nil {
		return err
	}

	// Set to a proper time.Duration value interpreting the server response
	// value in seconds.
	l.MaxQueryWindow = l.MaxQueryWindow * time.Second
	l.MaxAuditWindow = l.MaxAuditWindow * time.Second

	return nil
}

// Status is the status of the organization. It describes the usage of the plan
// an organization or licensee is billed for.
type Status struct {
	// DailyIngestUsedGB is the data volume in gigabytes that has been used
	// today.
	DailyIngestUsedGB float64 `json:"dailyIngestUsedGB"`
	// DailyIngestRemainingGB is the data volume in gigabytes that is remaining
	// today.
	DailyIngestRemainingGB float64 `json:"dailyIngestRemainingGB"`
	// DatasetsUsed is the amount of datasets used.
	DatasetsUsed int64 `json:"datasetsUsed"`
	// DatasetsUsed is the amount of datasets remaining.
	DatasetsRemaining int64 `json:"datasetsRemaining"`
	// UsersUsed is the amount of users used.
	UsersUsed int64 `json:"usersUsed"`
	// UsersRemaining is the amount of users remaining.
	UsersRemaining int64 `json:"usersRemaining"`
}

// SharedAccessKeys are the signing keys used to create shared access tokens
// that can be used by a third party to run queries on behalf of the
// organization. They can be rotated.
type SharedAccessKeys struct {
	// Primary signing key. Gets rotated to the secondary signing key after
	// rotation.
	Primary string `json:"primary"`
	// Secondary signing key. Gets rotated out.
	Secondary string `json:"secondary"`
}

// Organization represents an organization. For selfhost deployments, there is
// only one main organization, therefor it is referred to as deployment.
type Organization struct {
	// ID is the unique ID of the organization.
	ID string `json:"id"`
	// Name of the organization.
	Name string `json:"name"`
	// Slug of the organization.
	Slug string `json:"slug"`
	// Plan the deployment or organization is on.
	Plan Plan `json:"plan"`
	// PlanCreated is the time the plan was created.
	PlanCreated time.Time `json:"planCreated"`
	// PlanExpires is the time the plan will expire.
	PlanExpires time.Time `json:"planExpires"`
	// Trialed describes if the plan is trialed or not.
	Trialed bool `json:"trialed"`
	// PreviousPlan is the previous plan the deployment or organization was on.
	PreviousPlan Plan `json:"previousPlan"`
	// PreviousPlanCreated is the time the previous plan was created.
	PreviousPlanCreated time.Time `json:"previousPlanCreated"`
	// PreviousPlanExpired is the time the previous plan expired.
	PreviousPlanExpired time.Time `json:"previousPlanExpired"`
	// LastUsageSync is the last time the usage instance usage statistics were
	// synchronized.
	LastUsageSync time.Time `json:"lastUsageSync"`
	// Role the requesting user has on the deployment or the organization.
	Role UserRole `json:"role"`
	// PrimaryEmail of the user that issued the request.
	PrimaryEmail string `json:"primaryEmail"`
	// License of the deployment or organization.
	License License `json:"license"`
	// CreatedAt is the time the Organization was created.
	CreatedAt time.Time `json:"metaCreated"`
	// ModifiedAt is the time the Organization was last modified.
	ModifiedAt time.Time `json:"metaModified"`
	// Version of the organization.
	Version string `json:"metaVersion"`
}

// OrganizationCreateUpdateRequest is a request used to update an organization.
type OrganizationCreateUpdateRequest struct {
	// Name of the organization. Restricted to 30 characters.
	Name string `json:"name"`
}

// OrganizationsService handles communication with the organization related
// operations of the Axiom API. These methods can be used regardless of the
// use of Axiom Cloud or Axiom Selfhost.
//
// Axiom API Reference: /api/v1/orgs
type OrganizationsService service

// List all available organizations.
func (s *OrganizationsService) List(ctx context.Context) ([]*Organization, error) {
	var res []*Organization
	if err := s.client.call(ctx, http.MethodGet, s.basePath, nil, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// Get an organization by id.
func (s *OrganizationsService) Get(ctx context.Context, id string) (*Organization, error) {
	path := s.basePath + "/" + id

	var res Organization
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Update the organization identified by the given id with the given properties.
func (s *OrganizationsService) Update(ctx context.Context, id string, req OrganizationCreateUpdateRequest) (*Organization, error) {
	path := s.basePath + "/" + id

	var res Organization
	if err := s.client.call(ctx, http.MethodPut, path, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// License gets an organizations license.
func (s *OrganizationsService) License(ctx context.Context, id string) (*License, error) {
	path := s.basePath + "/" + id + "/license"

	var res License
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Status gets an organizations status.
func (s *OrganizationsService) Status(ctx context.Context, id string) (*Status, error) {
	path := s.basePath + "/" + id + "/status"

	var res Status
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// CloudOrganizationsService handles communication with the organization related
// operations of the Axiom API. Some of these methods are only available on
// Axiom Cloud. See OrganizationsService for methods, that exclusively work on
// Axiom Selfhost.
//
// Axiom API Reference: /api/v1/orgs
type CloudOrganizationsService struct {
	OrganizationsService
}

// ViewSharedAccessKeys rotates the shared access signing keys for the
// organization identified by the given id.
func (s *CloudOrganizationsService) ViewSharedAccessKeys(ctx context.Context, id string) (*SharedAccessKeys, error) {
	path := s.basePath + "/" + id + "/keys"

	var res SharedAccessKeys
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// RotateSharedAccessKeys rotates the shared access signing keys for the
// organization identified by the given id.
func (s *CloudOrganizationsService) RotateSharedAccessKeys(ctx context.Context, id string) (*SharedAccessKeys, error) {
	path := s.basePath + "/" + id + "/rotate-keys"

	var res SharedAccessKeys
	if err := s.client.call(ctx, http.MethodPut, path, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Create an organization with the given properties.
func (s *CloudOrganizationsService) Create(ctx context.Context, req OrganizationCreateUpdateRequest) (*Organization, error) {
	var res Organization
	if err := s.client.call(ctx, http.MethodPost, s.basePath, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Delete the organization identified by the given id.
func (s *CloudOrganizationsService) Delete(ctx context.Context, id string) error {
	return s.client.call(ctx, http.MethodDelete, s.basePath+"/"+id, nil, nil)
}
