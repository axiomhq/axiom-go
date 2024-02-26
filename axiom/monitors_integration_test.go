//go:build integration

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

	datasetID string

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

	s.monitor, err = s.client.Monitors.Create(s.suiteCtx, axiom.MonitorCreateRequest{
		Monitor: axiom.Monitor{
			AlertOnNoData: false,
			APLQuery:      fmt.Sprintf("['%s'] | summarize count() by bin_auto(_time)", s.datasetID),
			Description:   "A test monitor",
			Disabled:      false,
			Interval:      time.Minute,
			Name:          "Test Monitor",
			Operator:      "BelowOrEqual",
			Range:         5 * time.Minute,
			Threshold:     1,
		},
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

	err = s.client.Datasets.Delete(ctx, s.datasetID)
	s.NoError(err)

	s.IntegrationTestSuite.TearDownSuite()
}

func (s *MonitorsTestSuite) Test() {
	// Let's update the monitor.
	monitor, err := s.client.Monitors.Update(s.suiteCtx, s.monitor.ID, axiom.MonitorUpdateRequest{
		Monitor: axiom.Monitor{
			AlertOnNoData: false,
			APLQuery:      fmt.Sprintf("['%s'] | summarize count() by bin_auto(_time)", s.datasetID),
			Description:   "A very good test monitor",
			Disabled:      false,
			Interval:      time.Minute,
			Name:          "Test Monitor",
			Operator:      "BelowOrEqual",
			Range:         10 * time.Minute,
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
