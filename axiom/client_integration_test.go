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
	deploymentURL  string
	strictDecoding = true
)

func init() {
	flag.StringVar(&accessToken, "access-token", os.Getenv("AXM_ACCESS_TOKEN"), "Personal Access Token of the Test user")
	flag.StringVar(&deploymentURL, "deployment-url", os.Getenv("AXM_DEPLOYMENT_URL"), "URL of the deployment to test against")
	flag.BoolVar(&strictDecoding, "strict-decoding", os.Getenv("AXM_STRICT_DECODING") == "", "Disable strict JSON response decoding by setting -strict-decoding=false")
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
	s.Require().NotEmpty(deploymentURL, "integration test needs a deployment url set")

	s.T().Logf("strict decoding is set to %t", strictDecoding)

	s.suiteCtx, s.suiteCancel = context.WithTimeout(context.Background(), time.Minute)

	s.newClient()

	if strictDecoding {
		err := s.client.Options(axiom.SetStrictDecoding())
		s.Require().NoError(err)
	}

	var err error
	s.testUser, err = s.client.Users.Current(s.suiteCtx)
	s.Require().NoError(err)
	s.Require().NotNil(s.client)
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
	var err error
	s.client, err = axiom.NewClient(deploymentURL, accessToken, axiom.SetUserAgent("axiom-test"))
	s.Require().NoError(err)
	s.Require().NotNil(s.client)
}
