//go:build integration

package axiom_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
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
	s.Require().NotEmpty(organizations)

	// Get the first organization and make sure it is the same organization as
	// in the list call.
	organization, err := s.client.Organizations.Get(s.ctx, organizations[0].ID)
	s.Require().NoError(err)
	s.Require().NotNil(organization)

	s.Contains(organizations, organization)
}
