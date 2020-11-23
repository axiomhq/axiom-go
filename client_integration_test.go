// +build integration

package axiom_test

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go"
)

var (
	accessToken   string
	deploymentURL string
)

func init() {
	flag.StringVar(&accessToken, "access-token", os.Getenv("AXM_ACCESS_TOKEN"), "Personal Access Token of the Test user")
	flag.StringVar(&deploymentURL, "deployment-url", os.Getenv("AXM_DEPLOYMENT_URL"), "URL of the deployment to test against")
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

	s.suiteCtx, s.suiteCancel = context.WithTimeout(context.Background(), time.Minute)

	var err error
	s.client, err = axiom.NewClient(deploymentURL, accessToken, axiom.SetUserAgent("axiom-test"))
	s.Require().NoError(err)
	s.Require().NotNil(s.client)

	s.testUser, err = s.client.Users.Current(s.suiteCtx)
	s.Require().NoError(err)
	s.Require().NotNil(s.client)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.suiteCancel()
}

func (s *IntegrationTestSuite) SetupTest() {
	s.ctx, s.cancel = context.WithTimeout(s.suiteCtx, 15*time.Second)
}

func (s *IntegrationTestSuite) TearDownTest() {
	s.cancel()
}
