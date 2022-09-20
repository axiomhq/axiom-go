package sas

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
	"github.com/axiomhq/axiom-go/internal/test/testhelper"
)

var (
	testKeyStr = "aba84eee-3935-4b51-8aae-2c41b8693016"
	testKey    = uuid.MustParse(testKeyStr)

	getOptions = func(t *testing.T) Options {
		return Options{
			OrganizationID: "axiom",
			Dataset:        "logs",
			Filter: query.Filter{
				Op:            query.OpEqual,
				Field:         "customer",
				Value:         "vercel",
				CaseSensitive: true,
			},
			MinStartTime: testhelper.MustTimeParse(t, time.RFC3339, "2022-01-01T00:00:00Z"),
			MaxEndTime:   testhelper.MustTimeParse(t, time.RFC3339, "2023-01-01T00:00:00Z"),
		}
	}
)

func TestCreate(t *testing.T) {
	signature, err := Create(testKeyStr, getOptions(t))
	require.NoError(t, err)
	require.NotEmpty(t, signature)

	assert.Equal(t, "dt=logs&fl=%7B%22op%22%3A%22%3D%3D%22%2C%22fd%22%3A%22customer%22%2C%22vl%22%3A%22vercel%22%2C%22cs%22%3Atrue%7D&met=2023-01-01T00%3A00%3A00Z&mst=2022-01-01T00%3A00%3A00Z&oi=axiom&tk=iHP9f6pQtaElHSpCBw3TSiLSy_7xsHPj01SelJ9qfWA%3D", signature)

	// Now parse the signature and see if the signature params match the input.

	q, err := url.ParseQuery(signature)
	require.NoError(t, err)

	assert.Equal(t, "axiom", q.Get("oi"))
	assert.Equal(t, "logs", q.Get("dt"))
	assert.Equal(t, `{"op":"==","fd":"customer","vl":"vercel","cs":true}`, q.Get("fl"))
	assert.Equal(t, "2022-01-01T00:00:00Z", q.Get("mst"))
	assert.Equal(t, "2023-01-01T00:00:00Z", q.Get("met"))
	assert.Equal(t, "iHP9f6pQtaElHSpCBw3TSiLSy_7xsHPj01SelJ9qfWA=", q.Get("tk"))
}

func TestCreateToken(t *testing.T) {
	token, err := CreateToken(testKeyStr, getOptions(t))
	require.NoError(t, err)
	require.NotEmpty(t, token)

	assert.Equal(t, "iHP9f6pQtaElHSpCBw3TSiLSy_7xsHPj01SelJ9qfWA=", token)

	// Now build the token payload from scratch and see if the generated token
	// is the same.

	q := make(url.Values)
	q.Add("oi", "axiom")
	q.Add("dt", "logs")
	q.Add("fl", `{"op":"==","fd":"customer","vl":"vercel","cs":true}`)
	q.Add("mst", "2022-01-01T00:00:00Z")
	q.Add("met", "2023-01-01T00:00:00Z")

	var (
		pl = buildSignaturePayload(q)
		h  = hmac.New(sha256.New, testKey[:])
	)
	_, err = h.Write([]byte(pl))
	require.NoError(t, err)
	assert.Equal(t, token, base64.URLEncoding.EncodeToString(h.Sum(nil)))
}

func TestVerify(t *testing.T) {
	exp := getOptions(t)
	signature := "dt=logs&fl=%7B%22op%22%3A%22%3D%3D%22%2C%22fd%22%3A%22customer%22%2C%22vl%22%3A%22vercel%22%2C%22cs%22%3Atrue%7D&met=2023-01-01T00%3A00%3A00Z&mst=2022-01-01T00%3A00%3A00Z&oi=axiom&tk=iHP9f6pQtaElHSpCBw3TSiLSy_7xsHPj01SelJ9qfWA%3D"

	ok, options, err := Verify(testKeyStr, signature)
	require.NoError(t, err)
	require.NotEmpty(t, options)

	assert.True(t, ok)

	assert.Equal(t, exp, options)
}

func TestVerifyToken(t *testing.T) {
	options := getOptions(t)

	ok, err := VerifyToken(testKeyStr, "iHP9f6pQtaElHSpCBw3TSiLSy_7xsHPj01SelJ9qfWA=", options)
	require.NoError(t, err)

	assert.True(t, ok)
}
