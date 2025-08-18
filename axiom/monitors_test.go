package axiom

import (
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
		Type:          MonitorTypeThreshold,
	}
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

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
			"threshold": 1,
			"type": "Threshold"
		}`)
		assert.NoError(t, err)
	}
	client := setup(t, "GET /v2/monitors/testID", hf)

	res, err := client.Monitors.Get(t.Context(), "testID")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
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
		SecondDelay:                  10 * time.Second,
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
		SecondDelay:   10 * time.Second,
	}})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
	assert.Equal(t, 10*time.Second, res.SecondDelay)
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
