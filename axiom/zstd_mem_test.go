package axiom

import (
	"encoding/json"
	"io"
	"runtime"
	"runtime/debug"
	"sync"
	"testing"

	"github.com/klauspost/compress/zstd"
)

// TestZstdPeakHeapUsage demonstrates that the current zstd.NewWriter-per-flush
// pattern causes peak heap usage >300 MB even when flushing tiny log batches.
//
// This simulates real production behavior: the slog adapter flushes every 1s,
// and under concurrent HTTP requests, multiple IngestEvents calls can overlap.
// Each creates a ~4 MB zstd.Encoder that lives until GC collects it.
//
// With GC disabled (simulating GC lag under memory pressure), the encoders
// pile up and peak heap climbs to hundreds of MB.
func TestZstdPeakHeapUsage(t *testing.T) {
	// Disable GC to simulate what happens under memory pressure when GC
	// can't keep up — this is exactly what causes OOM in production.
	debug.SetGCPercent(-1)
	t.Cleanup(func() { debug.SetGCPercent(100) })

	const (
		// Simulate 150 seconds of the slog adapter (flush every 1s).
		// In production, IngestEvents is called from IngestChannel's flush().
		flushCount = 150
		// Small batch — even 5 events per flush triggers the full encoder alloc.
		eventsPerFlush = 5
	)

	events := make([]Event, eventsPerFlush)
	for i := range events {
		events[i] = Event{
			"_time": "2026-03-01T12:00:00Z",
			"level": "INFO",
			"msg":   "request",
		}
	}

	var peakHeapMB float64

	runtime.GC()
	for i := range flushCount {
		// Create a new zstd.Writer per flush — exactly what IngestEvents does.
		pr, pw := io.Pipe()
		zsw, err := zstd.NewWriter(pw)
		if err != nil {
			t.Fatal(err)
		}
		go func() {
			enc := json.NewEncoder(zsw)
			for _, e := range events {
				_ = enc.Encode(e)
			}
			zsw.Close()
			pw.Close()
		}()
		_, _ = io.ReadAll(pr)
		pr.Close()

		// Sample heap every 10 flushes.
		if i%10 == 0 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			heapMB := float64(m.HeapAlloc) / 1024 / 1024
			if heapMB > peakHeapMB {
				peakHeapMB = heapMB
			}
		}
	}

	// Final measurement.
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	finalHeapMB := float64(m.HeapAlloc) / 1024 / 1024
	if finalHeapMB > peakHeapMB {
		peakHeapMB = finalHeapMB
	}
	totalAllocMB := float64(m.TotalAlloc) / 1024 / 1024

	t.Logf("Current behavior (zstd.NewWriter per flush):")
	t.Logf("  Flushes:      %d (each with only %d tiny events)", flushCount, eventsPerFlush)
	t.Logf("  Peak heap:    %.0f MB", peakHeapMB)
	t.Logf("  Total alloc:  %.0f MB", totalAllocMB)

	if peakHeapMB < 300 {
		t.Errorf("Expected peak heap >= 300 MB, got %.0f MB (increase flushCount if needed)", peakHeapMB)
	}
}

// TestZstdPeakHeapUsage_Concurrent shows that concurrent flushes (which happen
// in production when multiple HTTP requests trigger tool calls simultaneously)
// make the problem worse — multiple encoders are alive at once.
func TestZstdPeakHeapUsage_Concurrent(t *testing.T) {
	debug.SetGCPercent(-1)
	t.Cleanup(func() { debug.SetGCPercent(100) })

	const (
		flushCount     = 100
		concurrency    = 4 // simulates 4 concurrent IngestEvents calls
		eventsPerFlush = 5
	)

	events := make([]Event, eventsPerFlush)
	for i := range events {
		events[i] = Event{
			"_time": "2026-03-01T12:00:00Z",
			"level": "INFO",
			"msg":   "request",
		}
	}

	runtime.GC()

	var wg sync.WaitGroup
	for range concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range flushCount {
				pr, pw := io.Pipe()
				zsw, err := zstd.NewWriter(pw)
				if err != nil {
					return
				}
				go func() {
					enc := json.NewEncoder(zsw)
					for _, e := range events {
						_ = enc.Encode(e)
					}
					zsw.Close()
					pw.Close()
				}()
				_, _ = io.ReadAll(pr)
				pr.Close()
			}
		}()
	}
	wg.Wait()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	heapMB := float64(m.HeapAlloc) / 1024 / 1024
	totalAllocMB := float64(m.TotalAlloc) / 1024 / 1024

	t.Logf("Concurrent behavior (%d goroutines × %d flushes):", concurrency, flushCount)
	t.Logf("  Events per flush: %d (tiny)", eventsPerFlush)
	t.Logf("  Heap after:       %.0f MB", heapMB)
	t.Logf("  Total alloc:      %.0f MB", totalAllocMB)
}

// TestZstdPeakHeapUsage_Pooled shows the fix: pooled writers keep heap flat.
func TestZstdPeakHeapUsage_Pooled(t *testing.T) {
	debug.SetGCPercent(-1)
	t.Cleanup(func() { debug.SetGCPercent(100) })

	const (
		flushCount     = 120
		eventsPerFlush = 5
	)

	events := make([]Event, eventsPerFlush)
	for i := range events {
		events[i] = Event{
			"_time": "2026-03-01T12:00:00Z",
			"level": "INFO",
			"msg":   "request",
		}
	}

	pool := sync.Pool{
		New: func() any {
			w, _ := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedFastest))
			return w
		},
	}

	runtime.GC()
	var peakHeapMB float64

	for i := range flushCount {
		pr, pw := io.Pipe()
		zsw := pool.Get().(*zstd.Encoder)
		zsw.Reset(pw)
		go func() {
			enc := json.NewEncoder(zsw)
			for _, e := range events {
				_ = enc.Encode(e)
			}
			zsw.Close()
			pool.Put(zsw)
			pw.Close()
		}()
		_, _ = io.ReadAll(pr)
		pr.Close()

		if i%10 == 0 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			heapMB := float64(m.HeapAlloc) / 1024 / 1024
			if heapMB > peakHeapMB {
				peakHeapMB = heapMB
			}
		}
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	finalHeapMB := float64(m.HeapAlloc) / 1024 / 1024
	if finalHeapMB > peakHeapMB {
		peakHeapMB = finalHeapMB
	}
	totalAllocMB := float64(m.TotalAlloc) / 1024 / 1024

	t.Logf("Fixed behavior (pooled zstd.Writer):")
	t.Logf("  Flushes:      %d (each with only %d tiny events)", flushCount, eventsPerFlush)
	t.Logf("  Peak heap:    %.0f MB", peakHeapMB)
	t.Logf("  Total alloc:  %.0f MB", totalAllocMB)
}
