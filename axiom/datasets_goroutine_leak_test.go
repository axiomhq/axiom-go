package axiom

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"
)

func TestIngestEvents_NoGoroutineLeak(t *testing.T) {
	t.Parallel()

	// Server that accepts ingest requests with a small delay to simulate real latency.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ingested":       1,
			"failed":         0,
			"failures":       []any{},
			"processedBytes": 42,
		})
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(
		SetURL(server.URL),
		SetToken("xaat-test-token"),
		SetOrganizationID("test-org"),
		SetNoRetry(),
		SetNoEnv(),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	// Warm up: one ingest to initialize pools, etc.
	events := []Event{{"msg": "hello"}}
	if _, err := client.IngestEvents(t.Context(), "test-dataset", events); err != nil {
		t.Fatalf("warm-up IngestEvents: %v", err)
	}

	// Let any transient goroutines settle.
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	baseGoroutines := runtime.NumGoroutine()

	// Run many ingest calls.
	const iterations = 100
	for i := range iterations {
		ctx := t.Context()
		ev := []Event{{"i": i, "msg": "test event"}}
		if _, err := client.IngestEvents(ctx, "test-dataset", ev); err != nil {
			t.Fatalf("IngestEvents iteration %d: %v", i, err)
		}
	}

	// Allow goroutines to clean up.
	runtime.GC()
	time.Sleep(500 * time.Millisecond)
	runtime.GC()
	time.Sleep(200 * time.Millisecond)

	finalGoroutines := runtime.NumGoroutine()
	leaked := finalGoroutines - baseGoroutines

	// Allow a small margin for runtime jitter, but 100 iterations should not
	// leave dozens of goroutines behind.
	const maxAcceptableLeak = 5
	if leaked > maxAcceptableLeak {
		t.Errorf("goroutine leak: started with %d, ended with %d, leaked %d (max acceptable: %d)",
			baseGoroutines, finalGoroutines, leaked, maxAcceptableLeak)
	}
}

func TestIngestEvents_NoGoroutineLeakOnRetry(t *testing.T) {
	t.Parallel()

	// Server that fails the first request, then succeeds — simulating a retry.
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount%2 == 1 {
			// First request: fail with 500 to trigger retry.
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ingested":       1,
			"failed":         0,
			"failures":       []any{},
			"processedBytes": 42,
		})
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(
		SetURL(server.URL),
		SetToken("xaat-test-token"),
		SetOrganizationID("test-org"),
		SetNoEnv(),
		// NOTE: retries enabled (default)
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	// Warm up.
	requestCount = 0
	events := []Event{{"msg": "hello"}}
	_, _ = client.IngestEvents(t.Context(), "test-dataset", events)

	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	baseGoroutines := runtime.NumGoroutine()

	// Run ingest calls that each trigger a retry (2 goroutines spawned per call).
	const iterations = 50
	requestCount = 0
	for i := range iterations {
		ctx := t.Context()
		ev := []Event{{"i": i}}
		_, _ = client.IngestEvents(ctx, "test-dataset", ev)
	}

	runtime.GC()
	time.Sleep(500 * time.Millisecond)
	runtime.GC()
	time.Sleep(200 * time.Millisecond)

	finalGoroutines := runtime.NumGoroutine()
	leaked := finalGoroutines - baseGoroutines

	const maxAcceptableLeak = 5
	if leaked > maxAcceptableLeak {
		t.Errorf("goroutine leak on retry: started with %d, ended with %d, leaked %d (max acceptable: %d)",
			baseGoroutines, finalGoroutines, leaked, maxAcceptableLeak)
	}
}
