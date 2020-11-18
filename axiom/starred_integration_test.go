// +build integration

package axiom_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
)

// StarredQueriesTestSuite tests all methods of the Axiom StarredQueries API against a
// live deployment.
type StarredQueriesTestSuite struct {
	IntegrationTestSuite

	datasetID string

	starredQuery *axiom.StarredQuery
}

func TestStarredQueriesTestSuite(t *testing.T) {
	suite.Run(t, new(StarredQueriesTestSuite))
}

func (s *StarredQueriesTestSuite) SetupSuite() {
	s.IntegrationTestSuite.SetupSuite()

	dataset, err := s.client.Datasets.Create(s.suiteCtx, axiom.DatasetCreateRequest{
		Name:        "test-" + randString(),
		Description: "This is a test dataset",
	})
	s.Require().NoError(err)
	s.Require().NotNil(dataset)

	s.datasetID = dataset.ID

	s.starredQuery, err = s.client.StarredQueries.Create(s.suiteCtx, axiom.StarredQuery{
		Kind:    axiom.Stream,
		Dataset: dataset.ID,
		Name:    "Test Query",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.starredQuery)
}

func (s *StarredQueriesTestSuite) TearDownSuite() {
	// Teardown routines use their own context to avoid not being run at all
	// when the suite gets cancelled or times out.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := s.client.StarredQueries.Delete(ctx, s.starredQuery.ID)
	s.NoError(err)

	s.IntegrationTestSuite.TearDownSuite()
}

func (s *StarredQueriesTestSuite) Test() {
	// Let's update the starredQuery.
	starredQuery, err := s.client.StarredQueries.Update(s.suiteCtx, s.starredQuery.ID, axiom.StarredQuery{
		Kind:    axiom.Analytics,
		Dataset: s.datasetID,
		Name:    "Updated Test Query",
	})
	s.Require().NoError(err)
	s.Require().NotNil(starredQuery)

	s.starredQuery = starredQuery

	// Get the starred query and make sure it matches what we have updated it
	// to.
	starredQuery, err = s.client.StarredQueries.Get(s.ctx, s.starredQuery.ID)
	s.Require().NoError(err)
	s.Require().NotNil(starredQuery)

	s.Equal(s.starredQuery, starredQuery)

	// List all starred queries and make sure the created starred query is part
	// of that list.
	starredQueries, err := s.client.StarredQueries.List(s.ctx, axiom.StarredQueriesListOptions{
		Kind: axiom.Analytics,
	})
	s.Require().NoError(err)
	s.Require().NotNil(starredQueries)

	s.Contains(starredQueries, s.starredQuery)
}
