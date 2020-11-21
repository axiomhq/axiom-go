// +build integration

package axiom_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go"
)

const ingestData = `[
	{
		"time": "17/May/2015:08:05:32 +0000",
		"remote_ip": "93.180.71.3",
		"remote_user": "-",
		"request": "GET /downloads/product_1 HTTP/1.1",
		"response": 304,
		"bytes": 0,
		"referrer": "-",
		"agent": "Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)"
	},
	{
		"time": "17/May/2015:08:05:32 +0000",
		"remote_ip": "93.180.71.3",
		"remote_user": "-",
		"request": "GET /downloads/product_1 HTTP/1.1",
		"response": 304,
		"bytes": 0,
		"referrer": "-",
		"agent": "Debian APT-HTTP/1.3 (0.8.16~exp12ubuntu10.21)"
	}
]`

// DatasetsTestSuite tests all methods of the Axiom Datasets API against a live
// deployment.
type DatasetsTestSuite struct {
	IntegrationTestSuite

	dataset *axiom.Dataset
}

func TestDatasetsTestSuite(t *testing.T) {
	suite.Run(t, new(DatasetsTestSuite))
}

func (s *DatasetsTestSuite) SetupSuite() {
	s.IntegrationTestSuite.SetupSuite()

	var err error
	s.dataset, err = s.client.Datasets.Create(s.suiteCtx, axiom.CreateDatasetRequest{
		Name:        "test",
		Description: "This is a test dataset.",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.dataset)
}

func (s *DatasetsTestSuite) TearDownSuite() {
	err := s.client.Datasets.Delete(s.suiteCtx, s.dataset.ID)
	s.Require().NoError(err)

	s.IntegrationTestSuite.TearDownSuite()
}

func (s *DatasetsTestSuite) TestUpdate() {
	// TODO(lukasmalkmus): Activate if awkward API validation (matching ID in
	// param and body) has been fixed.
	s.T().Skip()

	var err error
	s.dataset, err = s.client.Datasets.Update(s.ctx, s.dataset.ID, axiom.UpdateDatasetRequest{
		Description: "This is a soon to be filled test dataset.",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.dataset)
}

func (s *DatasetsTestSuite) TestGet() {
	dataset, err := s.client.Datasets.Get(s.ctx, s.dataset.ID)
	s.Require().NoError(err)
	s.Require().NotNil(dataset)

	s.Equal(s.dataset, dataset)
}

func (s *DatasetsTestSuite) TestList() {
	datasets, err := s.client.Datasets.List(s.ctx)
	s.Require().NoError(err)
	s.Require().NotNil(datasets)

	s.Contains(datasets, s.dataset)
}

func (s *DatasetsTestSuite) TestIngest() {
	var (
		ingested bytes.Buffer
		r        = io.TeeReader(strings.NewReader(ingestData), &ingested)
	)
	ingestStatus, err := s.client.Datasets.Ingest(s.ctx, s.dataset.ID, r, axiom.JSON, axiom.Identity, axiom.IngestOptions{})
	s.Require().NoError(err)
	s.Require().NotNil(ingestStatus)

	s.EqualValues(ingestStatus.Ingested, 2)
	s.Zero(ingestStatus.Failed)
	s.Empty(ingestStatus.Failures)
	s.EqualValues(ingestStatus.ProcessedBytes, ingested.Len())
}

func (s *DatasetsTestSuite) TestInfoAndStats() {
	// TODO(lukasmalkmus): Enable as soon as the API response has been fixed.
	s.T().Skip()

	datasetInfo, err := s.client.Datasets.Info(s.ctx, s.dataset.ID)
	s.Require().NoError(err)
	s.Require().NotNil(datasetInfo)

	s.Equal(datasetInfo.Name, s.dataset.Name)
	s.Equal(datasetInfo.NumEvents, 2)

	datasetStats, err := s.client.Datasets.Stats(s.ctx)
	s.Require().NoError(err)
	s.Require().NotNil(datasetStats)

	s.Contains(datasetStats, datasetInfo)
}

// TODO(lukasmalkmus): Query some stuff here.
