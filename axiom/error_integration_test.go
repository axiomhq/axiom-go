//go:build integration
// +build integration

package axiom_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
)

// ErrorTestSuite tests that the Axiom API returns proper errors against a live
// deployment.
type ErrorTestSuite struct {
	IntegrationTestSuite
}

func TestErrorTestSuite(t *testing.T) {
	suite.Run(t, new(ErrorTestSuite))
}

func (s *ErrorTestSuite) Test() {
	invalidDatasetName := "test-axiom-go-error-" + datasetSuffix

	_, err := s.client.Datasets.Info(s.ctx, invalidDatasetName)
	s.Require().Error(err)
	s.Require().ErrorIs(err, axiom.ErrNotFound)

	// Set invalid credentials...
	err = s.client.Options(axiom.SetAccessToken("xapt-123"))
	s.Require().NoError(err)

	// ...and see the same request fail with a different error
	// (unauthenticated).
	_, err = s.client.Datasets.Info(s.ctx, invalidDatasetName)
	s.Require().Error(err)
	s.Require().ErrorIs(err, axiom.ErrUnauthenticated)
}
