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

	keys, err := s.client.Organizations.ViewSigningKeys(s.ctx, organization.ID)
	s.Require().NoError(err)
	s.Require().NotNil(keys)

	s.NotEmpty(keys.Primary)
	s.NotEmpty(keys.Secondary)
	s.NotEqual(keys.Primary, keys.Secondary)

	// Rotate the signing keys on the organization and make sure the new keys
	// are returned.
	oldPrimaryKey, oldSecondaryKey := keys.Primary, keys.Secondary
	keys, err = s.client.Organizations.RotateSigningKeys(s.ctx, organization.ID)
	s.Require().NoError(err)
	s.Require().NotNil(keys)

	s.NotEqual(oldPrimaryKey, keys.Primary)
	s.NotEqual(oldSecondaryKey, keys.Secondary)
	s.NotEqual(oldSecondaryKey, keys.Primary)
	s.Equal(oldPrimaryKey, keys.Secondary)
}
