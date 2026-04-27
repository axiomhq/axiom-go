package axiom

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonitorsService_List(t *testing.T) {
	exp := []*Monitor{
		{
			AlertOnNoData: false,
			APLQuery:      "['dataset'] | summarize count() by bin_auto(_time)",
			Description:   "test",
			DisabledUntil: parseTimeOrPanic("2024-03-12T18:43:37Z"),
			ID:            "test",
			Interval:      0,
			Name:          "test",
			NotifierIDs:   nil,
			Operator:      Above,
			Range:         time.Minute,
			Threshold:     1,
		},
	}
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `[{
			"alertOnNoData": false,
			"aplQuery": "['dataset'] | summarize count() by bin_auto(_time)",
			"description": "test",
			"disabledUntil": "2024-03-12T18:43:37Z",
			"id": "test",
			"intervalMinutes": 0,
			"name": "test",
			"notifierIDs": null,
			"operator": "Above",
			"rangeMinutes": 1,
			"threshold": 1
		}]`)
		assert.NoError(t, err)
	}
	client := setup(t, "GET /v2/monitors", hf)

	res, err := client.Monitors.List(t.Context())
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestMonitorsService_Get(t *testing.T) {
	exp := &Monitor{
		AlertOnNoData: false,
		MPLQuery:      "`dataset`:`my-metric` | align to 5s using avg",
		Description:   "test",
		DisabledUntil: time.Time{},
		ID:            "testID",
		Interval:      0,
		Name:          "test",
		NotifierIDs:   nil,
		Operator:      Above,
		Range:         time.Minute,
		Threshold:     1,
		Type:          MonitorTypeThreshold,
	}
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		mplQuery := "`dataset`:`my-metric` | align to 5s using avg"
		_, err := fmt.Fprintf(w, `{
			"alertOnNoData": false,
			"mplQuery": %q,
			"description": "test",
			"id": "testID",
			"intervalMinutes": 0,
			"name": "test",
			"notifierIDs": null,
			"operator": "Above",
			"rangeMinutes": 1,
			"threshold": 1,
			"type": "Threshold"
		}`, mplQuery)
		assert.NoError(t, err)
	}
	client := setup(t, "GET /v2/monitors/testID", hf)

	res, err := client.Monitors.Get(t.Context(), "testID")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestMonitor_MarshalUnmarshalJSON(t *testing.T) {
	exp := Monitor{
		APLQuery:  "['dataset'] | summarize count() by bin_auto(_time)",
		MPLQuery:  "`dataset`:`my-metric` | align to 5s using avg",
		Interval:  2 * time.Minute,
		Range:     3 * time.Minute,
		Delay:     10 * time.Second,
		Threshold: 1,
	}

	b, err := json.Marshal(exp)
	require.NoError(t, err)

	var gotPayload map[string]any
	err = json.Unmarshal(b, &gotPayload)
	require.NoError(t, err)

	assert.Equal(t, exp.APLQuery, gotPayload["aplQuery"])
	assert.Equal(t, exp.MPLQuery, gotPayload["mplQuery"])
	assert.Equal(t, float64(2), gotPayload["intervalMinutes"])
	assert.Equal(t, float64(3), gotPayload["rangeMinutes"])
	assert.Equal(t, float64(10), gotPayload["secondDelay"])

	var got Monitor
	err = json.Unmarshal(b, &got)
	require.NoError(t, err)

	assert.Equal(t, exp.APLQuery, got.APLQuery)
	assert.Equal(t, exp.MPLQuery, got.MPLQuery)
	assert.Equal(t, exp.Interval, got.Interval)
	assert.Equal(t, exp.Range, got.Range)
	assert.Equal(t, exp.Delay, got.Delay)
}

func TestValidateMonitorQueries(t *testing.T) {
	t.Run("aplQuery only", func(t *testing.T) {
		err := validateMonitorQueries(Monitor{
			APLQuery: "['dataset'] | summarize count() by bin_auto(_time)",
		})
		require.NoError(t, err)
	})

	t.Run("mplQuery only", func(t *testing.T) {
		err := validateMonitorQueries(Monitor{
			MPLQuery: "`dataset`:`my-metric` | align to 5s using avg",
		})
		require.NoError(t, err)
	})

	t.Run("neither query", func(t *testing.T) {
		err := validateMonitorQueries(Monitor{})
		require.EqualError(t, err, "one of aplQuery or mplQuery is required")
	})

	t.Run("both queries", func(t *testing.T) {
		err := validateMonitorQueries(Monitor{
			APLQuery: "['dataset'] | summarize count() by bin_auto(_time)",
			MPLQuery: "`dataset`:`my-metric` | align to 5s using avg",
		})
		require.EqualError(t, err, "aplQuery and mplQuery are mutually exclusive, provide only one")
	})
}

