package axiom

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/internal/test/testhelper"
)

func TestTokensService_List(t *testing.T) {
	tokenTime := testhelper.MustTimeParse(t, time.RFC3339, "2024-04-19T17:55:53Z")
	exp := []*APIToken{
		{
			ID:          "test",
			Name:        "test",
			Description: "test",
			ExpiresAt:   tokenTime.UTC().Truncate(time.Second),
			DatasetCapabilities: map[string]DatasetCapabilities{
				"dataset": {
					Ingest: []Action{ActionCreate},
					Query:  []Action{ActionRead},
				},
			},
			OrganisationCapabilities: OrganisationCapabilities{
				APITokens: []Action{ActionCreate},
			},
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `[{
        "datasetCapabilities": {
            "dataset": {
                "ingest": [
                    "create"
                ],
                "query": [
                    "read"
                ]
            }
        },
		"expiresAt": "2024-04-19T17:55:53Z",
        "description": "test",
        "id": "test",
        "name": "test",
        "orgCapabilities": {
            "apiTokens": [
                "create"
            ]
        }
    }]`)
		assert.NoError(t, err)
	}
	client := setup(t, "GET /v2/tokens", hf)

	res, err := client.Tokens.List(t.Context())
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestTokensService_Get(t *testing.T) {
	tokenTime := testhelper.MustTimeParse(t, time.RFC3339, "2024-04-19T17:55:53Z")
	exp := &APIToken{
		ID:          "test",
		Name:        "test",
		Description: "test",
		ExpiresAt:   tokenTime.UTC().Truncate(time.Second),
		DatasetCapabilities: map[string]DatasetCapabilities{
			"dataset": {
				Ingest: []Action{ActionCreate},
				Query:  []Action{ActionRead},
			},
		},
		OrganisationCapabilities: OrganisationCapabilities{
			APITokens: []Action{ActionCreate},
		},
	}

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
        "datasetCapabilities": {
            "dataset": {
                "ingest": [
                    "create"
                ],
                "query": [
                    "read"
                ]
            }
        },
		"expiresAt": "2024-04-19T17:55:53Z",
        "description": "test",
        "id": "test",
        "name": "test",
        "orgCapabilities": {
            "apiTokens": [
                "create"
            ]
        }
    }`)
		assert.NoError(t, err)
	}
	client := setup(t, "GET /v2/tokens/test", hf)

	res, err := client.Tokens.Get(t.Context(), "test")
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestTokensService_Create(t *testing.T) {
	tokenTime := testhelper.MustTimeParse(t, time.RFC3339, "2024-04-19T17:55:53Z")
	exp := &CreateTokenResponse{
		APIToken: APIToken{
			Name:        "test",
			Description: "test",
			ExpiresAt:   tokenTime.UTC().Truncate(time.Second),
			DatasetCapabilities: map[string]DatasetCapabilities{
				"dataset": {
					Ingest: []Action{ActionCreate},
					Query:  []Action{ActionRead},
				},
			},
			OrganisationCapabilities: OrganisationCapabilities{
				APITokens: []Action{ActionCreate},
			}},
		Token: "test",
	}
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, mediaTypeJSON, r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
        "datasetCapabilities": {
            "dataset": {
                "ingest": [
                    "create"
                ],
                "query": [
                    "read"
                ]
            }
        },
		"expiresAt": "2024-04-19T17:55:53Z",
        "description": "test",
        "name": "test",
        "orgCapabilities": {
            "apiTokens": [
                "create"
            ]
        },
		"token":"test"
    }`)
		assert.NoError(t, err)
	}
	client := setup(t, "POST /v2/tokens", hf)

	res, err := client.Tokens.Create(t.Context(), CreateTokenRequest{
		Name:        "test",
		Description: "test",
		ExpiresAt:   tokenTime.UTC().Truncate(time.Second),
		DatasetCapabilities: map[string]DatasetCapabilities{
			"dataset": {
				Ingest: []Action{ActionCreate},
				Query:  []Action{ActionRead},
			},
		},
		OrganisationCapabilities: OrganisationCapabilities{
			APITokens: []Action{ActionCreate},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestTokensService_Regenerate(t *testing.T) {
	tokenTime := testhelper.MustTimeParse(t, time.RFC3339, "2024-04-19T17:55:53Z")
	exp := &CreateTokenResponse{
		APIToken: APIToken{
			Name:        "test",
			Description: "test",
			ExpiresAt:   tokenTime.Add(time.Hour * 24).UTC().Truncate(time.Second),
			DatasetCapabilities: map[string]DatasetCapabilities{
				"dataset": {
					Ingest: []Action{ActionCreate},
					Query:  []Action{ActionRead},
				},
			},
			OrganisationCapabilities: OrganisationCapabilities{
				APITokens: []Action{ActionCreate},
			},
		},
		Token: "test",
	}
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, mediaTypeJSON, r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", mediaTypeJSON)
		_, err := fmt.Fprint(w, `{
        "datasetCapabilities": {
            "dataset": {
                "ingest": [
                  "create"
                ],
                "query": [
                    "read"
                ]
            }
        },
		"expiresAt": "2024-04-20T17:55:53Z",
        "description": "test",
        "name": "test",
        "orgCapabilities": {
            "apiTokens": [
              	"create"
            ]
        },
		"token":"test"
    }`)
		assert.NoError(t, err)
	}
	client := setup(t, "POST /v2/tokens/test/regenerate", hf)

	res, err := client.Tokens.Regenerate(t.Context(), "test", RegenerateTokenRequest{
		ExistingTokenExpiresAt: tokenTime,
		NewTokenExpiresAt:      tokenTime.Add(time.Hour * 24),
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestTokensService_Delete(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}

	client := setup(t, "DELETE /v2/tokens/testID", hf)

	err := client.Tokens.Delete(t.Context(), "testID")
	require.NoError(t, err)
}

func TestAction_Marshal(t *testing.T) {
	exp := `{
		"action": "update"
	}`

	b, err := json.Marshal(struct {
		Action Action `json:"action"`
	}{
		Action: ActionUpdate,
	})
	require.NoError(t, err)
	require.NotEmpty(t, b)

	assert.JSONEq(t, exp, string(b))
}

func TestAction_Unmarshal(t *testing.T) {
	var act struct {
		Action Action `json:"action"`
	}
	err := json.Unmarshal([]byte(`{ "action": "update" }`), &act)
	require.NoError(t, err)

	assert.Equal(t, ActionUpdate, act.Action)
}

func TestAction_String(t *testing.T) {
	// Check outer bounds.
	assert.Empty(t, Action(0).String())
	assert.Empty(t, emptyAction.String())
	assert.Equal(t, emptyAction, Action(0))
	assert.Contains(t, (ActionDelete + 1).String(), "Action(")

	for a := ActionCreate; a <= ActionDelete; a++ {
		s := a.String()
		assert.NotEmpty(t, s)
		assert.NotContains(t, s, "Action(")
	}
}

func TestActionFromString(t *testing.T) {
	for a := ActionCreate; a <= ActionDelete; a++ {
		parsed, err := actionFromString(a.String())
		assert.NoError(t, err)
		assert.Equal(t, a, parsed)
	}
}
