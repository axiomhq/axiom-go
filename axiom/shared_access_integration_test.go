//go:build integration

package axiom_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/apl"
	"github.com/axiomhq/axiom-go/axiom/query"
	"github.com/axiomhq/axiom-go/axiom/sas"
)

var (
	now = time.Now().UTC()

	sharedAccessTestEvents = []axiom.Event{
		// Axiom, Project 1
		{axiom.TimestampField: now, "customer": "axiom", "project-id": "project-1", "favourites": 1},
		{axiom.TimestampField: now.Add(-time.Second), "customer": "axiom", "project-id": "project-1", "favourites": 2},

		// Axiom, Project 2
		{axiom.TimestampField: now, "customer": "axiom", "project-id": "project-2", "favourites": 3},
		{axiom.TimestampField: now.Add(-time.Second), "customer": "axiom", "project-id": "project-2", "favourites": 4},

		// Vercel, Project 1
		{axiom.TimestampField: now, "customer": "vercel", "project-id": "project-1", "favourites": 5},
		{axiom.TimestampField: now.Add(-time.Second), "customer": "vercel", "project-id": "project-1", "favourites": 6},

		// Vercel, Project 2
		{axiom.TimestampField: now, "customer": "vercel", "project-id": "project-2", "favourites": 7},
		{axiom.TimestampField: now.Add(-time.Second), "customer": "vercel", "project-id": "project-2", "favourites": 8},
	}
)

// SharedAccessTestSuite tests all functionality related to shared access
// against a live deployment.
type SharedAccessTestSuite struct {
	IntegrationTestSuite

	httpClient *http.Client
	signature  string

	keys    *axiom.SharedAccessKeys
	dataset *axiom.Dataset
}

func TestSharedAccessTestSuite(t *testing.T) {
	suite.Run(t, new(SharedAccessTestSuite))
}