func TestMonitorsService_Create(t *testing.T) {
	exp := &Monitor{
		AlertOnNoData: false,
		APLQuery:      "['dataset'] | summarize count() by bin_auto(_time)",
		Description:   "test",
		DisabledUntil: time.Time{},
		ID:            "testID",
		Interval:      0,
		Name:          "test",
		NotifierIDs:   nil,
		Operator:      Above,
		Range:         time.Minute,
		Threshold:     1,
	}
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, mediaTypeJSON, r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
			"alertOnNoData": false,
			"aplQuery": "['dataset'] | summarize count() by bin_auto(_time)",
			"description": "test",
			"id": "testID",
			"intervalMinutes": 0,
			"name": "test",
			"notifierIDs": null,
			"operator": "Above",
			"rangeMinutes": 1,
			"threshold": 1
		}`)
		assert.NoError(t, err)
	}
	client := setup(t, "POST /v2/monitors", hf)

	res, err := client.Monitors.Create(t.Context(), MonitorCreateRequest{Monitor{
		AlertOnNoData: false,
		APLQuery:      "['dataset'] | summarize count() by bin_auto(_time)",
		Description:   "test",
		DisabledUntil: time.Time{},
		ID:            "testID",
		Interval:      0,
		Name:          "test",
		NotifierIDs:   nil,
		Operator:      Above,
		Range:         time.Minute,
		Threshold:     1,
	}})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestMonitorsService_Update(t *testing.T) {
	exp := &Monitor{
		AlertOnNoData:                false,
		APLQuery:                     "['dataset'] | summarize count() by bin_auto(_time)",
		Description:                  "test",
		DisabledUntil:                parseTimeOrPanic("2024-05-12T18:43:37Z"),
		ID:                           "testID",
		Interval:                     0,
		Name:                         "newTest",
		NotifierIDs:                  nil,
		Operator:                     Above,
		Range:                        time.Minute,
		Threshold:                    1,
		Delay:                        time.Second * 10,
		NotifyEveryRun:               true,
		SkipResolved:                 false,
		TriggerFromNRuns:             5,
		TriggerAfterNPositiveResults: 3,
		Type:                         MonitorTypeThreshold,
		Tolerance:                    4,
		CompareDays:                  8,
	}
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, mediaTypeJSON, r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
			"alertOnNoData": false,
			"aplQuery": "['dataset'] | summarize count() by bin_auto(_time)",
			"description": "test",
			"disabledUntil": "2024-05-12T18:43:37Z",
			"id": "testID",
			"intervalMinutes": 0,
			"name": "newTest",
			"notifierIDs": null,
			"operator": "Above",
			"rangeMinutes": 1,
			"threshold": 1,
			"secondDelay": 10,
			"notifyEveryRun": true,
			"skipResolved": false,
			"triggerFromNRuns": 5,
			"triggerAfterNPositiveResults": 3,
			"type": "Threshold",
			"tolerance": 4,
			"compareDays": 8
		}`)
		assert.NoError(t, err)
	}
	client := setup(t, "PUT /v2/monitors/testID", hf)

	res, err := client.Monitors.Update(t.Context(), "testID", MonitorUpdateRequest{Monitor{
		AlertOnNoData: false,
		APLQuery:      "['dataset'] | summarize count() by bin_auto(_time)",
		Description:   "test",
		DisabledUntil: parseTimeOrPanic("2024-05-12T18:43:37Z"),
		ID:            "testID",
		Interval:      0,
		Name:          "newTest",
		NotifierIDs:   nil,
		Operator:      Above,
		Range:         time.Minute,
		Threshold:     1,
		Delay:         time.Second * 10,
	}})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
	assert.Equal(t, time.Second*10, res.Delay)
}

func TestMonitorsService_Delete(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}

	client := setup(t, "DELETE /v2/monitors/testID", hf)

	err := client.Monitors.Delete(t.Context(), "testID")
	require.NoError(t, err)
}
