package axiom_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
)

// AnnotationsTestSuite tests all methods of the Axiom Annotation API against a
// live deployment.
type AnnotationsTestSuite struct {
	IntegrationTestSuite

	datasetA   *axiom.Dataset
	datasetB   *axiom.Dataset
	annotation *axiom.Annotation
}

func TestAnnotationsTestSuite(t *testing.T) {
	suite.Run(t, new(AnnotationsTestSuite))
}

func (s *AnnotationsTestSuite) SetupSuite() {
	s.IntegrationTestSuite.SetupSuite()
}

func (s *AnnotationsTestSuite) TearDownSuite() {
	s.IntegrationTestSuite.TearDownSuite()
}

func (s *AnnotationsTestSuite) SetupTest() {
	s.IntegrationTestSuite.SetupTest()

	var err error

	s.datasetA, err = s.client.Datasets.Create(s.suiteCtx, axiom.DatasetCreateRequest{
		Name:        "test-axiom-go-annotations-a-" + datasetSuffix,
		Description: "This is a test dataset for annotations integration tests.",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.datasetA)

	s.datasetB, err = s.client.Datasets.Create(s.suiteCtx, axiom.DatasetCreateRequest{
		Name:        "test-axiom-go-annotations-b-" + datasetSuffix,
		Description: "This is a test dataset for annotations integration tests.",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.datasetA)

	s.annotation, err = s.client.Annotations.Create(s.ctx, &axiom.AnnotationCreateRequest{
		Title:    "Test Annotation",
		Datasets: []string{s.datasetA.ID},
		Type:     "deployment",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.annotation)
}

func (s *AnnotationsTestSuite) TearDownTest() {
	// Teardown routines use their own context to avoid not being run at all
	// when the suite gets cancelled or times out.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	err := s.client.Datasets.Delete(ctx, s.datasetA.ID)
	s.NoError(err)

	err = s.client.Datasets.Delete(ctx, s.datasetB.ID)
	s.NoError(err)

	err = s.client.Annotations.Delete(ctx, s.annotation.ID)
	s.NoError(err)

	s.IntegrationTestSuite.TearDownTest()
}

func (s *AnnotationsTestSuite) Test() {
	// Get annotation
	annotation, err := s.client.Annotations.Get(s.ctx, s.annotation.ID)
	s.Require().NoError(err)
	s.Require().Equal(s.annotation.ID, annotation.ID)
	s.Require().Equal(s.annotation.Title, annotation.Title)

	// List annotations without filterr
	annotations, err := s.client.Annotations.List(s.ctx, nil)
	s.Require().NoError(err)
	if s.Len(annotations, 1) {
		s.Equal(s.annotation.ID, annotations[0].ID)
	}

	// List annotations with filter
	annotations, err = s.client.Annotations.List(s.ctx, &axiom.AnnotationsFilter{
		Datasets: []string{s.datasetA.ID},
	})
	s.Require().NoError(err)
	if s.Len(annotations, 1) {
		s.Equal(s.annotation.ID, annotations[0].ID)
	}

	// Update annotation
	_, err = s.client.Annotations.Update(s.ctx, s.annotation.ID, axiom.AnnotationUpdateRequest{
		Datasets: []string{s.datasetB.ID},
	})
	s.Require().NoError(err)

	// List annotations with filter, this should return 0 items now
	annotations, err = s.client.Annotations.List(s.ctx, &axiom.AnnotationsFilter{
		Datasets: []string{s.datasetA.ID},
	})
	s.Require().NoError(err)
	s.Len(annotations, 0)
}
