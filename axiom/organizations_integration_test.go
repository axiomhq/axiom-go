// +build integration

package axiom_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
)

// OrganizationsTestSuite tests all methods of the Axiom Organizations API
// against a live deployment.
type OrganizationsTestSuite struct {
	IntegrationTestSuite
}

func TestOrganizationsTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationsTestSuite))
}

func (s *OrganizationsTestSuite) Test() {
	// List all organizations.
	organizations, err := s.client.Organizations.List(s.ctx)
	s.Require().NoError(err)
	s.Require().NotNil(organizations)
	s.Require().NotEmpty(organizations)

	// Get the first organization and make sure it is the same organization as
	// in the list call.
	organization, err := s.client.Organizations.Get(s.ctx, organizations[0].ID)
	s.Require().NoError(err)
	s.Require().NotNil(organization)

	s.Contains(organizations, organization)

	// Get the organizations license and make sure it matches the one which is
	// part of the Organization struct.
	license, err := s.client.Organizations.License(s.ctx, organization.ID)
	s.Require().NoError(err)
	s.Require().NotNil(license)

	s.Equal(&organization.License, license)

	// Let's update the organization. The name is not changed, we just want to
	// make sure the call works.
	organization, err = s.client.Organizations.Update(s.suiteCtx, organization.ID, axiom.OrganizationUpdateRequest{
		Name: organization.Name,
	})
	s.Require().NoError(err)
	s.Require().NotNil(organization)

	s.Equal(organizations[0].Name, organization.Name)
}
