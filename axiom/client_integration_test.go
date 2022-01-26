//go:build integration
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
	strictDecoding bool
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

	// Generic properties.
	isCloud bool

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

	if orgID != "" {
		s.isCloud = true
	}
	if datasetSuffix == "" {
		datasetSuffix = "local"
	}

	s.T().Logf("strict decoding is set to \"%t\"", strictDecoding)
	s.T().Logf("cloud enabled: \"%t\"", s.isCloud)

	s.newClient()

	s.suiteCtx, s.suiteCancel = context.WithTimeout(context.Background(), time.Minute)

	var err error
	s.testUser, err = s.client.Users.Current(s.suiteCtx)
	s.Require().NoError(err)
	s.Require().NotNil(s.testUser)

	s.T().Logf("using account %q", s.testUser.Name)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.NoError(s.suiteCtx.Err())
	s.suiteCancel()
}

func (s *IntegrationTestSuite) SetupTest() {
	s.ctx, s.cancel = context.WithTimeout(s.suiteCtx, 45*time.Second)
}

func (s *IntegrationTestSuite) TearDownTest() {
	s.NoError(s.ctx.Err())
	s.cancel()
}

func (s *IntegrationTestSuite) Test() {
	err := s.client.ValidateCredentials(s.ctx)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) newClient(additionalOptions ...axiom.Option) {
	var err error
	s.client, err = newClient(additionalOptions...)
	s.Require().NoError(err)
	s.Require().NotNil(s.client)
}

func newClient(additionalOptions ...axiom.Option) (*axiom.Client, error) {
	var (
		userAgent = "axiom-go-integration-test/" + datasetSuffix
		options   = []axiom.Option{axiom.SetNoEnv(), axiom.SetUserAgent(userAgent)}
	)

	if deploymentURL != "" {
		options = append(options, axiom.SetURL(deploymentURL))
	}
	if accessToken != "" {
		options = append(options, axiom.SetAccessToken(accessToken))
	}
	if orgID != "" {
		options = append(options, axiom.SetOrgID(orgID))
	}

	options = append(options, axiom.SetStrictDecoding(strictDecoding))
	options = append(options, additionalOptions...)

	return axiom.NewClient(options...)
}
