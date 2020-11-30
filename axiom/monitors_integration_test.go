// +build integration

package axiom_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
)

// MonitorsTestSuite tests all methods of the Axiom Monitors API against a
// live deployment.
type MonitorsTestSuite struct {
	IntegrationTestSuite

	monitor *axiom.Monitor
}

func TestMonitorsTestSuite(t *testing.T) {
	suite.Run(t, new(MonitorsTestSuite))
}

func (s *MonitorsTestSuite) SetupSuite() {
	s.IntegrationTestSuite.SetupSuite()

	var err error
	s.monitor, err = s.client.Monitors.Create(s.suiteCtx, axiom.Monitor{
		Name:        "Test Monitor",
		Description: "A test monitor",
		Comparison:  axiom.AboveOrEqual,
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.monitor)
}

func (s *MonitorsTestSuite) TearDownSuite() {
	// Teardown routines use their own context to avoid not being run at all
	// when the suite gets cancelled or times out.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := s.client.Monitors.Delete(ctx, s.monitor.ID)
	s.NoError(err)

	s.IntegrationTestSuite.TearDownSuite()
}

func (s *MonitorsTestSuite) TestUpdate() {
	s.T().Skip("Enable as soon as the API param and body ID check has been fixed!")

	monitor, err := s.client.Monitors.Update(s.suiteCtx, s.monitor.ID, axiom.Monitor{
		Name:        "Updated Test Monitor",
		Description: "A very good test monitor",
		// TODO(lukasmalkmus): Probably add user and dataset.
	})
	s.Require().NoError(err)
	s.Require().NotNil(monitor)

	s.monitor = monitor
}

func (s *MonitorsTestSuite) TestGet() {
	monitor, err := s.client.Monitors.Get(s.ctx, s.monitor.ID)
	s.Require().NoError(err)
	s.Require().NotNil(monitor)

	s.Equal(s.monitor, monitor)
}

func (s *MonitorsTestSuite) TestList() {
	monitors, err := s.client.Monitors.List(s.ctx)
	s.Require().NoError(err)
	s.Require().NotNil(monitors)

	s.Contains(monitors, s.monitor)
}
