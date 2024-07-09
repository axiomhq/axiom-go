package axiom_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
)

// MonitorsTestSuite tests all methods of the Axiom Monitors API against a
// live deployment.
type MonitorsTestSuite struct {
	IntegrationTestSuite

	// Setup once per suite.
	datasetID string

	// Setup once per test.
	monitor *axiom.Monitor
}

func TestMonitorsTestSuite(t *testing.T) {
	suite.Run(t, new(MonitorsTestSuite))
}

func (s *MonitorsTestSuite) SetupSuite() {
	s.IntegrationTestSuite.SetupSuite()

	dataset, err := s.client.Datasets.Create(s.suiteCtx, axiom.DatasetCreateRequest{
		Name:        "test-axiom-go-monitors-" + datasetSuffix,
		Description: "This is a test dataset for monitors integration tests.",
	})
	s.Require().NoError(err)
	s.Require().NotNil(dataset)

	s.datasetID = dataset.ID
}

func (s *MonitorsTestSuite) TearDownSuite() {
	// Teardown routines use their own context to avoid not being run at all
	// when the suite gets cancelled or times out.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(s.suiteCtx), time.Second*15)
	defer cancel()

	err := s.client.Datasets.Delete(ctx, s.datasetID)
	s.NoError(err)

	s.IntegrationTestSuite.TearDownSuite()
}

func (s *MonitorsTestSuite) SetupTest() {
	s.IntegrationTestSuite.SetupTest()

	var err error
	s.monitor, err = s.client.Monitors.Create(s.ctx, axiom.MonitorCreateRequest{
		Monitor: axiom.Monitor{
			AlertOnNoData: false,
			APLQuery:      fmt.Sprintf("['%s'] | summarize count() by bin_auto(_time)", s.datasetID),
			Description:   "A test monitor",
			Interval:      time.Minute,
			Name:          "Test Monitor",
			Operator:      axiom.BelowOrEqual,
			Range:         time.Minute * 5,
			Threshold:     1,
		},
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.monitor)
}

func (s *MonitorsTestSuite) TearDownTest() {
	// Teardown routines use their own context to avoid not being run at all
	// when the suite gets cancelled or times out.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(s.ctx), time.Second*15)
	defer cancel()

	err := s.client.Monitors.Delete(ctx, s.monitor.ID)
	s.NoError(err)

	s.IntegrationTestSuite.TearDownTest()
}

func (s *MonitorsTestSuite) Test() {
	// Let's update the monitor.
	monitor, err := s.client.Monitors.Update(s.ctx, s.monitor.ID, axiom.MonitorUpdateRequest{
		Monitor: axiom.Monitor{
			AlertOnNoData: false,
			APLQuery:      fmt.Sprintf("['%s'] | summarize count() by bin_auto(_time)", s.datasetID),
			Description:   "A very good test monitor",
			DisabledUntil: time.Now().Add(time.Minute * 10),
			Interval:      time.Minute,
			Name:          "Test Monitor",
			Operator:      axiom.BelowOrEqual,
			Range:         time.Minute * 10,
			Threshold:     5,
		},
	})
	s.Require().NoError(err)
	s.Require().NotNil(monitor)

	s.monitor = monitor

	// Get the monitor and make sure it matches what we have updated it to.
	monitor, err = s.client.Monitors.Get(s.ctx, s.monitor.ID)
	s.Require().NoError(err)
	s.Require().NotNil(monitor)

	s.Equal(s.monitor, monitor)

	// List all monitors and make sure the created monitor is part of that list.
	monitors, err := s.client.Monitors.List(s.ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(monitors)

	s.Contains(monitors, s.monitor)
}
