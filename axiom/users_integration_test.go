package axiom_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// UsersTestSuite tests all methods of the Axiom Users API against a live
// deployment.
type UsersTestSuite struct {
	IntegrationTestSuite
}

func TestUsersTestSuite(t *testing.T) {
	suite.Run(t, new(UsersTestSuite))
}

func (s *UsersTestSuite) Test() {
	user, err := s.client.Users.Current(s.suiteCtx)
	s.Require().NoError(err)
	s.Require().NotNil(user)
}
