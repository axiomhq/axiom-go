// +build integration

package axiom_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
)

// TeamsTestSuite tests all methods of the Axiom Teams API against a
// live deployment.
type TeamsTestSuite struct {
	IntegrationTestSuite

	team *axiom.Team
}

func TestTeamsTestSuite(t *testing.T) {
	suite.Run(t, new(TeamsTestSuite))
}

func (s *TeamsTestSuite) SetupSuite() {
	s.IntegrationTestSuite.SetupSuite()

	var err error
	s.team, err = s.client.Teams.Create(s.suiteCtx, axiom.TeamCreateRequest{
		Name: "Test Team",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.team)
}

func (s *TeamsTestSuite) TearDownSuite() {
	// Teardown routines use their own context to avoid not being run at all
	// when the suite gets cancelled or times out.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := s.client.Teams.Delete(ctx, s.team.ID)
	s.NoError(err)

	s.IntegrationTestSuite.TearDownSuite()
}

func (s *TeamsTestSuite) TestUpdate() {
	s.T().Skip("Enable as soon as the API param and body ID check has been fixed!")

	team, err := s.client.Teams.Update(s.suiteCtx, s.team.ID, axiom.Team{
		Name: "Updated Test Team",
		// TODO(lukasmalkmus): Probably add user an dataset.
	})
	s.Require().NoError(err)
	s.Require().NotNil(team)

	s.team = team
}

func (s *TeamsTestSuite) TestGet() {
	team, err := s.client.Teams.Get(s.ctx, s.team.ID)
	s.Require().NoError(err)
	s.Require().NotNil(team)

	s.Equal(s.team, team)
}

func (s *TeamsTestSuite) TestList() {
	teams, err := s.client.Teams.List(s.ctx)
	s.Require().NoError(err)
	s.Require().NotNil(teams)

	s.Contains(teams, s.team)
}
