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
	exp := "v1.4.0-20201118T1633+1f878f59d"

	hf := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := fmt.Fprint(w, `{
			"currentVersion": "v1.4.0-20201118T1633+1f878f59d"
		}`)
		assert.NoError(t, err)
	}

	client, teardown := setup(t, "/api/v1/version", hf)
	defer teardown()

	res, err := client.Version.Get(context.Background())
	require.NoError(t, err)

	assert.Equal(t, exp, res)
}
