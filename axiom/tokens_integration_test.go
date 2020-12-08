// +build integration

package axiom_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
)

type serviceMode uint8

const (
	modeIngest serviceMode = iota + 1
	modePersonal
)

// TokensTestSuite tests all methods of the Axiom Tokens API against a live
// deployment.
type TokensTestSuite struct {
	IntegrationTestSuite

	service *axiom.TokensService
	mode    serviceMode

	token *axiom.Token
}

func TestTokensTestSuite(t *testing.T) {
	suite.Run(t, &TokensTestSuite{
		mode: modeIngest,
	})
	suite.Run(t, &TokensTestSuite{
		mode: modePersonal,
	})
}

func (s *TokensTestSuite) SetupSuite() {
	s.IntegrationTestSuite.SetupSuite()

	switch s.mode {
	case modeIngest:
		s.service = s.client.Tokens.Ingest
	case modePersonal:
		s.service = s.client.Tokens.Personal
	default:
		s.Require().Fail("invalid service mode")
	}

	var err error
	s.token, err = s.service.Create(s.suiteCtx, axiom.TokenCreateRequest{
		Name:        "Test",
		Description: "A test token",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.token)
}

func (s *TokensTestSuite) TearDownSuite() {
	// Teardown routines use their own context to avoid not being run at all
	// when the suite gets cancelled or times out.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := s.service.Delete(ctx, s.token.ID)
	s.NoError(err)

	s.IntegrationTestSuite.TearDownSuite()
}

func (s *TokensTestSuite) Update() {
	// Let's update the token.
	token, err := s.service.Update(s.suiteCtx, s.token.ID, axiom.Token{
		Name:        "Test",
		Description: "A very good test token",
	})
	s.Require().NoError(err)
	s.Require().NotNil(token)

	s.token = token

	// Get the token and make sure it matches what we have updated it to.
	token, err = s.service.Get(s.ctx, s.token.ID)
	s.Require().NoError(err)
	s.Require().NotNil(token)

	s.Equal(s.token, token)

	// Let's get the raw token string and make sure it has the same scopes as
	// the token entity.
	rawToken, err := s.service.View(s.ctx, s.token.ID)
	s.Require().NoError(err)
	s.Require().NotNil(rawToken)

	s.NotEmpty(rawToken.Token)
	s.Equal(s.token.Scopes, rawToken.Scopes)

	// List all tokens and make sure the created token is part of that list.
	tokens, err := s.service.List(s.ctx)
	s.Require().NoError(err)
	s.Require().NotNil(tokens)

	s.Contains(tokens, s.token)
}
