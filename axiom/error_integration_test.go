//go:build integration
// +build integration

package axiom_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
)

// ErrorTestSuite implements a base test suite for integration tests.
type ErrorTestSuite struct {
	IntegrationTestSuite
}

func TestErrorTestSuite(t *testing.T) {
	suite.Run(t, new(ErrorTestSuite))
}

func (s *ErrorTestSuite) Test() {
	invalidDatasetName := "test-axiom-go-error-" + datasetSuffix

	_, err := s.client.Datasets.Info(s.ctx, invalidDatasetName)
	s.Require().ErrorIs(err, axiom.ErrNotFound)

	// Set invalid credentials...
	err = s.client.Options(axiom.SetAccessToken("xapt-123"))
	s.Require().NoError(err)

	// ...and see the same request fail with a different (unauthenticated)
	// error.
	_, err = s.client.Datasets.Info(s.ctx, invalidDatasetName)
	s.Require().ErrorIs(err, axiom.ErrUnauthenticated)
}
