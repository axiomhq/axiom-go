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

	dataset string
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

	s.dataset = "vfield-ds-" + datasetSuffix
	var err error
	_, err = s.client.Datasets.Create(s.ctx, axiom.DatasetCreateRequest{Name: s.dataset})
	s.Require().NoError(err)

	s.vfield, err = s.client.VirtualFields.Create(s.ctx, axiom.VirtualField{
		Dataset:    s.dataset,
		Name:       "TestField",
		Expression: "a + b",
		Type:       "number",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.vfield)
}

func (s *VirtualFieldsTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(s.ctx), time.Second*15)
	defer cancel()

	if s.vfield != nil {
		err := s.client.VirtualFields.Delete(ctx, s.vfield.ID)
		s.NoError(err)
	}

	if s.dataset != "" {
		err := s.client.Datasets.Delete(ctx, s.dataset)
		s.NoError(err)
	}

	s.IntegrationTestSuite.TearDownTest()
}

func (s *VirtualFieldsTestSuite) TestUpdateAndDeleteVirtualField() {
	// Create a new virtual field.
	vfield, err := s.client.VirtualFields.Update(s.ctx, s.vfield.ID, axiom.VirtualField{
		Dataset:    s.dataset,
		Name:       "UpdatedTestField",
		Expression: "a * b",
		Type:       "number",
	})
	s.Require().NoError(err)
	s.Require().NotNil(vfield)

	// Get the virtual field and ensure it matches what was created.
	fetchedField, err := s.client.VirtualFields.Get(s.ctx, vfield.ID)
	s.Require().NoError(err)
	s.Require().NotNil(fetchedField)
	s.Equal(vfield, fetchedField)
}