func (s *SharedAccessTestSuite) SetupSuite() {
	s.IntegrationTestSuite.SetupSuite()

	if !s.isCloud {
		s.T().Skip("Skipping Axiom Cloud integration tests")
	}

	s.httpClient = axiom.DefaultHTTPClient()

	var err error
	s.keys, err = s.client.Organizations.Cloud.ViewSharedAccessKeys(s.suiteCtx, orgID)
	s.Require().NoError(err)
	s.Require().NotNil(s.keys)

	s.dataset, err = s.client.Datasets.Create(s.suiteCtx, axiom.DatasetCreateRequest{
		Name:        "test-axiom-go-shared-access-" + datasetSuffix,
		Description: "This is a test dataset for shared access integration tests.",
	})
	s.Require().NoError(err)
	s.Require().NotNil(s.dataset)

	_, err = s.client.Datasets.IngestEvents(s.suiteCtx, s.dataset.ID, axiom.IngestOptions{}, sharedAccessTestEvents...)
	s.Require().NoError(err)

	// Create a shared access signature valid for the "vercel" customer and
	// queries within a 10 minute time window, which should be enough for this
	// test.
	s.signature, err = sas.Create(s.keys.Primary, sas.Options{
		OrganizationID: orgID,
		Dataset:        s.dataset.ID,
		Filter: query.Filter{
			Op:    query.OpEqual,
			Field: "customer",
			Value: "vercel",
		},
		MinStartTime: now.Add(5 * -time.Minute),
		MaxEndTime:   now.Add(5 * time.Minute),
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(s.signature)
}

func (s *SharedAccessTestSuite) TearDownSuite() {
	// Teardown routines use their own context to avoid not being run at all
	// when the suite gets cancelled or times out.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := s.client.Datasets.Delete(ctx, s.dataset.ID)
	s.NoError(err)

	s.IntegrationTestSuite.TearDownSuite()
}

//nolint:bodyclose // doQueryRequest() actually closes the body.
func (s *SharedAccessTestSuite) TestQueryValid() {
	// Sum favourites for the "vercel" customer within the last minute.
	q := query.Query{
		StartTime: now.Add(-time.Minute),
		EndTime:   now,
		Aggregations: []query.Aggregation{
			{
				Alias: "fav_count",
				Op:    query.OpSum,
				Field: "favourites",
			},
		},
	}

	var (
		res  query.Result
		resp = s.doQueryRequest(q, &res)
	)
	if s.Equal(http.StatusOK, resp.StatusCode) {
		s.EqualValues(8, res.Status.RowsExamined) // 8 events in the dataset
		s.EqualValues(4, res.Status.RowsMatched)  // 4 events belonging to vercel
		if s.Len(res.Buckets.Totals, 1) {
			agg := res.Buckets.Totals[0].Aggregations[0]
			s.EqualValues("fav_count", agg.Alias)
			s.EqualValues(26, agg.Value)
		}
	}

	// Sum favourites for the "vercel" customers "project-1" project within the
	// last minute.
	q = query.Query{
		StartTime: now.Add(-time.Minute),
		EndTime:   now,
		Aggregations: []query.Aggregation{
			{
				Alias: "fav_count",
				Op:    query.OpSum,
				Field: "favourites",
			},
		},
		Filter: query.Filter{
			Op:    query.OpEqual,
			Field: "project-id",
			Value: "project-1",
		},
	}

	resp = s.doQueryRequest(q, &res)
	if s.Equal(http.StatusOK, resp.StatusCode) {
		s.EqualValues(8, res.Status.RowsExamined) // 8 events in the dataset
		s.EqualValues(2, res.Status.RowsMatched)  // 2 events belonging to vercel, project-1
		if s.Len(res.Buckets.Totals, 1) {
			agg := res.Buckets.Totals[0].Aggregations[0]
			s.EqualValues("fav_count", agg.Alias)
			s.EqualValues(11, agg.Value)
		}
	}
}

//nolint:bodyclose // doAPLRequest() actually closes the body.
func (s *SharedAccessTestSuite) TestAPLValid() {
	s.T().Skip("Skipping test until we have the apl query endpoint working")

	rawAPL := "['logs'] | where customer == 'vercel' | sum(favourites) as fav_count"

	var (
		res  apl.Result
		resp = s.doAPLRequest(rawAPL, &res)
	)
	if s.Equal(http.StatusOK, resp.StatusCode) {
		s.EqualValues(8, res.Status.RowsExamined) // 8 events in the dataset
		s.EqualValues(4, res.Status.RowsMatched)  // 4 events belonging to vercel
		if s.Len(res.Buckets.Totals, 1) {
			agg := res.Buckets.Totals[0].Aggregations[0]
			s.EqualValues("fav_count", agg.Alias)
			s.EqualValues(26, agg.Value)
		}
	}

	rawAPL = "['logs'] | where customer == 'vercel' and project-id == 'project-1' | sum(favourites) as fav_count"

	resp = s.doAPLRequest(rawAPL, &res)
	if s.Equal(http.StatusOK, resp.StatusCode) {
		s.EqualValues(8, res.Status.RowsExamined) // 8 events in the dataset
		s.EqualValues(2, res.Status.RowsMatched)  // 2 events belonging to vercel, project-1
		if s.Len(res.Buckets.Totals, 1) {
			agg := res.Buckets.Totals[0].Aggregations[0]
			s.EqualValues("fav_count", agg.Alias)
			s.EqualValues(11, agg.Value)
		}
	}
}

func (s *SharedAccessTestSuite) doQueryRequest(q query.Query, v *query.Result) *http.Response {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(q)
	s.Require().NoError(err)

	u := fmt.Sprintf("%s/api/v1/datasets/%s/query", deploymentURL, s.dataset.ID)

	return s.doRequest(u, &buf, v)
}

//nolint:unused // Test that uses this method is currently skipped.
func (s *SharedAccessTestSuite) doAPLRequest(apl string, v *apl.Result) *http.Response {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(struct {
		Raw       string    `json:"apl"`
		StartTime time.Time `json:"startTime"` // Optional
		EndTime   time.Time `json:"endTime"`   // Optional
	}{
		Raw: apl,
	})
	s.Require().NoError(err)

	u := fmt.Sprintf("%s/api/v1/datasets/_apl", deploymentURL)

	return s.doRequest(u, &buf, v)
}

func (s *SharedAccessTestSuite) doRequest(urlStr string, body io.Reader, v interface{}) *http.Response {
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, urlStr, body)
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	req.URL.RawQuery = s.signature

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err)

	err = json.NewDecoder(resp.Body).Decode(v)
	s.Require().NoError(err)

	s.NoError(resp.Body.Close())

	return resp
}
