package axiom_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
)

// TokensTestSuite tests all methods of the Axiom Tokens API against a
// live deployment.
type TokensTestSuite struct {
	IntegrationTestSuite

	apiToken *axiom.APIToken
}

func TestTokensTestSuite(t *testing.T) {
	suite.Run(t, new(TokensTestSuite))
}

func (s *TokensTestSuite) SetupTest() {
	s.IntegrationTestSuite.SetupTest()

	createdToken, err := s.client.Tokens.Create(s.suiteCtx, axiom.CreateTokenRequest{
		Name:      "Test token",
		ExpiresAt: time.Now().Add(time.Hour * 24),
		DatasetCapabilities: map[string]axiom.DatasetCapabilities{
			"*": {Ingest: []axiom.Action{axiom.ActionCreate}}},
		OrganisationCapabilities: axiom.OrganisationCapabilities{
			Users: []axiom.Action{axiom.ActionCreate, axiom.ActionRead, axiom.ActionUpdate, axiom.ActionDelete},
		}})
	s.Require().NoError(err)
	s.Require().NotNil(createdToken)

	s.apiToken = &createdToken.APIToken
}

func (s *TokensTestSuite) TearDownTest() {
	// Teardown routines use their own context to avoid not being run at all
	// when the suite gets cancelled or times out.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	if s.apiToken != nil {
		err := s.client.Tokens.Delete(ctx, s.apiToken.ID)
		s.NoError(err)
	}

	s.IntegrationTestSuite.TearDownTest()
}

func (s *TokensTestSuite) Test() {
	// Get the token and make sure it matches what we have updated it to.
	token, err := s.client.Tokens.Get(s.ctx, s.apiToken.ID)
	s.Require().NoError(err)
	s.Require().NotNil(token)

	s.Equal(s.apiToken, token)

	// List all tokens and make sure the created token is part of that
	// list.
	tokens, err := s.client.Tokens.List(s.ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(tokens)

	s.Contains(tokens, s.apiToken)

	// Regenerate the token and make sure the new token is part of the list.
	regeneratedToken, err := s.client.Tokens.Regenerate(s.ctx, s.apiToken.ID, axiom.RegenerateTokenRequest{
		ExistingTokenExpiresAt: time.Now(),
		NewTokenExpiresAt:      time.Now().Add(time.Hour * 24),
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(tokens)

	oldToken := s.apiToken
	s.apiToken = &regeneratedToken.APIToken

	// List all tokens and make sure the created token is part of that
	// list.
	tokens, err = s.client.Tokens.List(s.ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(tokens)

	s.NotContains(tokens, oldToken)
	s.Contains(tokens, &regeneratedToken.APIToken)
}
