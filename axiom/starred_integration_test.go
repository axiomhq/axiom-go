//go:build integration
// +build integration

package axiom_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/query"
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
		Name:        "test-axiom-go-starred-queries-" + datasetSuffix,
		Description: "This is a test dataset for starred queries integration tests.",
	})
	s.Require().NoError(err)
	s.Require().NotNil(dataset)

	s.datasetID = dataset.ID

	s.starredQuery, err = s.client.StarredQueries.Create(s.suiteCtx, axiom.StarredQuery{
		Kind:    query.Stream,
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

	err = s.client.Datasets.Delete(ctx, s.datasetID)
	s.NoError(err)

	s.IntegrationTestSuite.TearDownSuite()
}

func (s *StarredQueriesTestSuite) Test() {
	// Let's update the starredQuery.
	starredQuery, err := s.client.StarredQueries.Update(s.suiteCtx, s.starredQuery.ID, axiom.StarredQuery{
		Kind:    query.Analytics,
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
	// TODO(lukasmalkmus): This needs a server-side fix.
	// starredQueries, err := s.client.StarredQueries.List(s.ctx, axiom.StarredQueriesListOptions{
	// 	Kind: query.Analytics,
	// })
	// s.Require().NoError(err)
	// s.Require().NotNil(starredQueries)

	// TODO(lukasmalkmus): This needs a server-side fix.
	// s.Contains(starredQueries, s.starredQuery)
}
