package axiom_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/otel"
	"github.com/axiomhq/axiom-go/internal/version"
)

var (
	enabled                bool
	accessToken            string
	orgID                  string
	apiURL                 string
	datasetSuffix          string
	strictDecoding         bool
	telemetryTracesURL     string
	telemetryTracesToken   string
	telemetryTracesDataset string
)

func init() {
	flag.BoolVar(&enabled, "enabled", os.Getenv("AXIOM_INTEGRATION_TESTS") != "", "Enable integration tests by setting -enabled=true")
	flag.StringVar(&accessToken, "access-token", os.Getenv("AXIOM_TOKEN"), "Personal token of the test user")
	flag.StringVar(&orgID, "org-id", os.Getenv("AXIOM_ORG_ID"), "Organization ID of the organization the test user belongs to")
	flag.StringVar(&apiURL, "deployment-url", os.Getenv("AXIOM_URL"), "URL of the deployment to test against")
	flag.StringVar(&datasetSuffix, "dataset-suffix", os.Getenv("AXIOM_DATASET_SUFFIX"), "Dataset suffix to append to test datasets")
	flag.BoolVar(&strictDecoding, "strict-decoding", os.Getenv("AXIOM_DISABLE_STRICT_DECODING") == "", "Disable strict JSON response decoding by setting -strict-decoding=false")
	flag.StringVar(&telemetryTracesURL, "telemetry-traces-url", os.Getenv("TELEMETRY_TRACES_URL"), "URL to send traces to")
	flag.StringVar(&telemetryTracesToken, "telemetry-traces-token", os.Getenv("TELEMETRY_TRACES_TOKEN"), "Token that has access to the traces dataset")
	flag.StringVar(&telemetryTracesDataset, "telemetry-traces-dataset", os.Getenv("TELEMETRY_TRACES_DATASET"), "Dataset to send traces to")
}

// IntegrationTestSuite implements a base test suite for integration tests.
type IntegrationTestSuite struct {
	suite.Suite

	// Setup once per suite.
	client      *axiom.Client
	testUser    *axiom.User
	suiteCtx    context.Context
	suiteCancel context.CancelFunc
	flushTraces func() error

	// Setup once per test.
	ctx    context.Context
	cancel context.CancelFunc
}

func (s *IntegrationTestSuite) SetupSuite() {
	if !enabled {
		s.T().Skip(
			"skipping integration tests;",
			"set AXIOM_INTEGRATION_TESTS=true AXIOM_URL=<URL> AXIOM_TOKEN=<TOKEN> AXIOM_ORG_ID=<ORG_ID> to run this test",
		)
	}

	s.Require().NotEmpty(accessToken, "missing required environment variable AXIOM_TOKEN to run integration tests")
	s.Require().NotEmpty(orgID, "missing required environment variable AXIOM_ORG_ID to run integration tests")

	if datasetSuffix == "" {
		datasetSuffix = "local"
	} else {
		s.T().Logf("using dataset suffix %q", datasetSuffix)
	}

	s.T().Logf("strict decoding is set to \"%t\"", strictDecoding)

	s.suiteCtx, s.suiteCancel = context.WithTimeout(s.T().Context(), time.Minute)

	if len(telemetryTracesURL+telemetryTracesToken+telemetryTracesDataset) > 0 {
		var err error
		s.flushTraces, err = otel.InitTracing(
			s.suiteCtx,
			telemetryTracesDataset,
			fmt.Sprintf("axiom-go-integration-test-%s", datasetSuffix),
			version.Get(),
			otel.SetNoEnv(),
			otel.SetURL(telemetryTracesURL),
			otel.SetToken(telemetryTracesToken),
		)
		s.Require().NoError(err)
	}

	s.newClient()

	var err error
	s.testUser, err = s.client.Users.Current(s.suiteCtx)
	s.Require().NoError(err)
	s.Require().NotNil(s.testUser)

	s.T().Logf("using account %q", s.testUser.Name)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if f := s.flushTraces; f != nil {
		s.NoError(f())
	}

	s.NoError(context.Cause(s.suiteCtx))
	s.suiteCancel()
}

func (s *IntegrationTestSuite) SetupTest() {
	s.ctx, s.cancel = context.WithTimeout(s.suiteCtx, time.Minute)
}

func (s *IntegrationTestSuite) TearDownTest() {
	s.NoError(context.Cause(s.ctx))
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

	if apiURL != "" {
		options = append(options, axiom.SetURL(apiURL))
	}
	if accessToken != "" {
		options = append(options, axiom.SetToken(accessToken))
	}
	if orgID != "" {
		options = append(options, axiom.SetOrganizationID(orgID))
	}

	options = append(options, axiom.SetStrictDecoding(strictDecoding))
	options = append(options, additionalOptions...)

	return axiom.NewClient(options...)
}
