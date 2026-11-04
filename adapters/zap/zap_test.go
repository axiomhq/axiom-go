package zap

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
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
			assert.NoError(t, err)
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

type rejectingRoundTripper struct {
	requests atomic.Int32
}

func (rt *rejectingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.requests.Add(1)
	_ = req.Body.Close()
	return &http.Response{StatusCode: http.StatusServiceUnavailable, Header: http.Header{}, Body: http.NoBody, Request: req}, nil
}

func TestWriteSyncer_Sync_FailedIngest(t *testing.T) {
	rt := &rejectingRoundTripper{}
	client, err := axiom.NewClient(
		axiom.SetURL("http://axiom.local"),
		axiom.SetToken("xaat-test"),
		axiom.SetClient(&http.Client{Transport: rt}),
		axiom.SetNoEnv(),
	)
	require.NoError(t, err)

	ws := &WriteSyncer{client: client, datasetName: "test", maxBufferCapacity: 1 << 30}

	payload := make([]byte, 8<<20)
	_, err = rand.Read(payload)
	require.NoError(t, err)

	_, err = ws.Write(payload)
	require.NoError(t, err)
	require.Error(t, ws.Sync())
	assert.EqualValues(t, 1, rt.requests.Load())

	_, err = ws.Write([]byte("next batch"))
	require.NoError(t, err)
	assert.Equal(t, "next batch", ws.buf.String())
}
