package axiom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotifiersService_List(t *testing.T) {
	exp := []*Notifier{
		{
			ID:   "aqIqAfZJVTXlaSiD6r",
			Name: "Cool Kids",
			Type: Email,
			Properties: map[string]interface{}{
				"UserIds": []interface{}{
					"e63a075e-393c-45ea-ac46-cf6917e930e3",
					"6a7fe355-1303-4071-be81-75fcf45a4c0f",
					"ab4479c4-4156-448d-a501-695e5dbf276c",
					"c6c7381b-b24d-4107-b82e-d6cd26a490a1",
				},
			},
			Created:  mustTimeParse(t, time.RFC3339, "2020-12-01T21:59:32.584410925Z"),
			Modified: mustTimeParse(t, time.RFC3339, "2020-12-01T21:59:32.584410925Z"),
			Version:  1606859972584410925,
		},
		{
			ID:   "d5I2Yv3Pg2Jx9Ne2Ay",
			Name: "Notify Me",
			Type: Email,
			Properties: map[string]interface{}{
				"UserIds": []interface{}{
					"752e2388-8f6d-467a-88cc-cfba5ec407f4",
				},
			},
			Created:  mustTimeParse(t, time.RFC3339, "2020-12-02T08:35:57.537528976Z"),
			Modified: mustTimeParse(t, time.RFC3339, "2020-12-02T08:35:57.537528976Z"),
			Version:  1606898157537528976,
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `[
			{
				"id": "aqIqAfZJVTXlaSiD6r",
				"name": "Cool Kids",
				"type": "email",
				"properties": {
					"UserIds": [
						"e63a075e-393c-45ea-ac46-cf6917e930e3",
						"6a7fe355-1303-4071-be81-75fcf45a4c0f",
						"ab4479c4-4156-448d-a501-695e5dbf276c",
						"c6c7381b-b24d-4107-b82e-d6cd26a490a1"
					]
				},
				"metaCreated": "2020-12-01T21:59:32.584410925Z",
				"metaModified": "2020-12-01T21:59:32.584410925Z",
				"metaVersion": 1606859972584410925,
				"disabledUntil": "0001-01-01T00:00:00Z"
			},
			{
				"id": "d5I2Yv3Pg2Jx9Ne2Ay",
				"name": "Notify Me",
				"type": "email",
				"properties": {
					"UserIds": [
						"752e2388-8f6d-467a-88cc-cfba5ec407f4"
					]
				},
				"metaCreated": "2020-12-02T08:35:57.537528976Z",
				"metaModified": "2020-12-02T08:35:57.537528976Z",
				"metaVersion": 1606898157537528976,
				"disabledUntil": "0001-01-01T00:00:00Z"
			}
		]`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/notifiers", hf)
	defer teardown()

	res, err := client.Notifiers.List(context.Background())
	require.NoError(t, err)

	assert.EqualValues(t, exp, res)
}

func TestNotifiersService_Get(t *testing.T) {
	exp := &Notifier{
		ID:   "aqIqAfZJVTXlaSiD6r",
		Name: "Cool Kids",
		Type: Email,
		Properties: map[string]interface{}{
			"UserIds": []interface{}{
				"e63a075e-393c-45ea-ac46-cf6917e930e3",
				"6a7fe355-1303-4071-be81-75fcf45a4c0f",
				"ab4479c4-4156-448d-a501-695e5dbf276c",
				"c6c7381b-b24d-4107-b82e-d6cd26a490a1",
			},
		},
		Created:  mustTimeParse(t, time.RFC3339, "2020-12-01T21:59:32.584410925Z"),
		Modified: mustTimeParse(t, time.RFC3339, "2020-12-01T21:59:32.584410925Z"),
		Version:  1606859972584410925,
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `{
			"id": "aqIqAfZJVTXlaSiD6r",
			"name": "Cool Kids",
			"type": "email",
			"properties": {
				"UserIds": [
					"e63a075e-393c-45ea-ac46-cf6917e930e3",
					"6a7fe355-1303-4071-be81-75fcf45a4c0f",
					"ab4479c4-4156-448d-a501-695e5dbf276c",
					"c6c7381b-b24d-4107-b82e-d6cd26a490a1"
				]
			},
			"metaCreated": "2020-12-01T21:59:32.584410925Z",
			"metaModified": "2020-12-01T21:59:32.584410925Z",
			"metaVersion": 1606859972584410925,
			"disabledUntil": "0001-01-01T00:00:00Z"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/notifiers/aqIqAfZJVTXlaSiD6r", hf)
	defer teardown()

	res, err := client.Notifiers.Get(context.Background(), "aqIqAfZJVTXlaSiD6r")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestNotifiersService_Create(t *testing.T) {
	exp := &Notifier{
		ID:       "ByiW67mUsS9FqZu0K0",
		Name:     "Test",
		Type:     Pagerduty,
		Created:  mustTimeParse(t, time.RFC3339, "2020-12-03T16:42:07.326658202Z"),
		Modified: mustTimeParse(t, time.RFC3339, "2020-12-03T16:42:07.326658202Z"),
		Version:  1607013727326658202,
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		_, err := fmt.Fprint(w, `{
			"id": "ByiW67mUsS9FqZu0K0",
			"name": "Test",
			"type": "pagerduty",
			"properties": null,
			"metaCreated": "2020-12-03T16:42:07.326658202Z",
			"metaModified": "2020-12-03T16:42:07.326658202Z",
			"metaVersion": 1607013727326658202,
			"disabledUntil": "0001-01-01T00:00:00Z"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/notifiers", hf)
	defer teardown()

	res, err := client.Notifiers.Create(context.Background(), Notifier{
		Name: "Test",
		Type: Pagerduty,
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestNotifiersService_Update(t *testing.T) {
	exp := &Notifier{
		ID:       "ByiW67mUsS9FqZu0K0",
		Name:     "Test",
		Type:     Webhook,
		Created:  mustTimeParse(t, time.RFC3339, "2020-12-03T16:42:07.326658202Z"),
		Modified: mustTimeParse(t, time.RFC3339, "2020-12-03T16:42:07.326658202Z"),
		Version:  1607013727326658202,
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		_, err := fmt.Fprint(w, `{
			"id": "ByiW67mUsS9FqZu0K0",
			"name": "Test",
			"type": "webhook",
			"properties": null,
			"metaCreated": "2020-12-03T16:42:07.326658202Z",
			"metaModified": "2020-12-03T16:42:07.326658202Z",
			"metaVersion": 1607013727326658202,
			"disabledUntil": "0001-01-01T00:00:00Z"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/notifiers/ByiW67mUsS9FqZu0K0", hf)
	defer teardown()

	res, err := client.Notifiers.Update(context.Background(), "ByiW67mUsS9FqZu0K0", Notifier{
		Name: "Test",
		Type: Webhook,
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestNotifiersService_Delete(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}

	client, teardown := setup(t, "/api/v1/notifiers/ByiW67mUsS9FqZu0K0", hf)
	defer teardown()

	err := client.Notifiers.Delete(context.Background(), "ByiW67mUsS9FqZu0K0")
	require.NoError(t, err)
}

func TestType_Marshal(t *testing.T) {
	exp := `{
		"type": "email"
	}`

	b, err := json.Marshal(struct {
		Type Type `json:"type"`
	}{
		Type: Email,
	})
	require.NoError(t, err)
	require.NotEmpty(t, b)

	assert.JSONEq(t, exp, string(b))
}

func TestType_Unmarshal(t *testing.T) {
	var act struct {
		Type Type `json:"type"`
	}
	err := json.Unmarshal([]byte(`{ "type": "email" }`), &act)
	require.NoError(t, err)

	assert.Equal(t, Email, act.Type)
}

func TestType_String(t *testing.T) {
	// Check outer bounds.
	assert.Contains(t, (Pagerduty - 1).String(), "Type(")
	assert.Contains(t, (Webhook + 1).String(), "Type(")

	for typ := Pagerduty; typ <= Webhook; typ++ {
		s := typ.String()
		assert.NotEmpty(t, s)
		assert.NotContains(t, s, "Type(")
	}
}
