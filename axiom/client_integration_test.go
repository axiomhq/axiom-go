// +build integration

package axiom_test

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
)

var (
	accessToken    string
	orgID          string
	deploymentURL  string
	datasetSuffix  string
	strictDecoding = true
)

func init() {
	flag.StringVar(&accessToken, "access-token", os.Getenv("AXIOM_TOKEN"), "Personal Access Token of the test user")
	flag.StringVar(&orgID, "org-id", os.Getenv("AXIOM_ORG_ID"), "Organization ID of the organization the test user belongs to")
	flag.StringVar(&deploymentURL, "deployment-url", os.Getenv("AXIOM_URL"), "URL of the deployment to test against")
	flag.StringVar(&datasetSuffix, "dataset-suffix", os.Getenv("AXIOM_DATASET_SUFFIX"), "Dataset suffix to append to test datasets")
	flag.BoolVar(&strictDecoding, "strict-decoding", os.Getenv("AXIOM_STRICT_DECODING") == "", "Disable strict JSON response decoding by setting -strict-decoding=false")
}

// IntegrationTestSuite implements a base test suite for integration tests.
type IntegrationTestSuite struct {
	suite.Suite

	// Setup once per suite.
	client      *axiom.Client
	testUser    *axiom.AuthenticatedUser
	suiteCtx    context.Context
	suiteCancel context.CancelFunc

	// Setup once per test.
	ctx    context.Context
	cancel context.CancelFunc
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.Require().NotEmpty(accessToken, "integration test needs a personal access token set")
	s.Require().True(orgID != "" || deploymentURL != "", "integration test needs an organization ID or deployment url set")

	s.T().Logf("strict decoding is set to \"%t\"", strictDecoding)

	s.suiteCtx, s.suiteCancel = context.WithTimeout(context.Background(), time.Minute)

	s.newClient()

	if datasetSuffix == "" {
		datasetSuffix = "local"
	}

	if strictDecoding {
		err := s.client.Options(axiom.SetStrictDecoding())
		s.Require().NoError(err)
	}

	var err error
	s.testUser, err = s.client.Users.Current(s.suiteCtx)
	s.Require().NoError(err)
	s.Require().NotNil(s.client)

	s.T().Logf("using account %q", s.testUser.Name)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.NoError(s.suiteCtx.Err())
	s.suiteCancel()
}

func (s *IntegrationTestSuite) SetupTest() {
	s.ctx, s.cancel = context.WithTimeout(s.suiteCtx, 15*time.Second)
}

func (s *IntegrationTestSuite) TearDownTest() {
	s.NoError(s.ctx.Err())
	s.cancel()
}

func (s *IntegrationTestSuite) newClient() {
	var (
		userAgent = "axiom-go-integration-test/" + datasetSuffix
		err       error
	)
	s.client, err = axiom.NewClient(
		axiom.SetURL(deploymentURL),
		axiom.SetAccessToken(accessToken),
		axiom.SetOrgID(orgID),
		axiom.SetUserAgent(userAgent),
	)
	s.Require().NoError(err)
	s.Require().NotNil(s.client)
}
