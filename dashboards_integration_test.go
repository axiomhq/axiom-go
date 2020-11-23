// +build integration

package axiom_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go"
)

// DashboardsTestSuite tests all methods of the Axiom Dashboards API against a
// live deployment.
type DashboardsTestSuite struct {
	IntegrationTestSuite

	dashboard *axiom.Dashboard
}

func TestDashboardsTestSuite(t *testing.T) {
	suite.Run(t, new(DashboardsTestSuite))
}

func (s *DashboardsTestSuite) SetupSuite() {
	s.IntegrationTestSuite.SetupSuite()

	// TODO(lukasmalkmus): Add a dashboard with a chart and layout.
	var err error
	s.dashboard, err = s.client.Dashboards.Create(s.suiteCtx, axiom.Dashboard{
		Name:            "Test Dashboard",
		Description:     "This is a test dashboard.",
		Owner:           s.testUser.ID,
		Charts:          []interface{}{},
		Layout:          []interface{}{},
		RefreshTime:     15 * time.Second,
		SchemaVersion:   2,
		TimeWindowStart: "qr-now-30m",
		TimeWindowEnd:   "qr-now",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.dashboard)
}

func (s *DashboardsTestSuite) TearDownSuite() {
	// Teardown routines use their own context to avoid not being run at all
	// when the suite gets cancelled or times out.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := s.client.Dashboards.Delete(ctx, s.dashboard.ID)
	s.Require().NoError(err)

	s.IntegrationTestSuite.TearDownSuite()
}

// TODO(lukasmalkmus): Add Update test case.

func (s *DashboardsTestSuite) TestGet() {
	dashboard, err := s.client.Dashboards.Get(s.ctx, s.dashboard.ID)
	s.Require().NoError(err)
	s.Require().NotNil(dashboard)

	s.Equal(s.dashboard, dashboard)
}

func (s *DashboardsTestSuite) TestList() {
	dashboards, err := s.client.Dashboards.List(s.ctx)
	s.Require().NoError(err)
	s.Require().NotNil(dashboards)

	s.Contains(dashboards, s.dashboard)
}
