package zap

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/ingest"
	"github.com/axiomhq/axiom-go/internal/test/adapters"
	"github.com/axiomhq/axiom-go/internal/test/testhelper"
)

// TestNew makes sure New() picks up the "AXIOM_DATASET" environment variable.
func TestNew(t *testing.T) {
	testhelper.SafeClearEnv(t)

	t.Setenv("AXIOM_TOKEN", "xaat-test")
	t.Setenv("AXIOM_ORG_ID", "123")

	core, err := New()
	require.ErrorIs(t, err, ErrMissingDatasetName)
	require.Nil(t, core)

	t.Setenv("AXIOM_DATASET", "test")

	core, err = New()
	require.NoError(t, err)
	require.NotNil(t, core)
}

func TestCore(t *testing.T) {
	now := time.Now()

	exp := fmt.Sprintf(`{"_time":"%s","level":"info","key":"value","msg":"my message"}`,
		now.Format(time.RFC3339Nano))

	hasRun := false
	hf := func(w http.ResponseWriter, r *http.Request) {
		zsr, err := zstd.NewReader(r.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zsr)
		require.NoError(t, err)

		assert.JSONEq(t, exp, string(b))

		hasRun = true

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
	}

	logger, _ := adapters.Setup(t, hf, func(dataset string, client *axiom.Client) (*zap.Logger, func()) {
		t.Helper()

		core, err := New(
			SetClient(client),
			SetDataset(dataset),
		)
		require.NoError(t, err)

		logger := zap.New(core)
		t.Cleanup(func() {
			err := logger.Sync()
			require.NoError(t, err)
		})

		return logger, func() {}
	})

	// Timestamp field is set manually to make the JSONEq assertion pass.
	logger.Info("my message",
		zap.String("key", "value"),
		zap.Time(ingest.TimestampField, now),
	)

	require.NoError(t, logger.Sync())

	assert.True(t, hasRun)
}
