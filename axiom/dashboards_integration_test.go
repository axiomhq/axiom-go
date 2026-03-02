package axiom_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/axiom-go/axiom"
)

// DashboardsTestSuite tests the v2 dashboards API against a live deployment.
type DashboardsTestSuite struct {
	IntegrationTestSuite

	dataset      *axiom.Dataset
	monitor      *axiom.Monitor
	dashboardUID string
}

func TestDashboardsTestSuite(t *testing.T) {
	suite.Run(t, new(DashboardsTestSuite))
}

func (s *DashboardsTestSuite) SetupTest() {
	s.IntegrationTestSuite.SetupTest()
	s.dataset = nil
	s.monitor = nil
	s.dashboardUID = ""

	var err error
	s.dataset, err = s.client.Datasets.Create(s.ctx, axiom.DatasetCreateRequest{
		Name:        "test-axiom-go-dashboards-" + datasetSuffix,
		Description: "This is a test dataset for dashboards integration tests.",
	})
	s.Require().NoError(err)

	_, err = s.client.Datasets.IngestEvents(s.ctx, s.dataset.ID, []axiom.Event{{"service": "integration", "status": 200}})
	s.Require().NoError(err)
}

func (s *DashboardsTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(s.ctx), time.Second*15)
	defer cancel()

	if s.dashboardUID != "" {
		err := s.client.Dashboards.Delete(ctx, s.dashboardUID)
		if !isNotFound(err) {
			s.NoError(err)
		}
	}

	if s.monitor != nil {
		err := s.client.Monitors.Delete(ctx, s.monitor.ID)
		if !isNotFound(err) {
			s.NoError(err)
		}
	}

	if s.dataset != nil {
		err := s.client.Datasets.Delete(ctx, s.dataset.ID)
		s.NoError(err)
	}

	s.IntegrationTestSuite.TearDownTest()
}

func (s *DashboardsTestSuite) TestRawCRUD() {
	uid := fmt.Sprintf("dash-raw-crud-%d", time.Now().UnixNano())

	createPayload, err := json.Marshal(map[string]any{
		"uid": uid,
		"dashboard": map[string]any{
			"name":            "raw crud dashboard",
			"owner":           s.testUser.ID,
			"description":     "raw dashboards CRUD integration test",
			"charts":          []map[string]any{{"id": "note-1", "type": "Note", "text": "hello"}},
			"layout":          []map[string]any{{"i": "note-1", "x": 0, "y": 0, "w": 4, "h": 4}},
			"refreshTime":     60,
			"schemaVersion":   2,
			"timeWindowStart": "qr-now-1h",
			"timeWindowEnd":   "qr-now",
		},
		"overwrite": true,
		"message":   "integration create",
	})
	s.Require().NoError(err)

	created, err := s.client.Dashboards.CreateRaw(s.ctx, createPayload)
	s.Require().NoError(err)
	s.dashboardUID = uid

	var createdPayload map[string]any
	s.Require().NoError(json.Unmarshal(created, &createdPayload))
	createdDashboard, ok := createdPayload["dashboard"].(map[string]any)
	s.Require().True(ok, "response missing object dashboard: %#v", createdPayload)
	s.Equal(uid, createdDashboard["uid"])

	updatePayload, err := json.Marshal(map[string]any{
		"dashboard": map[string]any{
			"name":            "raw crud dashboard updated",
			"owner":           s.testUser.ID,
			"description":     "raw dashboards CRUD integration test",
			"charts":          []map[string]any{{"id": "note-1", "type": "Note", "text": "updated"}},
			"layout":          []map[string]any{{"i": "note-1", "x": 0, "y": 0, "w": 4, "h": 4}},
			"refreshTime":     60,
			"schemaVersion":   2,
			"timeWindowStart": "qr-now-1h",
			"timeWindowEnd":   "qr-now",
		},
		"overwrite": true,
		"message":   "integration update",
	})
	s.Require().NoError(err)

	updated, err := s.client.Dashboards.UpdateRaw(s.ctx, uid, updatePayload)
	s.Require().NoError(err)

	var updatedPayload map[string]any
	s.Require().NoError(json.Unmarshal(updated, &updatedPayload))
	updatedDashboard, ok := updatedPayload["dashboard"].(map[string]any)
	s.Require().True(ok, "response missing object dashboard: %#v", updatedPayload)
	s.Equal(uid, updatedDashboard["uid"])

	got, err := s.client.Dashboards.GetRaw(s.ctx, uid)
	s.Require().NoError(err)

	var gotPayload map[string]any
	s.Require().NoError(json.Unmarshal(got, &gotPayload))
	gotDashboard, ok := gotPayload["dashboard"].(map[string]any)
	s.Require().True(ok, "response missing object dashboard: %#v", gotPayload)
	s.Equal("raw crud dashboard updated", gotDashboard["name"])

	listed, err := s.client.Dashboards.ListRaw(s.ctx, nil)
	s.Require().NoError(err)

	var listedDashboards []map[string]any
	s.Require().NoError(json.Unmarshal(listed, &listedDashboards))

	var found bool
	for _, dashboard := range listedDashboards {
		if dashboard["uid"] == uid {
			found = true
			break
		}
	}
	s.True(found, "created dashboard %q was not returned by ListRaw", uid)
}

