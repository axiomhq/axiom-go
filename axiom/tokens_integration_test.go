//go:build integration
// +build integration

package axiom_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
)

// APITokensTestSuite tests all methods of the Axiom API Tokens API against a
// live deployment.
type APITokensTestSuite struct {
	IntegrationTestSuite

	token *axiom.Token
}

func TestAPITokensTestSuite(t *testing.T) {
	suite.Run(t, &APITokensTestSuite{})
}

func (s *APITokensTestSuite) SetupSuite() {
	s.IntegrationTestSuite.SetupSuite()

	var err error
	s.token, err = s.client.Tokens.API.Create(s.suiteCtx, axiom.TokenCreateUpdateRequest{
		Name:        "Test",
		Description: "A test token",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.token)
}

func (s *APITokensTestSuite) TearDownSuite() {
	// Teardown routines use their own context to avoid not being run at all
	// when the suite gets cancelled or times out.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := s.client.Tokens.API.Delete(ctx, s.token.ID)
	s.NoError(err)

	s.IntegrationTestSuite.TearDownSuite()
}

func (s *APITokensTestSuite) Test() {
	// Let's update the token.
	token, err := s.client.Tokens.API.Update(s.suiteCtx, s.token.ID, axiom.TokenCreateUpdateRequest{
		Name:        "Test",
		Description: "A very good test token",
		Scopes:      []string{"*"},
	})
	s.Require().NoError(err)
	s.Require().NotNil(token)

	s.token = token

	// Get the token and make sure it matches what we have updated it to.
	token, err = s.client.Tokens.API.Get(s.ctx, s.token.ID)
	s.Require().NoError(err)
	s.Require().NotNil(token)

	s.Equal(s.token, token)

	// Let's get the raw token string and make sure it has the same scopes as
	// the token entity.
	rawToken, err := s.client.Tokens.API.View(s.ctx, s.token.ID)
	s.Require().NoError(err)
	s.Require().NotNil(rawToken)

	s.NotEmpty(rawToken.Token)
	s.Equal(s.token.Scopes, rawToken.Scopes)

	// List all tokens and make sure the created token is part of that list.
	tokens, err := s.client.Tokens.API.List(s.ctx)
	s.Require().NoError(err)
	s.Require().NotNil(tokens)

	s.Contains(tokens, s.token)
}

// IngestTokensTestSuite tests all methods of the Axiom Ingest Tokens API
// against a live deployment.
type IngestTokensTestSuite struct {
	IntegrationTestSuite

	token *axiom.Token
}

func TestIngestTokensTestSuite(t *testing.T) {
	suite.Run(t, &IngestTokensTestSuite{})
}

func (s *IngestTokensTestSuite) SetupSuite() {
	s.IntegrationTestSuite.SetupSuite()

	var err error
	s.token, err = s.client.Tokens.Ingest.Create(s.suiteCtx, axiom.TokenCreateUpdateRequest{
		Name:        "Test",
		Description: "A test token",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.token)
}

func (s *IngestTokensTestSuite) TearDownSuite() {
	// Teardown routines use their own context to avoid not being run at all
	// when the suite gets cancelled or times out.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := s.client.Tokens.Ingest.Delete(ctx, s.token.ID)
	s.NoError(err)

	s.IntegrationTestSuite.TearDownSuite()
}

func (s *IngestTokensTestSuite) Test() {
	// Let's update the token.
	token, err := s.client.Tokens.Ingest.Update(s.suiteCtx, s.token.ID, axiom.TokenCreateUpdateRequest{
		Name:        "Test",
		Description: "A very good test token",
		Scopes:      []string{"*"},
	})
	s.Require().NoError(err)
	s.Require().NotNil(token)

	s.token = token

	// Get the token and make sure it matches what we have updated it to.
	token, err = s.client.Tokens.Ingest.Get(s.ctx, s.token.ID)
	s.Require().NoError(err)
	s.Require().NotNil(token)

	s.Equal(s.token, token)

	// Let's get the raw token string and make sure it has the same scopes as
	// the token entity.
	rawToken, err := s.client.Tokens.Ingest.View(s.ctx, s.token.ID)
	s.Require().NoError(err)
	s.Require().NotNil(rawToken)

	s.NotEmpty(rawToken.Token)
	s.Equal(s.token.Scopes, rawToken.Scopes)

	// List all tokens and make sure the created token is part of that list.
	tokens, err := s.client.Tokens.Ingest.List(s.ctx)
	s.Require().NoError(err)
	s.Require().NotNil(tokens)

	s.Contains(tokens, s.token)

	// Create a separate client that uses the ingest token as authentication
	// token and test the Validate() method.
	oldClient, oldAccessToken := s.client, accessToken
	accessToken = rawToken.Token
	s.newClient()
	defer func() {
		s.client, accessToken = oldClient, oldAccessToken

		if strictDecoding {
			optsErr := s.client.Options(axiom.SetStrictDecoding())
			s.Require().NoError(optsErr)
		}
	}()

	err = s.client.Tokens.Ingest.Validate(s.ctx)
	s.Require().NoError(err)
}

// PersonalTokensTestSuite tests all methods of the Axiom Personal Tokens API
// against a live deployment.
type PersonalTokensTestSuite struct {
	IntegrationTestSuite

	token *axiom.Token
}

func TestPersonalTokensTestSuite(t *testing.T) {
	suite.Run(t, &PersonalTokensTestSuite{})
}

func (s *PersonalTokensTestSuite) SetupSuite() {
	s.IntegrationTestSuite.SetupSuite()

	var err error
	s.token, err = s.client.Tokens.Personal.Create(s.suiteCtx, axiom.TokenCreateUpdateRequest{
		Name:        "Test",
		Description: "A test token",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.token)
}

func (s *PersonalTokensTestSuite) TearDownSuite() {
	// Teardown routines use their own context to avoid not being run at all
	// when the suite gets cancelled or times out.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := s.client.Tokens.Personal.Delete(ctx, s.token.ID)
	s.NoError(err)

	s.IntegrationTestSuite.TearDownSuite()
}

func (s *PersonalTokensTestSuite) Test() {
	// Let's update the token.
	token, err := s.client.Tokens.Personal.Update(s.suiteCtx, s.token.ID, axiom.TokenCreateUpdateRequest{
		Name:        "Test",
		Description: "A very good test token",
		Scopes:      []string{"*"},
	})
	s.Require().NoError(err)
	s.Require().NotNil(token)

	s.token = token

	// Get the token and make sure it matches what we have updated it to.
	token, err = s.client.Tokens.Personal.Get(s.ctx, s.token.ID)
	s.Require().NoError(err)
	s.Require().NotNil(token)

	s.Equal(s.token, token)

	// Let's get the raw token string and make sure it has the same scopes as
	// the token entity.
	rawToken, err := s.client.Tokens.Personal.View(s.ctx, s.token.ID)
	s.Require().NoError(err)
	s.Require().NotNil(rawToken)

	s.NotEmpty(rawToken.Token)
	s.Equal(s.token.Scopes, rawToken.Scopes)

	// List all tokens and make sure the created token is part of that list.
	tokens, err := s.client.Tokens.Personal.List(s.ctx)
	s.Require().NoError(err)
	s.Require().NotNil(tokens)

	s.Contains(tokens, s.token)
}
