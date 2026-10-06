package axiom

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testRoleJSON = `{
	"id": "role-id",
	"name": "team-a-readers",
	"description": "Query team A datasets",
	"members": ["user-1"],
	"datasetCapabilities": {
		"team-a-logs": {
			"query": ["read"]
		}
	},
	"viewCapabilities": {
		"team-a-view": {
			"query": ["read"],
			"share": ["read"]
		}
	},
	"orgCapabilities": {
		"dashboards": ["read"],
		"auditLog": ["read"]
	}
}`

var testRole = &Role{
	ID:          "role-id",
	Name:        "team-a-readers",
	Description: "Query team A datasets",
	Members:     []string{"user-1"},
	DatasetCapabilities: map[string]RoleDatasetCapabilities{
		"team-a-logs": {Query: []Action{ActionRead}},
	},
	ViewCapabilities: map[string]RoleViewCapabilities{
		"team-a-view": {
			Query: []Action{ActionRead},
			Share: []Action{ActionRead},
		},
	},
	OrgCapabilities: RoleOrgCapabilities{
		Dashboards: []Action{ActionRead},
		AuditLog:   []Action{ActionRead},
	},
}

var testRoleRequest = RoleRequest{
	Name:        "team-a-readers",
	Description: "Query team A datasets",
	DatasetCapabilities: map[string]RoleDatasetCapabilities{
		"team-a-logs": {Query: []Action{ActionRead}},
	},
	OrgCapabilities: RoleOrgCapabilities{
		Dashboards: []Action{ActionRead},
	},
}

const testRoleRequestJSON = `{
	"name": "team-a-readers",
	"description": "Query team A datasets",
	"datasetCapabilities": {
		"team-a-logs": {
			"query": ["read"]
		}
	},
	"orgCapabilities": {
		"dashboards": ["read"]
	}
}`

const testGroupJSON = `{
	"id": "group-id",
	"name": "team-a",
	"description": "Team A",
	"roles": ["role-id"],
	"members": ["user-1", "user-2"],
	"isManaged": false
}`

var testGroup = &Group{
	ID:          "group-id",
	Name:        "team-a",
	Description: "Team A",
	Roles:       []string{"role-id"},
	Members:     []string{"user-1", "user-2"},
}

var testGroupRequest = GroupRequest{
	Name:        "team-a",
	Description: "Team A",
	Roles:       []string{"role-id"},
	Members:     []string{"user-1", "user-2"},
}

const testGroupRequestJSON = `{
	"name": "team-a",
	"description": "Team A",
	"roles": ["role-id"],
	"members": ["user-1", "user-2"]
}`

func TestRolesService_List(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, "["+testRoleJSON+"]")
		assert.NoError(t, err)
	}
	client := setup(t, "GET /v2/rbac/roles", hf)

	res, err := client.Roles.List(t.Context())
	require.NoError(t, err)

	assert.Equal(t, []*Role{testRole}, res)
}

func TestRolesService_Get(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, testRoleJSON)
		assert.NoError(t, err)
	}
	client := setup(t, "GET /v2/rbac/roles/role-id", hf)

	res, err := client.Roles.Get(t.Context(), "role-id")
	require.NoError(t, err)

	assert.Equal(t, testRole, res)
}

func TestRolesService_Create(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, mediaTypeJSON, r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.JSONEq(t, testRoleRequestJSON, string(body))

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err = fmt.Fprint(w, testRoleJSON)
		assert.NoError(t, err)
	}
	client := setup(t, "POST /v2/rbac/roles", hf)

	res, err := client.Roles.Create(t.Context(), testRoleRequest)
	require.NoError(t, err)

	assert.Equal(t, testRole, res)
}

func TestRolesService_Update(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, mediaTypeJSON, r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.JSONEq(t, testRoleRequestJSON, string(body))

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err = fmt.Fprint(w, testRoleJSON)
		assert.NoError(t, err)
	}
	client := setup(t, "PUT /v2/rbac/roles/role-id", hf)

	res, err := client.Roles.Update(t.Context(), "role-id", testRoleRequest)
	require.NoError(t, err)

	assert.Equal(t, testRole, res)
}

func TestRolesService_Delete(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}
	client := setup(t, "DELETE /v2/rbac/roles/role-id", hf)

	err := client.Roles.Delete(t.Context(), "role-id")
	require.NoError(t, err)
}

func TestGroupsService_List(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, "["+testGroupJSON+"]")
		assert.NoError(t, err)
	}
	client := setup(t, "GET /v2/rbac/groups", hf)

	res, err := client.Groups.List(t.Context())
	require.NoError(t, err)

	assert.Equal(t, []*Group{testGroup}, res)
}

func TestGroupsService_Get(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, testGroupJSON)
		assert.NoError(t, err)
	}
	client := setup(t, "GET /v2/rbac/groups/group-id", hf)

	res, err := client.Groups.Get(t.Context(), "group-id")
	require.NoError(t, err)

	assert.Equal(t, testGroup, res)
}

func TestGroupsService_Create(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, mediaTypeJSON, r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.JSONEq(t, testGroupRequestJSON, string(body))

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err = fmt.Fprint(w, testGroupJSON)
		assert.NoError(t, err)
	}
	client := setup(t, "POST /v2/rbac/groups", hf)

	res, err := client.Groups.Create(t.Context(), testGroupRequest)
	require.NoError(t, err)

	assert.Equal(t, testGroup, res)
}

func TestGroupsService_Update(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, mediaTypeJSON, r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.JSONEq(t, testGroupRequestJSON, string(body))

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err = fmt.Fprint(w, testGroupJSON)
		assert.NoError(t, err)
	}
	client := setup(t, "PUT /v2/rbac/groups/group-id", hf)

	res, err := client.Groups.Update(t.Context(), "group-id", testGroupRequest)
	require.NoError(t, err)

	assert.Equal(t, testGroup, res)
}

func TestGroupsService_Delete(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}
	client := setup(t, "DELETE /v2/rbac/groups/group-id", hf)

	err := client.Groups.Delete(t.Context(), "group-id")
	require.NoError(t, err)
}
