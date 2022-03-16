//go:build integration

package axiom_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/query"
)

// APITokensTestSuite tests all methods of the Axiom API Tokens API against a
// live deployment.
type APITokensTestSuite struct {
	IntegrationTestSuite

	dataset *axiom.Dataset
	token   *axiom.Token
}

func TestAPITokensTestSuite(t *testing.T) {
	suite.Run(t, &APITokensTestSuite{})
}

func (s *APITokensTestSuite) SetupSuite() {
	s.IntegrationTestSuite.SetupSuite()

	var err error
	s.dataset, err = s.client.Datasets.Create(s.suiteCtx, axiom.DatasetCreateRequest{
		Name:        "test-axiom-go-tokens-api-" + datasetSuffix,
		Description: "This is a test dataset for API tokens integration tests.",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.dataset)

	s.token, err = s.client.Tokens.API.Create(s.suiteCtx, axiom.TokenCreateUpdateRequest{
		Name:        "Test",
		Description: "A test token",
		Scopes:      []string{"*"},
		Permissions: []axiom.Permission{axiom.CanIngest},
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.token)
}

func (s *APITokensTestSuite) TearDownSuite() {
	// Teardown routines use their own context to avoid not being run at all
	// when the suite gets cancelled or times out.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := s.client.Datasets.Delete(ctx, s.dataset.ID)
	s.NoError(err)

	err = s.client.Tokens.API.Delete(ctx, s.token.ID)
	s.NoError(err)

	s.IntegrationTestSuite.TearDownSuite()
}

func (s *APITokensTestSuite) Test() {
	// Let's update the token.
	token, err := s.client.Tokens.API.Update(s.suiteCtx, s.token.ID, axiom.TokenCreateUpdateRequest{
		Name:        "Test",
		Description: "A very good test token",
		Scopes:      []string{"hopefully-non-existing-dataset"},
		Permissions: []axiom.Permission{axiom.CanQuery},
	})
	s.Require().NoError(err)
	s.Require().NotNil(token)

	s.Contains(token.Scopes, "hopefully-non-existing-dataset")
	s.Contains(token.Permissions, axiom.CanQuery)

	s.token = token

	// Get the token and make sure it matches what we have updated it to.
	token, err = s.client.Tokens.API.Get(s.ctx, s.token.ID)
	s.Require().NoError(err)
	s.Require().NotNil(token)

	s.Equal(s.token, token)

	// Let's get the raw token string and make sure it has the same scopes and
	// permissions as the token entity.
	rawToken, err := s.client.Tokens.API.View(s.ctx, s.token.ID)
	s.Require().NoError(err)
	s.Require().NotNil(rawToken)

	s.NotEmpty(rawToken.Token)
	s.Equal(s.token.Scopes, rawToken.Scopes)
	s.Equal(s.token.Permissions, rawToken.Permissions)

	// List all tokens and make sure the created token is part of that list.
	tokens, err := s.client.Tokens.API.List(s.ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(tokens)

	s.Contains(tokens, s.token)
}

func (s *APITokensTestSuite) TestScopesAndPermissions() {
	// Get the raw token to use it for authentication.
	rawToken, err := s.client.Tokens.API.View(s.ctx, s.token.ID)
	s.Require().NoError(err)

	// Create a separate client that uses the API token as authentication token.
	client, err := newClient(axiom.SetAccessToken(rawToken.Token))
	s.Require().NoError(err)
	s.Require().NotNil(client)

	// Let's make sure we cannot ingest into the dataset we initially created.
	_, err = client.Datasets.IngestEvents(s.ctx, s.dataset.ID, axiom.IngestOptions{}, ingestEvents...)
	s.Require().Error(err)
	s.Require().ErrorIs(err, axiom.ErrUnauthorized)

	// Update the token to allow querying the test dataset only.
	token, err := s.client.Tokens.API.Update(s.suiteCtx, s.token.ID, axiom.TokenCreateUpdateRequest{
		Name:        "Test",
		Description: "A very good test token with scopes and permissions",
		Scopes:      []string{s.dataset.ID},
		Permissions: []axiom.Permission{axiom.CanQuery},
	})
	s.Require().NoError(err)

	s.token = token

	// Let's make sure we cannot ingest...
	_, err = client.Datasets.IngestEvents(s.ctx, s.dataset.ID, axiom.IngestOptions{}, ingestEvents...)
	s.Require().ErrorIs(err, axiom.ErrUnauthorized)

	// ...but after updating the token to allow ingestion into the test dataset
	// only...
	token, err = s.client.Tokens.API.Update(s.suiteCtx, s.token.ID, axiom.TokenCreateUpdateRequest{
		Name:        "Test",
		Description: "A very good test token with scopes and permissions",
		Scopes:      []string{s.dataset.ID},
		Permissions: []axiom.Permission{axiom.CanIngest},
	})
	s.Require().NoError(err)

	s.token = token

	// ... we can now ingest...
	ingestStatus, err := client.Datasets.IngestEvents(s.ctx, s.dataset.ID, axiom.IngestOptions{}, ingestEvents...)
	s.Require().NoError(err)

	s.EqualValues(ingestStatus.Ingested, 2)

	// ...but not query.
	_, err = client.Datasets.Query(s.ctx, s.dataset.ID, query.Query{
		StartTime: time.Now().UTC().Add(-time.Minute),
		EndTime:   time.Now().UTC(),
	}, query.Options{})
	s.Require().Error(err)
	s.Require().ErrorIs(err, axiom.ErrUnauthorized)

	// After updating the token to allow querying the test dataset only...
	token, err = s.client.Tokens.API.Update(s.suiteCtx, s.token.ID, axiom.TokenCreateUpdateRequest{
		Name:        "Test",
		Description: "A very good test token with scopes and permissions",
		Scopes:      []string{s.dataset.ID},
		Permissions: []axiom.Permission{axiom.CanQuery},
	})
	s.Require().NoError(err)

	s.token = token

	// ...we can query now.
	_, err = client.Datasets.Query(s.ctx, s.dataset.ID, query.Query{
		StartTime: time.Now().UTC().Add(-time.Minute),
		EndTime:   time.Now().UTC(),
	}, query.Options{})
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
	})
	s.Require().NoError(err)
	s.Require().NotNil(token)

	s.Empty(token.Scopes)
	s.Empty(token.Permissions)

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
	s.Empty(token.Scopes)
	s.Empty(token.Permissions)

	// List all tokens and make sure the created token is part of that list.
	tokens, err := s.client.Tokens.Personal.List(s.ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(tokens)

	s.Contains(tokens, s.token)
}

func (s *PersonalTokensTestSuite) TestTokenRequestedCleaned() {
	// Let's make sure we can pass scopes and permissions with the request
	// without it to fail as scopes and permissions are not allowed on personal
	// tokens. They will be cleaned up automatically before making the request.
	token, err := s.client.Tokens.Personal.Update(s.suiteCtx, s.token.ID, axiom.TokenCreateUpdateRequest{
		Name:        "Test",
		Description: "A very good test token with scopes and permissions",
		Scopes:      []string{"*"},
		Permissions: []axiom.Permission{axiom.CanIngest, axiom.CanQuery},
	})
	s.Require().NoError(err)
	s.Require().NotNil(token)

	s.Empty(token.Scopes)
	s.Empty(token.Permissions)

	s.token = token
}
