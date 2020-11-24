// +build integration

package axiom_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go"
)

// UsersTestSuite tests all methods of the Axiom Users API against a live
// deployment.
type UsersTestSuite struct {
	IntegrationTestSuite

	user *axiom.User
}

func TestUsersTestSuite(t *testing.T) {
	suite.Run(t, new(UsersTestSuite))
}

func (s *UsersTestSuite) SetupSuite() {
	s.IntegrationTestSuite.SetupSuite()

	var err error
	s.user, err = s.client.Users.Create(s.suiteCtx, axiom.CreateUserRequest{
		Name:  "John Doe",
		Email: "john.doe@example.com",
		Role:  axiom.RoleAdmin,
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.user)

	// TODO(lukasmalkmus): Have API return an initialized permissions slice
	// (even when empty).
	s.user.Permissions = []string{}
}

func (s *UsersTestSuite) TearDownSuite() {
	// Teardown routines use their own context to avoid not being run at all
	// when the suite gets cancelled or times out.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := s.client.Users.Delete(ctx, s.user.ID)
	s.NoError(err)

	s.IntegrationTestSuite.TearDownSuite()
}

func (s *UsersTestSuite) TestUpdate() {
	s.T().Skip("Enable as soon as the API response has been fixed!")

	user, err := s.client.Users.Update(s.suiteCtx, s.user.ID, axiom.UpdateUserRequest{
		Name: "Johnny Doe",
	})
	s.Require().NoError(err)
	s.Require().NotNil(user)

	s.user = user
}

func (s *UsersTestSuite) TestUpdateRole() {
	s.T().Skip("Enable as soon as the API response has been fixed!")

	user, err := s.client.Users.UpdateRole(s.suiteCtx, s.user.ID, axiom.RoleUser)
	s.Require().NoError(err)
	s.Require().NotNil(user)

	s.user = user
}

func (s *UsersTestSuite) TestGet() {
	user, err := s.client.Users.Get(s.ctx, s.user.ID)
	s.Require().NoError(err)
	s.Require().NotNil(user)

	s.Equal(s.user, user)
}

func (s *UsersTestSuite) TestList() {
	users, err := s.client.Users.List(s.ctx)
	s.Require().NoError(err)
	s.Require().NotNil(users)

	s.Contains(users, s.user)
}
