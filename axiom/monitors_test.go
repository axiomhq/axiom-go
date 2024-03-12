package axiom

import (
	"context"
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
			"IntervalMinutes": 0,
			"name": "test",
			"NotifierIDs": null,
			"operator": "Above",
			"RangeMinutes": 1,
			"threshold": 1
		}]`)
		assert.NoError(t, err)
	}
	client := setup(t, "/v2/monitors", hf)

	res, err := client.Monitors.List(context.Background())
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
	}
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
			"alertOnNoData": false,
			"aplQuery": "['dataset'] | summarize count() by bin_auto(_time)",
			"description": "test",
			"id": "testID",
			"IntervalMinutes": 0,
			"name": "test",
			"NotifierIDs": null,
			"operator": "Above",
			"RangeMinutes": 1,
			"threshold": 1
		}`)
		assert.NoError(t, err)
	}
	client := setup(t, "/v2/monitors/testID", hf)

	res, err := client.Monitors.Get(context.Background(), "testID")
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
			"IntervalMinutes": 0,
			"name": "test",
			"NotifierIDs": null,
			"operator": "Above",
			"RangeMinutes": 1,
			"threshold": 1
		}`)
		assert.NoError(t, err)
	}
	client := setup(t, "/v2/monitors", hf)

	res, err := client.Monitors.Create(context.Background(), MonitorCreateRequest{Monitor{
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
		AlertOnNoData: false,
		APLQuery:      "['dataset'] | summarize count() by bin_auto(_time)",
		Description:   "test",
		DisabledUntil: parseTimeOrPanic("2024-05-12T18:43:37Z"),
		ID:            "testID",
		Interval:      0,
		Name:          "test",
		NotifierIDs:   nil,
		Operator:      Above,
		Range:         time.Minute,
		Threshold:     1,
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
			"IntervalMinutes": 0,
			"name": "test",
			"NotifierIDs": null,
			"operator": "Above",
			"RangeMinutes": 1,
			"threshold": 1
		}`)
		assert.NoError(t, err)
	}
	client := setup(t, "/v2/monitors/testID", hf)

	res, err := client.Monitors.Update(context.Background(), "testID", MonitorUpdateRequest{Monitor{
		AlertOnNoData: false,
		APLQuery:      "['dataset'] | summarize count() by bin_auto(_time)",
		Description:   "test",
		DisabledUntil: parseTimeOrPanic("2024-05-12T18:43:37Z"),
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

func TestMonitorsService_Delete(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}

	client := setup(t, "/v2/monitors/testID", hf)

	err := client.Monitors.Delete(context.Background(), "testID")
	require.NoError(t, err)
}
