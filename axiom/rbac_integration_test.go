package axiom_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
)

// RBACTestSuite tests all methods of the Axiom Roles and Groups APIs against a
// live deployment.
type RBACTestSuite struct {
	IntegrationTestSuite

	// Setup once per test.
	role  *axiom.Role
	group *axiom.Group
}

func TestRBACTestSuite(t *testing.T) {
	suite.Run(t, new(RBACTestSuite))
}

func (s *RBACTestSuite) SetupSuite() {
	s.IntegrationTestSuite.SetupSuite()
}

func (s *RBACTestSuite) TearDownSuite() {
	s.IntegrationTestSuite.TearDownSuite()
}

func (s *RBACTestSuite) SetupTest() {
	s.IntegrationTestSuite.SetupTest()

	var err error
	s.role, err = s.client.Roles.Create(s.ctx, axiom.RoleRequest{
		Name:        "test-role-" + datasetSuffix,
		Description: "Created by the axiom-go integration tests",
		DatasetCapabilities: map[string]axiom.RoleDatasetCapabilities{
			"test-rbac-" + datasetSuffix: {Query: []axiom.Action{axiom.ActionRead}},
		},
	})
	if httpErr := new(axiom.HTTPError); errors.As(err, httpErr) && httpErr.Status == http.StatusForbidden {
		s.T().Skipf("RBAC is not available for this organization: %s", err)
	}
	s.Require().NoError(err)
	s.Require().NotNil(s.role)

	s.group, err = s.client.Groups.Create(s.ctx, axiom.GroupRequest{
		Name:        "test-group-" + datasetSuffix,
		Description: "Created by the axiom-go integration tests",
		Roles:       []string{s.role.ID},
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.group)
}

func (s *RBACTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(s.ctx), time.Second*15)
	defer cancel()

	if s.group != nil {
		err := s.client.Groups.Delete(ctx, s.group.ID)
		s.NoError(err)
	}

	if s.role != nil {
		err := s.client.Roles.Delete(ctx, s.role.ID)
		s.NoError(err)
	}

	s.IntegrationTestSuite.TearDownTest()
}

func (s *RBACTestSuite) TestRoles() {
	role, err := s.client.Roles.Update(s.ctx, s.role.ID, axiom.RoleRequest{
		Name:        s.role.Name,
		Description: "Updated by the axiom-go integration tests",
		OrgCapabilities: axiom.RoleOrgCapabilities{
			Dashboards: []axiom.Action{axiom.ActionRead},
		},
	})
	s.Require().NoError(err)
	s.Require().NotNil(role)

	s.Equal(s.role.ID, role.ID)
	s.Equal("Updated by the axiom-go integration tests", role.Description)
	s.Empty(role.DatasetCapabilities)
	s.Equal([]axiom.Action{axiom.ActionRead}, role.OrgCapabilities.Dashboards)

	role, err = s.client.Roles.Get(s.ctx, s.role.ID)
	s.Require().NoError(err)
	s.Require().NotNil(role)

	s.Equal("Updated by the axiom-go integration tests", role.Description)

	roles, err := s.client.Roles.List(s.ctx)
	s.Require().NoError(err)
	s.Contains(roles, role)
}

func (s *RBACTestSuite) TestGroups() {
	s.Equal([]string{s.role.ID}, s.group.Roles)
	s.False(s.group.IsManaged)

	group, err := s.client.Groups.Update(s.ctx, s.group.ID, axiom.GroupRequest{
		Name:        s.group.Name,
		Description: "Updated by the axiom-go integration tests",
	})
	s.Require().NoError(err)
	s.Require().NotNil(group)

	s.Equal(s.group.ID, group.ID)
	s.Equal("Updated by the axiom-go integration tests", group.Description)
	s.Empty(group.Roles)

	group, err = s.client.Groups.Get(s.ctx, s.group.ID)
	s.Require().NoError(err)
	s.Require().NotNil(group)

	s.Equal("Updated by the axiom-go integration tests", group.Description)

	groups, err := s.client.Groups.List(s.ctx)
	s.Require().NoError(err)
	s.Contains(groups, group)
}

func (s *RBACTestSuite) TestDeleteRoleRemovesItFromGroups() {
	roleID := s.role.ID
	err := s.client.Roles.Delete(s.ctx, roleID)
	s.Require().NoError(err)
	s.role = nil

	_, err = s.client.Roles.Get(s.ctx, roleID)
	s.ErrorIs(err, axiom.ErrNotFound)

	group, err := s.client.Groups.Get(s.ctx, s.group.ID)
	s.Require().NoError(err)
	s.Empty(group.Roles)
}
