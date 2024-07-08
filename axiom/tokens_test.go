package axiom

import (
	"context"
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
	client := setup(t, "/v2/tokens", hf)

	res, err := client.Tokens.List(context.Background())
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
	client := setup(t, "/v2/tokens/test", hf)

	res, err := client.Tokens.Get(context.Background(), "test")
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
	client := setup(t, "/v2/tokens", hf)

	res, err := client.Tokens.Create(context.Background(), CreateTokenRequest{
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
			ExpiresAt:   tokenTime.Add(24 * time.Hour).UTC().Truncate(time.Second),
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
	client := setup(t, "/v2/tokens/test/regenerate", hf)

	res, err := client.Tokens.Regenerate(context.Background(), "test", RegenerateTokenRequest{
		ExistingTokenExpiresAt: tokenTime,
		NewTokenExpiresAt:      tokenTime.Add(24 * time.Hour),
	})
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestTokensService_Delete(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}

	client := setup(t, "/v2/tokens/testID", hf)

	err := client.Tokens.Delete(context.Background(), "testID")
	require.NoError(t, err)
}
