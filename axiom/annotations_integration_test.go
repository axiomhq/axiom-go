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

	// Setup once per test.
	datasetA         *axiom.Dataset
	datasetB         *axiom.Dataset
	annotation       *axiom.Annotation
	rangedAnnotation *axiom.Annotation
}

func TestAnnotationsTestSuite(t *testing.T) {
	suite.Run(t, new(AnnotationsTestSuite))
}

func (s *AnnotationsTestSuite) SetupTest() {
	s.IntegrationTestSuite.SetupTest()

	var err error

	s.datasetA, err = s.client.Datasets.Create(s.ctx, axiom.DatasetCreateRequest{
		Name:        "test-axiom-go-annotations-a-" + datasetSuffix,
		Description: "This is a test dataset for annotations integration tests.",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.datasetA)

	s.datasetB, err = s.client.Datasets.Create(s.ctx, axiom.DatasetCreateRequest{
		Name:        "test-axiom-go-annotations-b-" + datasetSuffix,
		Description: "This is a test dataset for annotations integration tests.",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.datasetB)

	s.annotation, err = s.client.Annotations.Create(s.ctx, &axiom.AnnotationCreateRequest{
		Title:    "Test Annotation",
		Datasets: []string{s.datasetA.ID},
		Type:     "deployment",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.annotation)

	s.rangedAnnotation, err = s.client.Annotations.Create(s.ctx, &axiom.AnnotationCreateRequest{
		Title:    "Test Ranged Annotation",
		Datasets: []string{s.datasetB.ID},
		Type:     "maintenance",
		Time:     s.annotation.Time.Add(-2 * time.Hour),
		EndTime:  s.annotation.Time.Add(-time.Hour),
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.rangedAnnotation)
}

func (s *AnnotationsTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(s.ctx), time.Second*15)
	defer cancel()

	if s.annotation != nil {
		err := s.client.Annotations.Delete(ctx, s.annotation.ID)
		s.NoError(err)
	}

	if s.rangedAnnotation != nil {
		err := s.client.Annotations.Delete(ctx, s.rangedAnnotation.ID)
		s.NoError(err)
	}

	if s.datasetA != nil {
		err := s.client.Datasets.Delete(ctx, s.datasetA.ID)
		s.NoError(err)
	}

	if s.datasetB != nil {
		err := s.client.Datasets.Delete(ctx, s.datasetB.ID)
		s.NoError(err)
	}

	s.IntegrationTestSuite.TearDownTest()
}

func (s *AnnotationsTestSuite) Test() {
	// Get annotation.
	annotation, err := s.client.Annotations.Get(s.ctx, s.annotation.ID)
	s.Require().NoError(err)
	s.Require().Equal(s.annotation.ID, annotation.ID)
	s.Require().Equal(s.annotation.Title, annotation.Title)

	// List annotations without filter.
	annotations, err := s.client.Annotations.List(s.ctx, nil)
	s.Require().NoError(err)
	s.Greater(len(annotations), 0)

	// List annotations with filter.
	annotations, err = s.client.Annotations.List(s.ctx, &axiom.AnnotationsFilter{
		Datasets: []string{s.datasetA.ID},
	})
	s.Require().NoError(err)
	if s.Len(annotations, 1) {
		s.Equal(s.annotation.ID, annotations[0].ID)
	}

	// List annotations of multiple datasets.
	annotations, err = s.client.Annotations.List(s.ctx, &axiom.AnnotationsFilter{
		Datasets: []string{s.datasetA.ID, s.datasetB.ID},
	})
	s.Require().NoError(err)
	s.ElementsMatch([]string{s.annotation.ID, s.rangedAnnotation.ID}, annotationIDs(annotations))

	// List annotations in a time range that covers both annotations.
	annotations, err = s.client.Annotations.List(s.ctx, &axiom.AnnotationsFilter{
		Datasets: []string{s.datasetA.ID, s.datasetB.ID},
		Start:    s.annotation.Time.Add(-3 * time.Hour),
		End:      s.annotation.Time.Add(time.Minute),
	})
	s.Require().NoError(err)
	s.ElementsMatch([]string{s.annotation.ID, s.rangedAnnotation.ID}, annotationIDs(annotations))

	// List annotations in a time range that only covers the ranged annotation.
	annotations, err = s.client.Annotations.List(s.ctx, &axiom.AnnotationsFilter{
		Datasets: []string{s.datasetA.ID, s.datasetB.ID},
		Start:    s.annotation.Time.Add(-3 * time.Hour),
		End:      s.annotation.Time.Add(-90 * time.Minute),
	})
	s.Require().NoError(err)
	s.ElementsMatch([]string{s.rangedAnnotation.ID}, annotationIDs(annotations))

	// Update annotation.
	_, err = s.client.Annotations.Update(s.ctx, s.annotation.ID, &axiom.AnnotationUpdateRequest{
		Datasets: []string{s.datasetB.ID},
	})
	s.Require().NoError(err)

	// List annotations with filter, this should return 0 items now.
	annotations, err = s.client.Annotations.List(s.ctx, &axiom.AnnotationsFilter{
		Datasets: []string{s.datasetA.ID},
	})
	s.Require().NoError(err)
	s.Len(annotations, 0)
}

func annotationIDs(annotations []*axiom.Annotation) []string {
	ids := make([]string, len(annotations))
	for i, annotation := range annotations {
		ids[i] = annotation.ID
	}
	return ids
}
