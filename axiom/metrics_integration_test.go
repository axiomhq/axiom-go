package axiom_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	collmetricpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/protobuf/proto"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/mpl"
)

// MetricsTestSuite tests metrics operations against the edge endpoint of a live
// deployment.
type MetricsTestSuite struct {
	IntegrationTestSuite

	dataset *axiom.Dataset
}

func TestMetricsTestSuite(t *testing.T) {
	suite.Run(t, new(MetricsTestSuite))
}

func (s *MetricsTestSuite) SetupSuite() {
	if edgeURL == "" || edgeToken == "" {
		s.T().Skip("skipping metrics integration tests; set AXIOM_EDGE_URL and AXIOM_EDGE_TOKEN to run them")
	}

	s.IntegrationTestSuite.SetupSuite()

	var err error
	s.dataset, err = s.client.Datasets.Create(s.suiteCtx, axiom.DatasetCreateRequest{
		Name:           "test-axiom-go-metrics-" + datasetSuffix,
		Kind:           "otel:metrics:v1",
		Description:    "This is a test dataset for metrics integration tests.",
		EdgeDeployment: edgeDeployment,
	})
	s.Require().NoError(err)
}

func (s *MetricsTestSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(s.suiteCtx), time.Second*15)
	defer cancel()

	if s.dataset != nil {
		err := s.client.Datasets.Delete(ctx, s.dataset.ID)
		s.NoError(err)
	}

	s.IntegrationTestSuite.TearDownSuite()
}

func (s *MetricsTestSuite) TestQuery() {
	sampleTime := time.Now().Truncate(time.Minute)

	// Metrics are ingested as OTLP protobuf only.
	b, err := proto.Marshal(&collmetricpb.ExportMetricsServiceRequest{
		ResourceMetrics: []*metricpb.ResourceMetrics{{
			ScopeMetrics: []*metricpb.ScopeMetrics{{
				Metrics: []*metricpb.Metric{{
					Name: "axiom_go_test",
					Data: &metricpb.Metric_Gauge{Gauge: &metricpb.Gauge{
						DataPoints: []*metricpb.NumberDataPoint{{
							Attributes: []*commonpb.KeyValue{{
								Key:   "test",
								Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: "mpl"}},
							}},
							TimeUnixNano: uint64(sampleTime.UnixNano()),
							Value:        &metricpb.NumberDataPoint_AsDouble{AsDouble: 42},
						}},
					}},
				}},
			}},
		}},
	})
	s.Require().NoError(err)

	// Ingest through the edge, which requires an API token.
	ingestClient, err := newClient(axiom.SetToken(edgeToken))
	s.Require().NoError(err)

	path, err := url.JoinPath(edgeURL, "v1/metrics")
	s.Require().NoError(err)

	req, err := ingestClient.NewRequest(s.ctx, http.MethodPost, path, bytes.NewReader(b))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("X-Axiom-Dataset", s.dataset.ID)

	_, err = ingestClient.Do(req, nil)
	s.Require().NoError(err)

	// Query with the personal token of the main client.
	client, err := newClient(axiom.SetEdgeURL(edgeURL))
	s.Require().NoError(err)

	q := fmt.Sprintf("param $test: string; `%s`:`axiom_go_test` | where `test` == $test | align to 1m using last", s.dataset.ID)

	var res *mpl.Result
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		var err error
		res, err = client.QueryMPL(s.ctx, q, sampleTime.Add(-5*time.Minute), sampleTime.Add(5*time.Minute),
			mpl.SetParam("test", `"mpl"`),
		)
		if assert.NoError(c, err) {
			assert.NotEmpty(c, res.Series)
		}
	}, 30*time.Second, time.Second, "ingested metric did not become queryable")

	s.Require().Len(res.Series, 1)
	series := res.Series[0]
	s.Equal("axiom_go_test", series.Metric)
	s.Equal("mpl", series.Tags["test"])
	s.Require().Equal(time.Minute, series.Resolution)
	s.NotEmpty(res.TraceID)

	i := int(sampleTime.Sub(series.Start) / series.Resolution)
	s.Require().GreaterOrEqual(i, 0)
	s.Require().Less(i, len(series.Data))
	s.Require().NotNil(series.Data[i])
	s.EqualValues(42, *series.Data[i])
}
