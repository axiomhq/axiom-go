package axiom

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/axiom-go/axiom/query"
)

var (
	testKeyStr = "aba84eee-3935-4b51-8aae-2c41b8693016"
	testKey    = uuid.MustParse(testKeyStr)

	getOptions = func(t *testing.T) SharedAccessOptions {
		return SharedAccessOptions{
			OrganizationID: "axiom",
			Dataset:        "logs",
			Filter: &query.Filter{
				Op:            query.OpEqual,
				Field:         "customer",
				Value:         "vercel",
				CaseSensitive: true,
			},
			MinStartTime: mustTimeParse(t, time.RFC3339, "2022-01-01T00:00:00Z"),
			MaxEndTime:   mustTimeParse(t, time.RFC3339, "2023-01-01T00:00:00Z"),
		}
	}
)

func TestCreateSharedAccessToken(t *testing.T) {
	token, err := CreateSharedAccessToken(testKeyStr, getOptions(t))
	require.NoError(t, err)
	require.Equal(t, "Gb3v2nSrqdWNu_gIQHSDvrofyCVI56HHZUuDi8OZ5Hs=", token)

	// Now build the token payload from scratch and see if the generated token
	// is the same.

	q := make(url.Values)
	q.Add("oi", "axiom")
	q.Add("dt", "logs")
	q.Add("fl[op]", "==")
	q.Add("fl[fd]", "customer")
	q.Add("fl[vl]", "vercel")
	q.Add("fl[cs]", "true")
	q.Add("mst", "2022-01-01T00:00:00Z")
	q.Add("met", "2023-01-01T00:00:00Z")
	qs := q.Encode()

	h := hmac.New(sha256.New, testKey[:])
	_, err = h.Write([]byte(qs))
	require.NoError(t, err)
	require.Equal(t, token, base64.URLEncoding.EncodeToString(h.Sum(nil)))
}

func TestCreateSharedAccessSignature(t *testing.T) {
	signature, err := CreateSharedAccessSignature(testKeyStr, getOptions(t))
	require.NoError(t, err)
	require.Equal(t, "dt=logs&fl%5Bcs%5D=true&fl%5Bfd%5D=customer&fl%5Bop%5D=%3D%3D&fl%5Bvl%5D=vercel&met=2023-01-01T00%3A00%3A00Z&mst=2022-01-01T00%3A00%3A00Z&oi=axiom&tk=Gb3v2nSrqdWNu_gIQHSDvrofyCVI56HHZUuDi8OZ5Hs%3D", signature)

	// Now parse the signature and see if the signature params match the input.

	q, err := url.ParseQuery(signature)
	require.NoError(t, err)

	assert.Equal(t, "axiom", q.Get("oi"))
	assert.Equal(t, "logs", q.Get("dt"))
	assert.Equal(t, query.OpEqual.String(), q.Get("fl[op]"))
	assert.Equal(t, "customer", q.Get("fl[fd]"))
	assert.Equal(t, "vercel", q.Get("fl[vl]"))
	assert.Equal(t, "true", q.Get("fl[cs]"))
	assert.Equal(t, "2022-01-01T00:00:00Z", q.Get("mst"))
	assert.Equal(t, "2023-01-01T00:00:00Z", q.Get("met"))
	assert.Equal(t, "Gb3v2nSrqdWNu_gIQHSDvrofyCVI56HHZUuDi8OZ5Hs=", q.Get("tk"))
}
