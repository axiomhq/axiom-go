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

func TestOrganizationsService_List(t *testing.T) {
	exp := []*Organization{
		{
			ID:                  "axiom",
			Name:                "Slovak Industries Ltd",
			Slug:                "",
			Plan:                Trial,
			PlanCreated:         mustTimeParse(t, time.RFC3339, "1970-01-01T00:00:00Z"),
			PlanExpires:         mustTimeParse(t, time.RFC3339, "1970-01-01T00:00:00Z"),
			Trialed:             false,
			PreviousPlan:        Free,
			PreviousPlanCreated: mustTimeParse(t, time.RFC3339, "1970-01-01T00:00:00Z"),
			PreviousPlanExpired: mustTimeParse(t, time.RFC3339, "1970-01-01T00:00:00Z"),
			LastUsageSync:       mustTimeParse(t, time.RFC3339, "0001-01-01T00:00:00Z"),
			Role:                RoleAdmin,
			PrimaryEmail:        "herb@axiom.sh",
			License: License{
				ID:                  "98baf1f7-0b51-403f-abc1-2ee91972a225",
				Issuer:              "console.dev.axiomtestlabs.co",
				IssuedTo:            "testorg-9t84.LAMdQbdnHiGOYCKLp0",
				IssuedAt:            mustTimeParse(t, time.RFC3339, "2021-01-19T17:55:53Z"),
				ValidFrom:           mustTimeParse(t, time.RFC3339, "2021-01-19T17:55:53Z"),
				ExpiresAt:           mustTimeParse(t, time.RFC3339, "2022-01-19T17:55:53Z"),
				Tier:                Enterprise,
				DailyIngestGB:       100,
				MaxUsers:            50,
				MaxTeams:            10,
				MaxDatasets:         25,
				MaxQueriesPerSecond: 25,
				MaxQueryWindow:      time.Hour * 24 * 30,
				MaxAuditWindow:      time.Hour * 24 * 30,
				WithRBAC:            true,
				WithAuths: []string{
					"local",
					"sso",
				},
				Error: "",
			},
			Created:  mustTimeParse(t, time.RFC3339, "1970-01-01T00:00:00Z"),
			Modified: mustTimeParse(t, time.RFC3339, "2021-03-11T13:27:28.501218883Z"),
			Version:  "1615469248501218883",
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `[
			{
				"id": "axiom",
				"name": "Slovak Industries Ltd",
				"slug": "",
				"plan": "trial",
				"planCreated": "1970-01-01T00:00:00Z",
				"planExpires": "1970-01-01T00:00:00Z",
				"trialed": false,
				"previousPlan": "free",
				"previousPlanCreated": "1970-01-01T00:00:00Z",
				"previousPlanExpired": "1970-01-01T00:00:00Z",
				"lastUsageSync": "0001-01-01T00:00:00Z",
				"role": "admin",
				"primaryEmail": "herb@axiom.sh",
				"license": {
					"id": "98baf1f7-0b51-403f-abc1-2ee91972a225",
					"issuer": "console.dev.axiomtestlabs.co",
					"issuedTo": "testorg-9t84.LAMdQbdnHiGOYCKLp0",
					"issuedAt": "2021-01-19T17:55:53Z",
					"validFrom": "2021-01-19T17:55:53Z",
					"expiresAt": "2022-01-19T17:55:53Z",
					"tier": "enterprise",
					"dailyIngestGb": 100,
					"maxUsers": 50,
					"maxTeams": 10,
					"maxDatasets": 25,
					"maxQueriesPerSecond": 25,
					"maxQueryWindowSeconds": 2592000,
					"maxAuditWindowSeconds": 2592000,
					"withRBAC": true,
					"withAuths": [
						"local",
						"sso"
					],
					"error": ""
				},
				"metaCreated": "1970-01-01T00:00:00Z",
				"metaModified": "2021-03-11T13:27:28.501218883Z",
				"metaVersion": "1615469248501218883"
			}
		]`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/orgs", hf)
	defer teardown()

	res, err := client.Organizations.List(context.Background())
	require.NoError(t, err)

	assert.EqualValues(t, exp, res)
}

func TestOrganizationsService_Get(t *testing.T) {
	//nolint:dupl // Fine to have a bit of duplication in a test file.
	exp := &Organization{
		ID:                  "axiom",
		Name:                "Slovak Industries Ltd",
		Slug:                "",
		Plan:                Trial,
		PlanCreated:         mustTimeParse(t, time.RFC3339, "1970-01-01T00:00:00Z"),
		PlanExpires:         mustTimeParse(t, time.RFC3339, "1970-01-01T00:00:00Z"),
		Trialed:             false,
		PreviousPlan:        Free,
		PreviousPlanCreated: mustTimeParse(t, time.RFC3339, "1970-01-01T00:00:00Z"),
		PreviousPlanExpired: mustTimeParse(t, time.RFC3339, "1970-01-01T00:00:00Z"),
		LastUsageSync:       mustTimeParse(t, time.RFC3339, "0001-01-01T00:00:00Z"),
		Role:                RoleAdmin,
		PrimaryEmail:        "herb@axiom.sh",
		License: License{
			ID:                  "98baf1f7-0b51-403f-abc1-2ee91972a225",
			Issuer:              "console.dev.axiomtestlabs.co",
			IssuedTo:            "testorg-9t84.LAMdQbdnHiGOYCKLp0",
			IssuedAt:            mustTimeParse(t, time.RFC3339, "2021-01-19T17:55:53Z"),
			ValidFrom:           mustTimeParse(t, time.RFC3339, "2021-01-19T17:55:53Z"),
			ExpiresAt:           mustTimeParse(t, time.RFC3339, "2022-01-19T17:55:53Z"),
			Tier:                Enterprise,
			DailyIngestGB:       100,
			MaxUsers:            50,
			MaxTeams:            10,
			MaxDatasets:         25,
			MaxQueriesPerSecond: 25,
			MaxQueryWindow:      time.Hour * 24 * 30,
			MaxAuditWindow:      time.Hour * 24 * 30,
			WithRBAC:            true,
			WithAuths: []string{
				"local",
				"sso",
			},
			Error: "",
		},
		Created:  mustTimeParse(t, time.RFC3339, "1970-01-01T00:00:00Z"),
		Modified: mustTimeParse(t, time.RFC3339, "2021-03-11T13:27:28.501218883Z"),
		Version:  "1615469248501218883",
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `{
			"id": "axiom",
			"name": "Slovak Industries Ltd",
			"slug": "",
			"plan": "trial",
			"planCreated": "1970-01-01T00:00:00Z",
			"planExpires": "1970-01-01T00:00:00Z",
			"trialed": false,
			"previousPlan": "free",
			"previousPlanCreated": "1970-01-01T00:00:00Z",
			"previousPlanExpired": "1970-01-01T00:00:00Z",
			"lastUsageSync": "0001-01-01T00:00:00Z",
			"role": "admin",
			"primaryEmail": "herb@axiom.sh",
			"license": {
				"id": "98baf1f7-0b51-403f-abc1-2ee91972a225",
				"issuer": "console.dev.axiomtestlabs.co",
				"issuedTo": "testorg-9t84.LAMdQbdnHiGOYCKLp0",
				"issuedAt": "2021-01-19T17:55:53Z",
				"validFrom": "2021-01-19T17:55:53Z",
				"expiresAt": "2022-01-19T17:55:53Z",
				"tier": "enterprise",
				"dailyIngestGb": 100,
				"maxUsers": 50,
				"maxTeams": 10,
				"maxDatasets": 25,
				"maxQueriesPerSecond": 25,
				"maxQueryWindowSeconds": 2592000,
				"maxAuditWindowSeconds": 2592000,
				"withRBAC": true,
				"withAuths": [
					"local",
					"sso"
				],
				"error": ""
			},
			"metaCreated": "1970-01-01T00:00:00Z",
			"metaModified": "2021-03-11T13:27:28.501218883Z",
			"metaVersion": "1615469248501218883"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/orgs/axiom", hf)
	defer teardown()

	res, err := client.Organizations.Get(context.Background(), "axiom")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestOrganizationsService_License(t *testing.T) {
	exp := &License{
		ID:                  "98baf1f7-0b51-403f-abc1-2ee91972a225",
		Issuer:              "console.dev.axiomtestlabs.co",
		IssuedTo:            "testorg-9t84.LAMdQbdnHiGOYCKLp0",
		IssuedAt:            mustTimeParse(t, time.RFC3339, "2021-01-19T17:55:53Z"),
		ValidFrom:           mustTimeParse(t, time.RFC3339, "2021-01-19T17:55:53Z"),
		ExpiresAt:           mustTimeParse(t, time.RFC3339, "2022-01-19T17:55:53Z"),
		Tier:                Enterprise,
		DailyIngestGB:       100,
		MaxUsers:            50,
		MaxTeams:            10,
		MaxDatasets:         25,
		MaxQueriesPerSecond: 25,
		MaxQueryWindow:      time.Hour * 24 * 30,
		MaxAuditWindow:      time.Hour * 24 * 30,
		WithRBAC:            true,
		WithAuths: []string{
			"local",
			"sso",
		},
		Error: "",
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `{
			"id": "98baf1f7-0b51-403f-abc1-2ee91972a225",
			"issuer": "console.dev.axiomtestlabs.co",
			"issuedTo": "testorg-9t84.LAMdQbdnHiGOYCKLp0",
			"issuedAt": "2021-01-19T17:55:53Z",
			"validFrom": "2021-01-19T17:55:53Z",
			"expiresAt": "2022-01-19T17:55:53Z",
			"tier": "enterprise",
			"dailyIngestGb": 100,
			"maxUsers": 50,
			"maxTeams": 10,
			"maxDatasets": 25,
			"maxQueriesPerSecond": 25,
			"maxQueryWindowSeconds": 2592000,
			"maxAuditWindowSeconds": 2592000,
			"withRBAC": true,
			"withAuths": [
				"local",
				"sso"
			],
			"error": ""
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/orgs/axiom/license", hf)
	defer teardown()

	res, err := client.Organizations.License(context.Background(), "axiom")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestOrganizationsService_Update(t *testing.T) {
	//nolint:dupl // Fine to have a bit of duplication in a test file.
	exp := &Organization{
		ID:                  "axiom",
		Name:                "Axiom Industries Ltd",
		Slug:                "",
		Plan:                Trial,
		PlanCreated:         mustTimeParse(t, time.RFC3339, "1970-01-01T00:00:00Z"),
		PlanExpires:         mustTimeParse(t, time.RFC3339, "1970-01-01T00:00:00Z"),
		Trialed:             false,
		PreviousPlan:        Free,
		PreviousPlanCreated: mustTimeParse(t, time.RFC3339, "1970-01-01T00:00:00Z"),
		PreviousPlanExpired: mustTimeParse(t, time.RFC3339, "1970-01-01T00:00:00Z"),
		LastUsageSync:       mustTimeParse(t, time.RFC3339, "0001-01-01T00:00:00Z"),
		Role:                RoleAdmin,
		PrimaryEmail:        "herb@axiom.sh",
		License: License{
			ID:                  "98baf1f7-0b51-403f-abc1-2ee91972a225",
			Issuer:              "console.dev.axiomtestlabs.co",
			IssuedTo:            "testorg-9t84.LAMdQbdnHiGOYCKLp0",
			IssuedAt:            mustTimeParse(t, time.RFC3339, "2021-01-19T17:55:53Z"),
			ValidFrom:           mustTimeParse(t, time.RFC3339, "2021-01-19T17:55:53Z"),
			ExpiresAt:           mustTimeParse(t, time.RFC3339, "2022-01-19T17:55:53Z"),
			Tier:                Enterprise,
			DailyIngestGB:       100,
			MaxUsers:            50,
			MaxTeams:            10,
			MaxDatasets:         25,
			MaxQueriesPerSecond: 25,
			MaxQueryWindow:      time.Hour * 24 * 30,
			MaxAuditWindow:      time.Hour * 24 * 30,
			WithRBAC:            true,
			WithAuths: []string{
				"local",
				"sso",
			},
			Error: "",
		},
		Created:  mustTimeParse(t, time.RFC3339, "1970-01-01T00:00:00Z"),
		Modified: mustTimeParse(t, time.RFC3339, "2021-03-11T13:27:28.501218883Z"),
		Version:  "1615469248501218883",
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		_, err := fmt.Fprint(w, `{
			"id": "axiom",
			"name": "Axiom Industries Ltd",
			"slug": "",
			"plan": "trial",
			"planCreated": "1970-01-01T00:00:00Z",
			"planExpires": "1970-01-01T00:00:00Z",
			"trialed": false,
			"previousPlan": "free",
			"previousPlanCreated": "1970-01-01T00:00:00Z",
			"previousPlanExpired": "1970-01-01T00:00:00Z",
			"lastUsageSync": "0001-01-01T00:00:00Z",
			"role": "admin",
			"primaryEmail": "herb@axiom.sh",
			"license": {
				"id": "98baf1f7-0b51-403f-abc1-2ee91972a225",
				"issuer": "console.dev.axiomtestlabs.co",
				"issuedTo": "testorg-9t84.LAMdQbdnHiGOYCKLp0",
				"issuedAt": "2021-01-19T17:55:53Z",
				"validFrom": "2021-01-19T17:55:53Z",
				"expiresAt": "2022-01-19T17:55:53Z",
				"tier": "enterprise",
				"dailyIngestGb": 100,
				"maxUsers": 50,
				"maxTeams": 10,
				"maxDatasets": 25,
				"maxQueriesPerSecond": 25,
				"maxQueryWindowSeconds": 2592000,
				"maxAuditWindowSeconds": 2592000,
				"withRBAC": true,
				"withAuths": [
					"local",
					"sso"
				],
				"error": ""
			},
			"metaCreated": "1970-01-01T00:00:00Z",
			"metaModified": "2021-03-11T13:27:28.501218883Z",
			"metaVersion": "1615469248501218883"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/orgs/4miTfZKp29VByAQgTd", hf)
	defer teardown()

	res, err := client.Organizations.Update(context.Background(), "4miTfZKp29VByAQgTd", OrganizationUpdateRequest{
		Name: "Axiom Industries Ltd",
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestLicense(t *testing.T) {
	exp := License{
		ID:                  "98baf1f7-0b51-403f-abc1-2ee91972a225",
		Tier:                Free,
		MaxUsers:            50,
		MaxTeams:            10,
		MaxDatasets:         25,
		MaxQueriesPerSecond: 25,
		MaxQueryWindow:      time.Hour * 24 * 30,
		MaxAuditWindow:      time.Hour * 24 * 30,
	}

	b, err := json.Marshal(exp)
	require.NoError(t, err)
	require.NotEmpty(t, b)

	var act License
	err = json.Unmarshal(b, &act)
	require.NoError(t, err)

	assert.Equal(t, exp, act)
}

func TestLicense_MarshalJSON(t *testing.T) {
	exp := `{
		"id": "",
		"issuer": "",
		"issuedTo": "",
		"issuedAt": "0001-01-01T00:00:00Z",
		"validFrom": "0001-01-01T00:00:00Z",
		"expiresAt": "0001-01-01T00:00:00Z",
		"tier": "Plan(0)",
		"dailyIngestGb": 0,
		"maxUsers": 0,
		"maxTeams": 0,
		"maxDatasets": 0,
		"maxQueriesPerSecond": 0,
		"maxQueryWindowSeconds": 3600,
		"maxAuditWindowSeconds": 3600,
		"withRBAC": false,
		"withAuths": null,
		"error": ""
	}`

	act, err := License{
		MaxQueryWindow: time.Hour,
		MaxAuditWindow: time.Hour,
	}.MarshalJSON()
	require.NoError(t, err)
	require.NotEmpty(t, act)

	assert.JSONEq(t, exp, string(act))
}

func TestLicense_UnmarshalJSON(t *testing.T) {
	exp := License{
		MaxQueryWindow: time.Hour,
		MaxAuditWindow: time.Hour,
	}

	var act License
	err := act.UnmarshalJSON([]byte(`{
		"maxQueryWindowSeconds": 3600,
		"maxAuditWindowSeconds": 3600
	}`))
	require.NoError(t, err)

	assert.Equal(t, exp, act)
}
