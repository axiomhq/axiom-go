package axiom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

//go:generate ../bin/stringer -type=Plan -linecomment -output=organizations_string.go

// Plan represents the plan of a deployment or organization.
type Plan uint8

// All available deployment or organization plans.
const (
	Free       Plan = iota + 1 // free
	Trial                      // trial
	Pro                        // pro
	Enterprise                 // enterprise
)

// MarshalJSON implements json.Marshaler. It is in place to marshal the Plan to
// its string representation because that's what the server expects.
func (plan Plan) MarshalJSON() ([]byte, error) {
	return json.Marshal(plan.String())
}

// UnmarshalJSON implements json.Unmarshaler. It is in place to unmarshal the
// Plan from the string representation the server returns.
func (plan *Plan) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch s {
	case Free.String():
		*plan = Free
	case Trial.String():
		*plan = Trial
	case Pro.String():
		*plan = Pro
	case Enterprise.String():
		*plan = Enterprise
	default:
		return fmt.Errorf("unknown plan %q", s)
	}

	return nil
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
	// MaxQueriesPerSecond is the maximum amount of queries allowed per second.
	MaxQueriesPerSecond int `json:"maxQueriesPerSecond"`
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

// MarshalJSON implements json.Marshaler. It is in place to marshal the
// MaxQueryWindow and MaxAuditWindow to seconds because that's what the server
// expects.
func (l License) MarshalJSON() ([]byte, error) {
	type localLicense License

	// Set to the value in seconds.
	l.MaxQueryWindow = time.Duration(l.MaxQueryWindow.Seconds())
	l.MaxAuditWindow = time.Duration(l.MaxAuditWindow.Seconds())

	return json.Marshal(localLicense(l))
}

// UnmarshalJSON implements json.Unmarshaler. It is in place to unmarshal the
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

// Organization represents an organization. For selfhost deployments, there is
// only one main organization, therefor it is referred to as deployment.
type Organization struct {
	// ID is the unique id of the organization.
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
	// Role the requesting user has on the deployment or the organization.
	Role UserRole `json:"role"`
	// PrimaryEmail of the user that issued the request.
	PrimaryEmail string `json:"primaryEmail"`
	// License of the deployment or organization.
	License License `json:"license"`
	// Created is the time the Organization was created.
	Created time.Time `json:"metaCreated"`
	// Modified is the time the Organization was modified.
	Modified time.Time `json:"metaModified"`
	// Version of the organization.
	Version int64 `json:"metaVersion"`
}

// OrganizationUpdateRequest is a request used to update an organization.
type OrganizationUpdateRequest struct {
	// Name of the user.
	Name string `json:"name"`
}

// OrganizationsService handles communication with the organization related
// operations of the Axiom API.
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

// License gets an organizations license.
func (s *OrganizationsService) License(ctx context.Context, organizationID string) (*License, error) {
	path := s.basePath + "/" + organizationID + "/license"

	var res License
	if err := s.client.call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Update the organization identified by the given id with the given properties.
func (s *OrganizationsService) Update(ctx context.Context, id string, req OrganizationUpdateRequest) (*Organization, error) {
	path := s.basePath + "/" + id

	var res Organization
	if err := s.client.call(ctx, http.MethodPut, path, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}