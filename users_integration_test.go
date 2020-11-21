// +build integration

package axiom_test

import (
	"testing"

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
	err := s.client.Users.Delete(s.suiteCtx, s.user.ID)
	s.Require().NoError(err)

	s.IntegrationTestSuite.TearDownSuite()
}

func (s *UsersTestSuite) TestUsers() {
	// TODO(lukasmalkmus): Add Update test case.

	user, err := s.client.Users.Get(s.ctx, s.user.ID)
	s.Require().NoError(err)
	s.Require().NotNil(user)

	s.Equal(s.user, user)

	// TODO(lukasmalkmus): Enable if we finally support limiting.
	// users, err := s.client.Users.List(s.ctx, axiom.ListOptions{Limit: 1})
	// s.Require().NoError(err)
	// s.Require().NotNil(users)

	// s.Len(users, 1)

	users, err := s.client.Users.List(s.ctx, axiom.ListOptions{})
	s.Require().NoError(err)
	s.Require().NotNil(users)

	s.Contains(users, user)
}
