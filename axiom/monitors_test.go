package axiom

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestMonitorsService_List(t *testing.T) {
	exp := []*Monitor{
		{
			AlertOnNoData:   false,
			AplQuery:        "['dataset'] | summarize count() by bin_auto(_time)",
			Description:     "test",
			Disabled:        false,
			ID:              "test",
			IntervalMinutes: 0,
			MatchEveryN:     0,
			MatchValue:      "",
			Name:            "test",
			NotifierIds:     nil,
			Operator:        "Above",
			RangeMinutes:    1,
			Threshold:       1,
		},
	}
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `[{
			"alertOnNoData": false,
			"aplQuery": "['dataset'] | summarize count() by bin_auto(_time)",
			"description": "test",
			"disabled": false,
			"id": "test",
			"intervalMinutes": 0,
			"name": "test",
			"notifierIds": null,
			"operator": "Above",
			"rangeMinutes": 1,
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
		AlertOnNoData:   false,
		AplQuery:        "['dataset'] | summarize count() by bin_auto(_time)",
		Description:     "test",
		Disabled:        false,
		ID:              "testID",
		IntervalMinutes: 0,
		MatchEveryN:     0,
		MatchValue:      "",
		Name:            "test",
		NotifierIds:     nil,
		Operator:        "Above",
		RangeMinutes:    1,
		Threshold:       1,
	}
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
			"alertOnNoData": false,
			"aplQuery": "['dataset'] | summarize count() by bin_auto(_time)",
			"description": "test",
			"disabled": false,
			"id": "testID",
			"intervalMinutes": 0,
			"name": "test",
			"notifierIds": null,
			"operator": "Above",
			"rangeMinutes": 1,
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
		AlertOnNoData:   false,
		AplQuery:        "['dataset'] | summarize count() by bin_auto(_time)",
		Description:     "test",
		Disabled:        false,
		ID:              "testID",
		IntervalMinutes: 0,
		MatchEveryN:     0,
		MatchValue:      "",
		Name:            "test",
		NotifierIds:     nil,
		Operator:        "Above",
		RangeMinutes:    1,
		Threshold:       1,
	}
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, mediaTypeJSON, r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
			"alertOnNoData": false,
			"aplQuery": "['dataset'] | summarize count() by bin_auto(_time)",
			"description": "test",
			"disabled": false,
			"id": "testID",
			"intervalMinutes": 0,
			"name": "test",
			"notifierIds": null,
			"operator": "Above",
			"rangeMinutes": 1,
			"threshold": 1
		}`)
		assert.NoError(t, err)
	}
	client := setup(t, "/v2/monitors", hf)

	res, err := client.Monitors.Create(context.Background(), Monitor{
		AlertOnNoData:   false,
		AplQuery:        "['dataset'] | summarize count() by bin_auto(_time)",
		Description:     "test",
		Disabled:        false,
		ID:              "testID",
		IntervalMinutes: 0,
		MatchEveryN:     0,
		MatchValue:      "",
		Name:            "test",
		NotifierIds:     nil,
		Operator:        "Above",
		RangeMinutes:    1,
		Threshold:       1,
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestMonitorsService_Update(t *testing.T) {
	exp := &Monitor{
		AlertOnNoData:   false,
		AplQuery:        "['dataset'] | summarize count() by bin_auto(_time)",
		Description:     "test",
		Disabled:        false,
		ID:              "testID",
		IntervalMinutes: 0,
		MatchEveryN:     0,
		MatchValue:      "",
		Name:            "test",
		NotifierIds:     nil,
		Operator:        "Above",
		RangeMinutes:    1,
		Threshold:       1,
	}
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, mediaTypeJSON, r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
			"alertOnNoData": false,
			"aplQuery": "['dataset'] | summarize count() by bin_auto(_time)",
			"description": "test",
			"disabled": false,
			"id": "testID",
			"intervalMinutes": 0,
			"name": "test",
			"notifierIds": null,
			"operator": "Above",
			"rangeMinutes": 1,
			"threshold": 1
		}`)
		assert.NoError(t, err)
	}
	client := setup(t, "/v2/monitors/testID", hf)

	res, err := client.Monitors.Update(context.Background(), "testID", Monitor{
		AlertOnNoData:   false,
		AplQuery:        "['dataset'] | summarize count() by bin_auto(_time)",
		Description:     "test",
		Disabled:        false,
		ID:              "testID",
		IntervalMinutes: 0,
		MatchEveryN:     0,
		MatchValue:      "",
		Name:            "test",
		NotifierIds:     nil,
		Operator:        "Above",
		RangeMinutes:    1,
		Threshold:       1,
	})
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
