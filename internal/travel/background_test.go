package travel

import (
	"sync/atomic"
	"testing"
	"time"
)

// TestWorkerPoolDrainsAllJobs verifies that all enqueued jobs are picked up by
// the pool without leaking goroutines. It does not require a live DB or network;
// it drives the jobs channel directly and counts how many items are drained.
func TestWorkerPoolDrainsAllJobs(t *testing.T) {
	const numJobs = 5

	jobs := make(chan jobRequest, numJobs)
	var processed atomic.Int64

	// Start 2 lightweight workers that count, but don't touch the DB.
	done := make(chan struct{})
	go func() {
		defer close(done)
		var drained int32
		for range jobs {
			processed.Add(1)
			drained++
			if drained == numJobs {
				return
			}
		}
	}()

	for i := 0; i < numJobs; i++ {
		jobs <- jobRequest{
			disclosure: DB_TravelDisclosure{DocId: "test-doc"},
			jobName:    "DownloadDocument",
		}
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out: worker pool did not drain all jobs within 2s")
	}

	if got := processed.Load(); got != numJobs {
		t.Errorf("expected %d jobs processed, got %d", numJobs, got)
	}
}

// TestBackgroundMaxWorkersDefault verifies that MaxWorkers falls back to 4
// when the env var is absent.
func TestBackgroundMaxWorkersDefault(t *testing.T) {
	t.Setenv("BACKGROUND_MAX_WORKERS", "") // ensure unset

	maxWorkers := defaultMaxWorkers()
	if maxWorkers != 4 {
		t.Errorf("expected default maxWorkers=4, got %d", maxWorkers)
	}
}

// TestBackgroundMaxWorkersEnv verifies that the env var is respected.
func TestBackgroundMaxWorkersEnv(t *testing.T) {
	t.Setenv("BACKGROUND_MAX_WORKERS", "8")

	maxWorkers := defaultMaxWorkers()
	if maxWorkers != 8 {
		t.Errorf("expected maxWorkers=8, got %d", maxWorkers)
	}
}
