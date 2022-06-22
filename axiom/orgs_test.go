//nolint:dupl // Fine to have a bit of duplication in a test file.
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

// HINT(lukasmalkmus): Most of the tests below just test against the "cloud"
// endpoints. However, this is fine as the implementation is just an extension
// of the "selfhost" one.

func TestCloudOrganizationsService_List(t *testing.T) {
	exp := []*Organization{
		{
			ID:                  "axiom",
			Name:                "Axiom Industries Ltd",
			Slug:                "",
			Plan:                Pro,
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
			CreatedAt:  mustTimeParse(t, time.RFC3339, "1970-01-01T00:00:00Z"),
			ModifiedAt: mustTimeParse(t, time.RFC3339, "2021-03-11T13:27:28.501218883Z"),
			Version:    "1615469248501218883",
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `[
			{
				"id": "axiom",
				"name": "Axiom Industries Ltd",
				"slug": "",
				"plan": "pro",
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

	res, err := client.Organizations.Cloud.List(context.Background())
	require.NoError(t, err)

	assert.EqualValues(t, exp, res)
}

func TestCloudOrganizationsService_Get(t *testing.T) {
	exp := &Organization{
		ID:                  "axiom",
		Name:                "Axiom Industries Ltd",
		Slug:                "",
		Plan:                Pro,
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
		CreatedAt:  mustTimeParse(t, time.RFC3339, "1970-01-01T00:00:00Z"),
		ModifiedAt: mustTimeParse(t, time.RFC3339, "2021-03-11T13:27:28.501218883Z"),
		Version:    "1615469248501218883",
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
			"id": "axiom",
			"name": "Axiom Industries Ltd",
			"slug": "",
			"plan": "pro",
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

	res, err := client.Organizations.Cloud.Get(context.Background(), "axiom")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestCloudOrganizationsService_Update(t *testing.T) {
	exp := &Organization{
		ID:                  "axiom",
		Name:                "Malk Industries Ltd",
		Slug:                "",
		Plan:                Pro,
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
		CreatedAt:  mustTimeParse(t, time.RFC3339, "1970-01-01T00:00:00Z"),
		ModifiedAt: mustTimeParse(t, time.RFC3339, "2021-03-11T13:27:28.501218883Z"),
		Version:    "1615469248501218883",
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, mediaTypeJSON, r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
			"id": "axiom",
			"name": "Malk Industries Ltd",
			"slug": "",
			"plan": "pro",
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

	res, err := client.Organizations.Cloud.Update(context.Background(), "axiom", OrganizationCreateUpdateRequest{
		Name: "Malk Industries Ltd",
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestCloudOrganizationsService_License(t *testing.T) {
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

		w.Header().Set("Content-Type", mediaTypeJSON)
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

	res, err := client.Organizations.Cloud.License(context.Background(), "axiom")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestCloudOrganizationsService_Status(t *testing.T) {
	exp := &Status{
		DailyIngestUsedGB:      0,
		DailyIngestRemainingGB: 10000,
		DatasetsUsed:           0,
		DatasetsRemaining:      1000,
		UsersUsed:              4,
		UsersRemaining:         96,
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
			"dailyIngestUsedGB": 0,
			"dailyIngestRemainingGB": 10000,
			"datasetsUsed": 0,
			"datasetsRemaining": 1000,
			"usersUsed": 4,
			"usersRemaining": 96
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/orgs/axiom/status", hf)
	defer teardown()

	res, err := client.Organizations.Cloud.Status(context.Background(), "axiom")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestCloudOrganizationsService_ViewSharedAccessKeys(t *testing.T) {
	exp := &SharedAccessKeys{
		Primary:   "0d83c05a-e77e-4c20-b35d-6e6077832e58",
		Secondary: "75bb5815-8459-4b6e-a08f-1eb8058db44e",
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
			"primary": "0d83c05a-e77e-4c20-b35d-6e6077832e58",
			"secondary": "75bb5815-8459-4b6e-a08f-1eb8058db44e"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/orgs/axiom/keys", hf)
	defer teardown()

	res, err := client.Organizations.Cloud.ViewSharedAccessKeys(context.Background(), "axiom")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestCloudOrganizationsService_RotateSharedAccessKeys(t *testing.T) {
	exp := &SharedAccessKeys{
		Primary:   "75bb5815-8459-4b6e-a08f-1eb8058db44e",
		Secondary: "0d83c05a-e77e-4c20-b35d-6e6077832e58",
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
			"primary": "75bb5815-8459-4b6e-a08f-1eb8058db44e",
			"secondary": "0d83c05a-e77e-4c20-b35d-6e6077832e58"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/orgs/axiom/rotate-keys", hf)
	defer teardown()

	res, err := client.Organizations.Cloud.RotateSharedAccessKeys(context.Background(), "axiom")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestCloudOrganizationsService_Create(t *testing.T) {
	exp := &Organization{
		ID:                  "malkovitch",
		Name:                "Malk Industries Ltd",
		Slug:                "",
		Plan:                Pro,
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
		CreatedAt:  mustTimeParse(t, time.RFC3339, "2021-03-11T13:27:28.501218883Z"),
		ModifiedAt: mustTimeParse(t, time.RFC3339, "2021-03-11T13:27:28.501218883Z"),
		Version:    "1615469248501218883",
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, mediaTypeJSON, r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
			"id": "malkovitch",
			"name": "Malk Industries Ltd",
			"slug": "",
			"plan": "pro",
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
			"metaCreated": "2021-03-11T13:27:28.501218883Z",
			"metaModified": "2021-03-11T13:27:28.501218883Z",
			"metaVersion": "1615469248501218883"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/orgs", hf)
	defer teardown()

	res, err := client.Organizations.Cloud.Create(context.Background(), OrganizationCreateUpdateRequest{
		Name: "Malk Industries Ltd",
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestCloudOrganizationsService_Delete(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}

	client, teardown := setup(t, "/api/v1/orgs/malkovitch", hf)
	defer teardown()

	err := client.Organizations.Cloud.Delete(context.Background(), "malkovitch")
	require.NoError(t, err)
}

func TestPlan_Marshal(t *testing.T) {
	exp := `{
		"plan": "free"
	}`

	b, err := json.Marshal(struct {
		Plan Plan `json:"plan"`
	}{
		Plan: Free,
	})
	require.NoError(t, err)
	require.NotEmpty(t, b)

	assert.JSONEq(t, exp, string(b))
}

func TestPlan_Unmarshal(t *testing.T) {
	var act struct {
		Plan Plan `json:"plan"`
	}
	err := json.Unmarshal([]byte(`{ "plan": "free" }`), &act)
	require.NoError(t, err)

	assert.Equal(t, Free, act.Plan)
}

func TestPlan_String(t *testing.T) {
	// Check outer bounds.
	assert.Empty(t, Plan(0).String())
	assert.Empty(t, emptyPlan.String())
	assert.Equal(t, emptyPlan, Plan(0))
	assert.Contains(t, (Comped + 1).String(), "Plan(")

	for p := Free; p <= Comped; p++ {
		s := p.String()
		assert.NotEmpty(t, s)
		assert.NotContains(t, s, "Plan(")
	}
}

func TestPlanFromString(t *testing.T) {
	for plan := Free; plan <= Comped; plan++ {
		s := plan.String()

		parsedPlan, err := planFromString(s)
		assert.NoError(t, err)

		assert.NotEmpty(t, s)
		assert.Equal(t, plan, parsedPlan)
	}
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
		"tier": "",
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
