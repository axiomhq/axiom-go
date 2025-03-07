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

	"github.com/axiomhq/axiom-go/internal/test/testhelper"
)

func TestOrganizationsService_List(t *testing.T) {
	exp := []*Organization{
		{
			ID:            "axiom",
			Name:          "Axiom Industries Ltd",
			Slug:          "",
			Plan:          "enterprise",
			PlanCreated:   testhelper.MustTimeParse(t, time.RFC3339, "1970-01-01T00:00:00Z"),
			LastUsageSync: testhelper.MustTimeParse(t, time.RFC3339, "0001-01-01T00:00:00Z"),
			Role:          RoleAdmin,
			PrimaryEmail:  "herb@axiom.sh",
			License: License{
				ID:                  "98baf1f7-0b51-403f-abc1-2ee91972a225",
				Issuer:              "console.dev.axiomtestlabs.co",
				IssuedTo:            "testorg-9t84.LAMdQbdnHiGOYCKLp0",
				IssuedAt:            testhelper.MustTimeParse(t, time.RFC3339, "2021-01-19T17:55:53Z"),
				ValidFrom:           testhelper.MustTimeParse(t, time.RFC3339, "2021-01-19T17:55:53Z"),
				ExpiresAt:           testhelper.MustTimeParse(t, time.RFC3339, "2022-01-19T17:55:53Z"),
				Plan:                "enterprise",
				MonthlyIngestGB:     100,
				MonthlyQueryGbHours: 100,
				MaxUsers:            50,
				MaxDatasets:         25,
				MaxMonitors:         10,
				MaxFields:           25,
				MaxEndpoints:        10,
				MaxQueryWindow:      time.Hour * 24 * 30,
				MaxAuditWindow:      time.Hour * 24 * 30,
				WithRBAC:            true,
				WithAuths: []string{
					"local",
					"sso",
				},
				Error: "",
			},
			PaymentStatus: Success,
			CreatedAt:     testhelper.MustTimeParse(t, time.RFC3339, "1970-01-01T00:00:00Z"),
			ModifiedAt:    testhelper.MustTimeParse(t, time.RFC3339, "2021-03-11T13:27:28.501218883Z"),
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
				"plan": "enterprise",
				"orgCreated": "1970-01-01T00:00:00Z",
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
					"monthlyIngestGb": 100,
					"monthlyIngestGbHours": 100,
					"maxUsers": 50,
					"maxDatasets": 25,
					"maxMonitors": 10,
					"maxFields": 25,
					"maxEndpoints": 10,
					"maxQueryWindowSeconds": 2592000,
					"maxAuditWindowSeconds": 2592000,
					"withRBAC": true,
					"withAuths": [
						"local",
						"sso"
					],
					"error": ""
				},
				"paymentStatus": "success",
				"metaCreated": "1970-01-01T00:00:00Z",
				"metaModified": "2021-03-11T13:27:28.501218883Z",
				"metaVersion": "1615469248501218883"
			}
		]`)
		assert.NoError(t, err)
	}

	client := setup(t, "GET /v2/orgs", hf)

	res, err := client.Organizations.List(context.Background())
	require.NoError(t, err)

	assert.EqualValues(t, exp, res)
}

func TestOrganizationsService_Get(t *testing.T) {
	exp := &Organization{
		ID:            "axiom",
		Name:          "Axiom Industries Ltd",
		Slug:          "",
		Plan:          "enterprise",
		PlanCreated:   testhelper.MustTimeParse(t, time.RFC3339, "1970-01-01T00:00:00Z"),
		LastUsageSync: testhelper.MustTimeParse(t, time.RFC3339, "0001-01-01T00:00:00Z"),
		Role:          RoleAdmin,
		PrimaryEmail:  "herb@axiom.sh",
		License: License{
			ID:                  "98baf1f7-0b51-403f-abc1-2ee91972a225",
			Issuer:              "console.dev.axiomtestlabs.co",
			IssuedTo:            "testorg-9t84.LAMdQbdnHiGOYCKLp0",
			IssuedAt:            testhelper.MustTimeParse(t, time.RFC3339, "2021-01-19T17:55:53Z"),
			ValidFrom:           testhelper.MustTimeParse(t, time.RFC3339, "2021-01-19T17:55:53Z"),
			ExpiresAt:           testhelper.MustTimeParse(t, time.RFC3339, "2022-01-19T17:55:53Z"),
			Plan:                "enterprise",
			MonthlyIngestGB:     100,
			MonthlyQueryGbHours: 100,
			MaxUsers:            50,
			MaxDatasets:         25,
			MaxMonitors:         10,
			MaxFields:           25,
			MaxEndpoints:        10,
			MaxQueryWindow:      time.Hour * 24 * 30,
			MaxAuditWindow:      time.Hour * 24 * 30,
			WithRBAC:            true,
			WithAuths: []string{
				"local",
				"sso",
			},
			Error: "",
		},
		PaymentStatus: Success,
		CreatedAt:     testhelper.MustTimeParse(t, time.RFC3339, "1970-01-01T00:00:00Z"),
		ModifiedAt:    testhelper.MustTimeParse(t, time.RFC3339, "2021-03-11T13:27:28.501218883Z"),
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
			"id": "axiom",
			"name": "Axiom Industries Ltd",
			"slug": "",
			"plan": "enterprise",
			"orgCreated": "1970-01-01T00:00:00Z",
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
				"monthlyIngestGb": 100,
				"monthlyIngestGbHours": 100,
				"maxUsers": 50,
				"maxDatasets": 25,
				"maxMonitors": 10,
				"maxFields": 25,
				"maxEndpoints": 10,
				"maxQueryWindowSeconds": 2592000,
				"maxAuditWindowSeconds": 2592000,
				"withRBAC": true,
				"withAuths": [
					"local",
					"sso"
				],
				"error": ""
			},
			"paymentStatus": "success",
			"metaCreated": "1970-01-01T00:00:00Z",
			"metaModified": "2021-03-11T13:27:28.501218883Z",
			"metaVersion": "1615469248501218883"
		}`)
		assert.NoError(t, err)
	}

	client := setup(t, "GET /v2/orgs/axiom", hf)

	res, err := client.Organizations.Get(context.Background(), "axiom")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestPaymentStatus_Marshal(t *testing.T) {
	exp := `{
		"paymentStatus": "success"
	}`

	b, err := json.Marshal(struct {
		PaymentStatus PaymentStatus `json:"paymentStatus"`
	}{
		PaymentStatus: Success,
	})
	require.NoError(t, err)
	require.NotEmpty(t, b)

	assert.JSONEq(t, exp, string(b))
}

