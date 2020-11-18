// +build integration

package axiom_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
)

// VirtualFieldsTestSuite tests all methods of the Axiom Virtual Fields API
// against a live deployment.
type VirtualFieldsTestSuite struct {
	IntegrationTestSuite

	datasetID string

	virtualField *axiom.VirtualField
}

func TestVirtualFieldsTestSuite(t *testing.T) {
	suite.Run(t, new(VirtualFieldsTestSuite))
}

func (s *VirtualFieldsTestSuite) SetupSuite() {
	s.IntegrationTestSuite.SetupSuite()

	dataset, err := s.client.Datasets.Create(s.suiteCtx, axiom.DatasetCreateRequest{
		Name:        "test-" + randString(),
		Description: "This is a test dataset",
	})
	s.Require().NoError(err)
	s.Require().NotNil(dataset)

	s.datasetID = dataset.ID

	s.virtualField, err = s.client.VirtualFields.Create(s.suiteCtx, axiom.VirtualField{
		Dataset:     dataset.ID,
		Name:        "Failed Requests",
		Description: "Statuses >= 400",
		Alias:       "status_failed",
		Expression:  "response >= 400",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.virtualField)
}

func (s *VirtualFieldsTestSuite) TearDownSuite() {
	// Teardown routines use their own context to avoid not being run at all
	// when the suite gets cancelled or times out.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := s.client.VirtualFields.Delete(ctx, s.virtualField.ID)
	s.NoError(err)

	err = s.client.Datasets.Delete(ctx, s.datasetID)
	s.NoError(err)

	s.IntegrationTestSuite.TearDownSuite()
}

func (s *VirtualFieldsTestSuite) Test() {
	// Let's update the virtual field.
	virtualField, err := s.client.VirtualFields.Update(s.suiteCtx, s.virtualField.ID, axiom.VirtualField{
		Dataset:     s.datasetID,
		Name:        "Failed Requests",
		Description: "Statuses > 399",
		Alias:       "status_failed",
		Expression:  "response > 399",
	})
	s.Require().NoError(err)
	s.Require().NotNil(virtualField)

	s.virtualField = virtualField

	// Get the virtual field and make sure it matches what we have updated it
	// to.
	virtualField, err = s.client.VirtualFields.Get(s.ctx, s.virtualField.ID)
	s.Require().NoError(err)
	s.Require().NotNil(virtualField)

	s.Equal(s.virtualField, virtualField)

	// List all virtual fields and make sure the created virtual field is part
	// of that list.
	virtualFields, err := s.client.VirtualFields.List(s.ctx, axiom.VirtualFieldListOptions{
		Dataset: s.datasetID,
	})
	s.Require().NoError(err)
	s.Require().NotNil(virtualFields)

	s.Contains(virtualFields, s.virtualField)
}
