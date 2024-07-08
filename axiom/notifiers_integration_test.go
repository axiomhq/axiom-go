package axiom_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
)

// NotifiersTestSuite tests all methods of the Axiom Notifiers API against a
// live deployment.
type NotifiersTestSuite struct {
	IntegrationTestSuite

	// Setup once per test.
	notifier *axiom.Notifier
}

func TestNotifiersTestSuite(t *testing.T) {
	suite.Run(t, new(NotifiersTestSuite))
}

func (s *NotifiersTestSuite) SetupSuite() {
	s.IntegrationTestSuite.SetupSuite()
}

func (s *NotifiersTestSuite) TearDownSuite() {
	s.IntegrationTestSuite.TearDownSuite()
}

func (s *NotifiersTestSuite) SetupTest() {
	s.IntegrationTestSuite.SetupTest()

	var err error
	s.notifier, err = s.client.Notifiers.Create(s.ctx, axiom.Notifier{
		Name: "Test Notifier",
		Properties: axiom.NotifierProperties{
			Email: &axiom.EmailConfig{
				Emails: []string{"john@example.com"},
			},
		},
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.notifier)
}

func (s *NotifiersTestSuite) TearDownTest() {
	// Teardown routines use their own context to avoid not being run at all
	// when the suite gets cancelled or times out.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(s.ctx), time.Second*15)
	defer cancel()

	err := s.client.Notifiers.Delete(ctx, s.notifier.ID)
	s.NoError(err)

	s.IntegrationTestSuite.TearDownTest()
}

func (s *NotifiersTestSuite) Test() {
	// Let's update the notifier.
	notifier, err := s.client.Notifiers.Update(s.ctx, s.notifier.ID, axiom.Notifier{
		Name: "Updated Test Notifier",
		Properties: axiom.NotifierProperties{
			Email: &axiom.EmailConfig{
				Emails: []string{"fred@example.com"},
			},
		},
	})
	s.Require().NoError(err)
	s.Require().NotNil(notifier)

	s.notifier = notifier

	// Get the notifier and make sure it matches what we have updated it to.
	notifier, err = s.client.Notifiers.Get(s.ctx, s.notifier.ID)
	s.Require().NoError(err)
	s.Require().NotNil(notifier)

	s.Equal(s.notifier, notifier)

	// List all notifiers and make sure the created notifier is part of that
	// list.
	notifiers, err := s.client.Notifiers.List(s.ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(notifiers)

	s.Contains(notifiers, s.notifier)
}
