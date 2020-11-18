// +build integration

package axiom_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/query"
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
		Name:        "test-" + randString(),
		Description: "This is a test dataset",
	})
	s.Require().NoError(err)
	s.Require().NotNil(dataset)

	s.datasetID = dataset.ID

	s.monitor, err = s.client.Monitors.Create(s.suiteCtx, axiom.Monitor{
		Name:        "Test Monitor",
		Description: "A test monitor",
		Dataset:     dataset.ID,
		Comparison:  axiom.AboveOrEqual,
		Query: query.Query{
			StartTime: time.Now().Add(-5 * time.Minute),
		},
		Frequency: time.Minute,
		Duration:  5 * time.Minute,
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
	monitor, err := s.client.Monitors.Update(s.suiteCtx, s.monitor.ID, axiom.Monitor{
		Name:        "Updated Test Monitor",
		Description: "A very good test monitor",
		Dataset:     s.datasetID,
		Comparison:  axiom.AboveOrEqual,
		Query: query.Query{
			StartTime: time.Now().Add(-5 * time.Minute),
		},
		Frequency: time.Minute,
		Duration:  5 * time.Minute,
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
	s.Require().NotNil(monitors)

	s.Contains(monitors, s.monitor)
}
