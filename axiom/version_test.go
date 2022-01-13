package axiom

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionService_Get(t *testing.T) {
	exp := "1.17.0"

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `{
			"currentVersion": "1.17.0"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/version", hf)
	defer teardown()

	res, err := client.Version.Get(context.Background())
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}
