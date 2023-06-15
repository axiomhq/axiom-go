package axiom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=Plan,PaymentStatus -linecomment -output=orgs_string.go

// Plan represents the plan of an [Organization].
type Plan uint8

// All available [Organization] plans.
const (
	emptyPlan Plan = iota //

	Personal   // personal
	Basic      // basic
	Team       // teamMonthly
	Enterprise // enterprise
	Comped     // comped
)

func planFromString(s string) (plan Plan, err error) {
	switch s {
	case emptyPlan.String():
		plan = emptyPlan
	case Personal.String():
		plan = Personal
	case Basic.String():
		plan = Basic
	case Team.String():
		plan = Team
	case Enterprise.String():
		plan = Enterprise
	case Comped.String():
		plan = Comped
	default:
		err = fmt.Errorf("unknown plan %q", s)
	}

	return plan, err
}

// MarshalJSON implements [json.Marshaler]. It is in place to marshal the plan
// to its string representation because that's what the server expects.
func (p Plan) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// UnmarshalJSON implements [json.Unmarshaler]. It is in place to unmarshal the
// plan from the string representation the server returns.
func (p *Plan) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}

	*p, err = planFromString(s)

	return err
}

// PaymentStatus represents the payment status of an [Organization].
type PaymentStatus uint8

// All available [Organization] payment statuses.
const (
	emptyPaymentStatus PaymentStatus = iota //

	Success      // success
	NotAvailable // na
	Failed       // failed
	Blocked      // blocked
)

func paymentStatusFromString(s string) (paymentStatus PaymentStatus, err error) {
	switch s {
	case emptyPaymentStatus.String():
		paymentStatus = emptyPaymentStatus
	case Success.String():
		paymentStatus = Success
	case NotAvailable.String():
		paymentStatus = NotAvailable
	case Failed.String():
		paymentStatus = Failed
	case Blocked.String():
		paymentStatus = Blocked
	default:
		err = fmt.Errorf("unknown payment status %q", s)
	}

	return paymentStatus, err
}

// MarshalJSON implements [json.Marshaler]. It is in place to marshal the
// payment status to its string representation because that's what the server
// expects.
func (ps PaymentStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(ps.String())
}

// UnmarshalJSON implements [json.Unmarshaler]. It is in place to unmarshal the
// payment status from the string representation the server returns.
func (ps *PaymentStatus) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}

	*ps, err = paymentStatusFromString(s)

	return err
}

// License of an [Organization].
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
	// Plan associated with the license.
	Plan Plan `json:"tier"`
	// MonthlyIngestGB is the monthly amount of data in gigabytes that can be
	// ingested as part of the license.
	MonthlyIngestGB uint64 `json:"monthlyIngestGb"`
	// MaxUsers is the maximum amount of teams allowed.
	MaxUsers uint64 `json:"maxUsers"`
	// MaxTeams is the maximum amount of user allowed.
	MaxTeams uint64 `json:"maxTeams"`
	// MaxDatasets is the maximum amount of datasets allowed.
	MaxDatasets uint64 `json:"maxDatasets"`
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

// MarshalJSON implements [json.Marshaler]. It is in place to marshal the
// MaxQueryWindow and MaxAuditWindow to seconds because that's what the server
// expects.
func (l License) MarshalJSON() ([]byte, error) {
	type localLicense License

	// Set to the value in seconds.
	l.MaxQueryWindow = time.Duration(l.MaxQueryWindow.Seconds())
	l.MaxAuditWindow = time.Duration(l.MaxAuditWindow.Seconds())

	return json.Marshal(localLicense(l))
}

// UnmarshalJSON implements [json.Unmarshaler]. It is in place to unmarshal the
// MaxQueryWindow and MaxAuditWindow into a proper [time.Duration] value because
// the server returns it in seconds.
func (l *License) UnmarshalJSON(b []byte) error {
	type localLicense *License

	if err := json.Unmarshal(b, localLicense(l)); err != nil {
		return err
	}

	// Set to a proper [time.Duration] value by interpreting the server response
	// value in seconds.
	l.MaxQueryWindow = l.MaxQueryWindow * time.Second
	l.MaxAuditWindow = l.MaxAuditWindow * time.Second

	return nil
}

// Organization represents an organization.
type Organization struct {
	// ID is the unique ID of the organization.
	ID string `json:"id"`
	// Name of the organization.
	Name string `json:"name"`
	// Slug of the organization.
	Slug string `json:"slug"`
	// Trial describes if the plan is trialed or not.
	Trial bool `json:"inTrial"`
	// Plan the organization is on.
	Plan Plan `json:"plan"`
	// PlanCreated is the time the plan was created.
	PlanCreated time.Time `json:"planCreated"`
	// LastUsageSync is the last time the usage instance usage statistics were
	// synchronized.
	LastUsageSync time.Time `json:"lastUsageSync"`
	// Role the requesting user has on the deployment or the organization.
	Role UserRole `json:"role"`
	// PrimaryEmail of the user that issued the request.
	PrimaryEmail string `json:"primaryEmail"`
	// License of the organization.
	License License `json:"license"`
	// PaymentStatus is the status of the current payment for the organization.
	PaymentStatus PaymentStatus `json:"paymentStatus"`
	// CreatedAt is the time the organization was created.
	CreatedAt time.Time `json:"metaCreated"`
	// ModifiedAt is the time the organization was last modified.
	ModifiedAt time.Time `json:"metaModified"`
}

type wrappedOrganization struct {
	Organization

	// HINT(lukasmalkmus): Ignore these fields because they do not provide any
	// value to the user.
	FirstFailedPayment any `json:"firstFailedPayment"`
	ExternalPlan       any `json:"externalPlan"`
	Version            any `json:"metaVersion"`
}

// OrganizationsService handles communication with the organization related
// operations of the Axiom API.
//
// Axiom API Reference: /v1/orgs
type OrganizationsService service

// List all available organizations.
func (s *OrganizationsService) List(ctx context.Context) ([]*Organization, error) {
	ctx, span := s.client.trace(ctx, "Organizations.List")
	defer span.End()

	var res []*wrappedOrganization
	if err := s.client.Call(ctx, http.MethodGet, s.basePath, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	organizations := make([]*Organization, len(res))
	for i, r := range res {
		organizations[i] = &r.Organization
	}

	return organizations, nil
}

// Get an organization by id.
func (s *OrganizationsService) Get(ctx context.Context, id string) (*Organization, error) {
	ctx, span := s.client.trace(ctx, "Organizations.Get", trace.WithAttributes(
		attribute.String("axiom.dataset_id", id),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, id)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res wrappedOrganization
	if err := s.client.Call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res.Organization, nil
}
