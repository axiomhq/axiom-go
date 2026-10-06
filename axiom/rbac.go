package axiom

import (
	"context"
	"net/http"
	"net/url"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Role represents a custom role that grants capabilities to its members.
type Role struct {
	// ID is the unique ID of the role.
	ID string `json:"id"`
	// Name of the role. Must be unique within the organization.
	Name string `json:"name"`
	// Description of the role.
	Description string `json:"description"`
	// Members are the IDs of the users whose base role is this role. Members
	// are read-only: a user's base role is changed with
	// [UsersService.UpdateUsersRole]. Users who receive the role through a
	// [Group] are not included.
	Members []string `json:"members"`
	// DatasetCapabilities are the capabilities the role grants on individual
	// datasets, keyed by dataset name.
	DatasetCapabilities map[string]RoleDatasetCapabilities `json:"datasetCapabilities"`
	// ViewCapabilities are the capabilities the role grants on individual
	// views, keyed by view name.
	ViewCapabilities map[string]RoleViewCapabilities `json:"viewCapabilities"`
	// OrgCapabilities are the organization-wide capabilities the role grants.
	OrgCapabilities RoleOrgCapabilities `json:"orgCapabilities"`
}

// RoleRequest represents a request to create or update a [Role].
type RoleRequest struct {
	// Name of the role. Must be unique within the organization and must not
	// start with "axiom-".
	Name string `json:"name"`
	// Description of the role.
	Description string `json:"description"`
	// DatasetCapabilities are the capabilities the role grants on individual
	// datasets, keyed by dataset name.
	DatasetCapabilities map[string]RoleDatasetCapabilities `json:"datasetCapabilities,omitempty"`
	// ViewCapabilities are the capabilities the role grants on individual
	// views, keyed by view name. The wildcard view "*" is not allowed.
	ViewCapabilities map[string]RoleViewCapabilities `json:"viewCapabilities,omitempty"`
	// OrgCapabilities are the organization-wide capabilities the role grants.
	OrgCapabilities RoleOrgCapabilities `json:"orgCapabilities"`
}

// RoleDatasetCapabilities are the capabilities a [Role] grants on a dataset.
type RoleDatasetCapabilities struct {
	// Ingest is the ingest capability. Allowed: [ActionCreate].
	Ingest []Action `json:"ingest,omitempty"`
	// Query is the query capability. Allowed: [ActionRead].
	Query []Action `json:"query,omitempty"`
	// StarredQueries is the starred queries capability.
	StarredQueries []Action `json:"starredQueries,omitempty"`
	// VirtualFields is the virtual fields capability.
	VirtualFields []Action `json:"virtualFields,omitempty"`
	// Trim is the trim capability. Allowed: [ActionUpdate].
	Trim []Action `json:"trim,omitempty"`
	// Vacuum is the vacuum capability. Allowed: [ActionUpdate].
	Vacuum []Action `json:"vacuum,omitempty"`
	// Share is the share capability. Allowed: [ActionCreate], [ActionRead],
	// [ActionDelete].
	Share []Action `json:"share,omitempty"`
}

// RoleViewCapabilities are the capabilities a [Role] grants on a view.
type RoleViewCapabilities struct {
	// Query is the query capability. Allowed: [ActionRead].
	Query []Action `json:"query,omitempty"`
	// Share is the share capability. Allowed: [ActionCreate], [ActionRead],
	// [ActionDelete].
	Share []Action `json:"share,omitempty"`
}

// RoleOrgCapabilities are the organization-wide capabilities a [Role] grants.
type RoleOrgCapabilities struct {
	// Annotations is the annotations capability.
	Annotations []Action `json:"annotations,omitempty"`
	// APITokens is the API tokens capability.
	APITokens []Action `json:"apiTokens,omitempty"`
	// AuditLog is the audit log capability. Allowed: [ActionRead].
	AuditLog []Action `json:"auditLog,omitempty"`
	// Billing is the billing capability. Allowed: [ActionRead],
	// [ActionUpdate].
	Billing []Action `json:"billing,omitempty"`
	// Dashboards is the dashboards capability.
	Dashboards []Action `json:"dashboards,omitempty"`
	// Datasets is the datasets capability.
	Datasets []Action `json:"datasets,omitempty"`
	// Endpoints is the endpoints capability.
	Endpoints []Action `json:"endpoints,omitempty"`
	// Integrations is the integrations capability.
	Integrations []Action `json:"integrations,omitempty"`
	// Labels is the labels capability.
	Labels []Action `json:"labels,omitempty"`
	// Monitors is the monitors capability.
	Monitors []Action `json:"monitors,omitempty"`
	// Notifiers is the notifiers capability.
	Notifiers []Action `json:"notifiers,omitempty"`
	// RBAC is the roles and groups capability.
	RBAC []Action `json:"rbac,omitempty"`
	// SharedAccessKeys is the shared access keys capability. Allowed:
	// [ActionRead], [ActionUpdate].
	SharedAccessKeys []Action `json:"sharedAccessKeys,omitempty"`
	// Users is the users capability.
	Users []Action `json:"users,omitempty"`
	// Views is the views capability.
	Views []Action `json:"views,omitempty"`
}

// Group represents a group of users that receive the roles assigned to it.
type Group struct {
	// ID is the unique ID of the group.
	ID string `json:"id"`
	// Name of the group. Must be unique within the organization.
	Name string `json:"name"`
	// Description of the group.
	Description string `json:"description"`
	// Roles are the IDs of the roles assigned to the group.
	Roles []string `json:"roles"`
	// Members are the IDs of the users in the group.
	Members []string `json:"members"`
	// IsManaged is true if the group is synced from an identity provider.
	// The members of a managed group can't be changed through the API.
	IsManaged bool `json:"isManaged"`
}

// GroupRequest represents a request to create or update a [Group].
type GroupRequest struct {
	// Name of the group. Must be unique within the organization.
	Name string `json:"name"`
	// Description of the group.
	Description string `json:"description"`
	// Roles are the IDs of the roles assigned to the group.
	Roles []string `json:"roles"`
	// Members are the IDs of the users in the group. The authenticated user
	// can't add themselves to a group. Ignored for managed groups.
	Members []string `json:"members"`
}

// RolesService handles communication with the role related operations of the
// Axiom API.
//
// Axiom API Reference: /v2/rbac/roles
type RolesService service

// List all roles.
func (s *RolesService) List(ctx context.Context) ([]*Role, error) {
	ctx, span := s.client.trace(ctx, "Roles.List")
	defer span.End()

	var res []*Role
	if err := s.client.Call(ctx, http.MethodGet, s.basePath, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return res, nil
}

// Get a role by id.
func (s *RolesService) Get(ctx context.Context, id string) (*Role, error) {
	ctx, span := s.client.trace(ctx, "Roles.Get", trace.WithAttributes(
		attribute.String("axiom.role_id", id),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, id)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res Role
	if err := s.client.Call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// Create a role with the given properties.
func (s *RolesService) Create(ctx context.Context, req RoleRequest) (*Role, error) {
	ctx, span := s.client.trace(ctx, "Roles.Create", trace.WithAttributes(
		attribute.String("axiom.param.name", req.Name),
	))
	defer span.End()

	var res Role
	if err := s.client.Call(ctx, http.MethodPost, s.basePath, req, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// Update the role identified by the given id with the given properties.
func (s *RolesService) Update(ctx context.Context, id string, req RoleRequest) (*Role, error) {
	ctx, span := s.client.trace(ctx, "Roles.Update", trace.WithAttributes(
		attribute.String("axiom.role_id", id),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, id)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res Role
	if err := s.client.Call(ctx, http.MethodPut, path, req, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// Delete the role identified by the given id. Users whose base role is the
// deleted role are reset to [RoleNone].
func (s *RolesService) Delete(ctx context.Context, id string) error {
	ctx, span := s.client.trace(ctx, "Roles.Delete", trace.WithAttributes(
		attribute.String("axiom.role_id", id),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, id)
	if err != nil {
		return spanError(span, err)
	}

	if err := s.client.Call(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return spanError(span, err)
	}

	return nil
}

// GroupsService handles communication with the group related operations of
// the Axiom API.
//
// Axiom API Reference: /v2/rbac/groups
type GroupsService service

// List all groups.
func (s *GroupsService) List(ctx context.Context) ([]*Group, error) {
	ctx, span := s.client.trace(ctx, "Groups.List")
	defer span.End()

	var res []*Group
	if err := s.client.Call(ctx, http.MethodGet, s.basePath, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return res, nil
}

// Get a group by id.
func (s *GroupsService) Get(ctx context.Context, id string) (*Group, error) {
	ctx, span := s.client.trace(ctx, "Groups.Get", trace.WithAttributes(
		attribute.String("axiom.group_id", id),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, id)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res Group
	if err := s.client.Call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// Create a group with the given properties.
func (s *GroupsService) Create(ctx context.Context, req GroupRequest) (*Group, error) {
	ctx, span := s.client.trace(ctx, "Groups.Create", trace.WithAttributes(
		attribute.String("axiom.param.name", req.Name),
	))
	defer span.End()

	var res Group
	if err := s.client.Call(ctx, http.MethodPost, s.basePath, req, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// Update the group identified by the given id with the given properties.
func (s *GroupsService) Update(ctx context.Context, id string, req GroupRequest) (*Group, error) {
	ctx, span := s.client.trace(ctx, "Groups.Update", trace.WithAttributes(
		attribute.String("axiom.group_id", id),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, id)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res Group
	if err := s.client.Call(ctx, http.MethodPut, path, req, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// Delete the group identified by the given id. Managed groups can't be
// deleted.
func (s *GroupsService) Delete(ctx context.Context, id string) error {
	ctx, span := s.client.trace(ctx, "Groups.Delete", trace.WithAttributes(
		attribute.String("axiom.group_id", id),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, id)
	if err != nil {
		return spanError(span, err)
	}

	if err := s.client.Call(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return spanError(span, err)
	}

	return nil
}