func (s *DashboardsTestSuite) TestAllChartTypes() {
	uid := fmt.Sprintf("dash-all-charts-%d", time.Now().UnixNano())

	monitor, err := s.client.Monitors.Create(s.ctx, axiom.MonitorCreateRequest{
		Monitor: axiom.Monitor{
			Name:        "test-dashboard-monitor",
			Description: "Monitor used by dashboards integration test",
			Type:        axiom.MonitorTypeThreshold,
			APLQuery:    fmt.Sprintf("['%s'] | summarize count()", s.dataset.ID),
			Operator:    axiom.Above,
			Threshold:   0,
			Interval:    time.Minute,
			Range:       time.Minute,
		},
	})
	s.Require().NoError(err)
	s.monitor = monitor

	baseQuery := map[string]any{"apl": fmt.Sprintf("['%s'] | summarize count()", s.dataset.ID)}

	charts := []map[string]any{
		{"id": "timeseries-1", "datasetId": s.dataset.ID, "type": "TimeSeries", "name": "Time Series", "query": baseQuery},
		{"id": "heatmap-1", "datasetId": s.dataset.ID, "type": "Heatmap", "name": "Heatmap", "query": baseQuery},
		{"id": "logstream-1", "datasetId": s.dataset.ID, "type": "LogStream", "name": "Log Stream", "query": baseQuery},
		{"id": "pie-1", "datasetId": s.dataset.ID, "type": "Pie", "name": "Pie", "query": baseQuery},
		{"id": "scatter-1", "datasetId": s.dataset.ID, "type": "Scatter", "name": "Scatter", "query": baseQuery},
		{"id": "table-1", "datasetId": s.dataset.ID, "type": "Table", "name": "Table", "query": baseQuery},
		{"id": "topk-1", "datasetId": s.dataset.ID, "type": "TopK", "name": "Top List", "query": baseQuery},
		{"id": "statistic-1", "datasetId": s.dataset.ID, "type": "Statistic", "name": "Statistic", "query": baseQuery},
		{"id": "sectionheader-1", "datasetId": s.dataset.ID, "type": "SectionHeader", "name": "Section Header", "query": baseQuery},
		{"id": "note-1", "type": "Note", "text": "Integration note"},
		{"id": "monitorlist-1", "type": "MonitorList", "name": "Monitor List", "selectedMonitors": []string{s.monitor.ID}, "columns": map[string]any{"status": true, "history": false, "dataset": false, "type": false, "notifiers": false}},
		{"id": "smartfilter-1", "type": "SmartFilter", "name": "Filter Bar", "filters": []map[string]any{{"type": "search", "id": "sf-search"}}},
		{"id": "spacer-1", "type": "Spacer", "name": "Spacer"},
		{"id": "placeholder-1", "type": "Placeholder"},
	}

	layout := make([]map[string]any, len(charts))
	for i := range charts {
		layout[i] = map[string]any{
			"i": charts[i]["id"],
			"x": (i % 3) * 4,
			"y": (i / 3) * 4,
			"w": 4,
			"h": 4,
		}
	}

	rawPayload, err := json.Marshal(map[string]any{
		"uid": uid,
		"dashboard": map[string]any{
			"name":            "all chart types",
			"owner":           s.testUser.ID,
			"description":     "integration coverage for all chart types",
			"charts":          charts,
			"layout":          layout,
			"refreshTime":     60,
			"schemaVersion":   2,
			"timeWindowStart": "qr-now-1h",
			"timeWindowEnd":   "qr-now",
		},
		"overwrite": true,
		"message":   "integration test create",
	})
	s.Require().NoError(err)

	created, err := s.client.Dashboards.CreateRaw(s.ctx, rawPayload)
	s.Require().NoError(err)
	s.dashboardUID = uid

	var createdPayload map[string]any
	s.Require().NoError(json.Unmarshal(created, &createdPayload))
	createdDashboard, ok := createdPayload["dashboard"].(map[string]any)
	s.Require().True(ok, "response missing object dashboard: %#v", createdPayload)
	s.Equal(uid, createdDashboard["uid"])

	got, err := s.client.Dashboards.GetRaw(s.ctx, uid)
	s.Require().NoError(err)

	var gotPayload map[string]any
	s.Require().NoError(json.Unmarshal(got, &gotPayload))
	gotDashboard, ok := gotPayload["dashboard"].(map[string]any)
	s.Require().True(ok, "response missing object dashboard: %#v", gotPayload)
	gotCharts, ok := gotDashboard["charts"].([]any)
	s.Require().True(ok, "response missing array dashboard.charts: %#v", gotDashboard)

	expectedTypes := map[string]bool{
		"TimeSeries":    false,
		"Heatmap":       false,
		"LogStream":     false,
		"Pie":           false,
		"Scatter":       false,
		"Table":         false,
		"TopK":          false,
		"Statistic":     false,
		"SectionHeader": false,
		"Note":          false,
		"MonitorList":   false,
		"SmartFilter":   false,
		"Spacer":        false,
		"Placeholder":   false,
	}

	for _, chartAny := range gotCharts {
		chart, ok := chartAny.(map[string]any)
		s.Require().True(ok, "chart should be an object: %#v", chartAny)
		typ, ok := chart["type"].(string)
		s.Require().True(ok, "chart missing type string: %#v", chart)
		if _, ok := expectedTypes[typ]; ok {
			expectedTypes[typ] = true
		}
	}

	for typ, present := range expectedTypes {
		s.True(present, "missing chart type %s in stored dashboard", typ)
	}
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}

	var httpErr axiom.HTTPError
	return errors.As(err, &httpErr) && httpErr.Status == 404
}
