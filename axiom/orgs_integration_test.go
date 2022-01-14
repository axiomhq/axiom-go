//go:build integration
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
	organizations, err := s.client.Organizations.Selfhost.List(s.ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(organizations)

	// Get the first organization and make sure it is the same organization as
	// in the list call.
	organization, err := s.client.Organizations.Selfhost.Get(s.ctx, organizations[0].ID)
	s.Require().NoError(err)
	s.Require().NotNil(organization)

	s.Contains(organizations, organization)

	// Let's update the organization. The name is not changed, we just want to
	// make sure the call works.
	// HINT(lukasmalkmus): This only works when the authenticated user is an
	// owner. Our CI user isn't, for good reason. Just skip this test for now.
	// organization, err = s.client.Organizations.Selfhost.Update(s.suiteCtx, organization.ID, axiom.OrganizationUpdateRequest{
	// 	Name: organization.Name,
	// })
	// s.Require().NoError(err)
	// s.Require().NotNil(organization)

	// s.Equal(organizations[0].Name, organization.Name)

	// Get the organizations license and make sure it matches the one which is
	// part of the Organization struct.
	license, err := s.client.Organizations.Selfhost.License(s.ctx, organization.ID)
	s.Require().NoError(err)
	s.Require().NotNil(license)

	s.Equal(&organization.License, license)

	// Get the organizations status.
	status, err := s.client.Organizations.Selfhost.Status(s.ctx, organization.ID)
	s.Require().NoError(err)
	s.Require().NotNil(status)

	s.NotEmpty(status)
}

func (s *OrganizationsTestSuite) TestCloud() {
	if !s.isCloud {
		s.T().Skip("Skipping Axiom Cloud integration tests")
	}

	// Create an organization.
	organization, err := s.client.Organizations.Cloud.Create(s.ctx, axiom.OrganizationCreateUpdateRequest{
		Name: "tag-org-" + datasetSuffix, // HINT(lukasmalkmus): Trimmed down to not hit 30 char org limit.
	})
	s.Require().NoError(err)
	s.Require().NotNil(organization)

	s.Equal("tag-org-"+datasetSuffix, organization.Name)

	// Rotate the signing keys on the organization and make sure the new keys
	// are returned.
	oldPrimaryKey, oldSecondaryKey := organization.SigningKeys.Primary, organization.SigningKeys.Secondary
	organization, err = s.client.Organizations.Cloud.RotateSigningKeys(s.ctx, organization.ID)
	s.Require().NoError(err)
	s.Require().NotNil(organization)

	s.NotEqual(oldPrimaryKey, organization.SigningKeys.Primary)
	s.NotEqual(oldSecondaryKey, organization.SigningKeys.Secondary)
	s.NotEqual(oldSecondaryKey, organization.SigningKeys.Primary)
	s.Equal(oldPrimaryKey, organization.SigningKeys.Secondary)

	// Delete the organization.
	err = s.client.Organizations.Cloud.Delete(s.ctx, organization.ID)
	s.Require().NoError(err)
}