func TestPaymentStatus_Unmarshal(t *testing.T) {
	var act struct {
		PaymentStatus PaymentStatus `json:"paymentStatus"`
	}
	err := json.Unmarshal([]byte(`{ "paymentStatus": "success" }`), &act)
	require.NoError(t, err)

	assert.Equal(t, Success, act.PaymentStatus)
}

func TestPaymentStatus_String(t *testing.T) {
	// Check outer bounds.
	assert.Empty(t, PaymentStatus(0).String())
	assert.Empty(t, emptyPaymentStatus.String())
	assert.Equal(t, emptyPaymentStatus, PaymentStatus(0))
	assert.Contains(t, (Blocked + 1).String(), "PaymentStatus(")

	for p := Success; p <= Blocked; p++ {
		s := p.String()
		assert.NotEmpty(t, s)
		assert.NotContains(t, s, "PaymentStatus(")
	}
}

func TestPaymentStatusFromString(t *testing.T) {
	for p := Success; p <= Blocked; p++ {
		parsed, err := paymentStatusFromString(p.String())
		assert.NoError(t, err)
		assert.Equal(t, p, parsed)
	}
}

func TestLicense(t *testing.T) {
	exp := License{
		ID:             "98baf1f7-0b51-403f-abc1-2ee91972a225",
		Plan:           "personal",
		MaxUsers:       50,
		MaxDatasets:    25,
		MaxMonitors:    10,
		MaxFields:      25,
		MaxEndpoints:   10,
		MaxQueryWindow: time.Hour * 24 * 30,
		MaxAuditWindow: time.Hour * 24 * 30,
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
		"monthlyIngestGb": 0,
		"monthlyIngestGbHours": 0,
		"storageAllowanceGb": 0,
		"maxUsers": 0,
		"maxDatasets": 0,
		"maxMonitors": 0,
		"maxFields": 0,
		"maxEndpoints": 0,
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
