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

	// Setup once per test.
	vfield *axiom.VirtualFieldWithID
}

func TestVirtualFieldsTestSuite(t *testing.T) {
	suite.Run(t, new(VirtualFieldsTestSuite))
}

func (s *VirtualFieldsTestSuite) SetupSuite() {
	s.IntegrationTestSuite.SetupSuite()
}

func (s *VirtualFieldsTestSuite) TearDownSuite() {
	s.IntegrationTestSuite.TearDownSuite()
}

func (s *VirtualFieldsTestSuite) SetupTest() {
	s.IntegrationTestSuite.SetupTest()

	var err error
	s.vfield, err = s.client.VirtualFields.Create(s.ctx, axiom.VirtualField{
		Dataset:    "test-dataset",
		Name:       "TestField",
		Expression: "a + b",
		Type:       "number",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.vfield)
}

func (s *VirtualFieldsTestSuite) TearDownTest() {
	// Teardown routines use their own context to avoid not being run at all
	// when the suite gets cancelled or times out.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(s.ctx), time.Second*15)
	defer cancel()

	err := s.client.VirtualFields.Delete(ctx, s.vfield.ID)
	s.NoError(err)

	s.IntegrationTestSuite.TearDownTest()
}

func (s *VirtualFieldsTestSuite) Test() {
	// Update the virtual field.
	vfield, err := s.client.VirtualFields.Update(s.ctx, s.vfield.ID, axiom.VirtualField{
		Dataset:    "test-dataset",
		Name:       "UpdatedTestField",
		Expression: "a - b",
		Type:       "number",
	})
	s.Require().NoError(err)
	s.Require().NotNil(vfield)

	s.vfield = vfield

	// Get the virtual field and make sure it matches the updated values.
	vfield, err = s.client.VirtualFields.Get(s.ctx, s.vfield.ID)
	s.Require().NoError(err)
	s.Require().NotNil(vfield)

	s.Equal(s.vfield, vfield)

	// List all virtual fields for the dataset and ensure the created field is part of the list.
	vfields, err := s.client.VirtualFields.List(s.ctx, "test-dataset")
	s.Require().NoError(err)
	s.Require().NotEmpty(vfields)

	s.Contains(vfields, s.vfield)
}

func (s *VirtualFieldsTestSuite) TestCreateAndDeleteVirtualField() {
	// Create a new virtual field.
	vfield, err := s.client.VirtualFields.Create(s.ctx, axiom.VirtualField{
		Dataset:    "test-dataset",
		Name:       "NewTestField",
		Expression: "x * y",
		Type:       "number",
	})
	s.Require().NoError(err)
	s.Require().NotNil(vfield)

	// Get the virtual field and ensure it matches what was created.
	fetchedField, err := s.client.VirtualFields.Get(s.ctx, vfield.ID)
	s.Require().NoError(err)
	s.Require().NotNil(fetchedField)
	s.Equal(vfield, fetchedField)

	// Delete the virtual field.
	err = s.client.VirtualFields.Delete(s.ctx, vfield.ID)
	s.Require().NoError(err)

	// Ensure the virtual field no longer exists.
	_, err = s.client.VirtualFields.Get(s.ctx, vfield.ID)
	s.Error(err)
}

func (s *VirtualFieldsTestSuite) TestListVirtualFields() {
	// List all virtual fields for the dataset and ensure the created field is part of the list.
	vfields, err := s.client.VirtualFields.List(s.ctx, "test-dataset")
	s.Require().NoError(err)
	s.Require().NotEmpty(vfields)

	s.Contains(vfields, s.vfield)
}
