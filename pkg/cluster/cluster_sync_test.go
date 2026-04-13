package cluster

import (
	"testing"
	"time"
)

func TestRequestClusterSyncWaitsForCompletion(t *testing.T) {
	originalSyncNow := syncNow
	syncNow = make(chan chan error, 1)
	defer func() {
		syncNow = originalSyncNow
	}()

	doneSignal := make(chan struct{})

	go func() {
		done := <-syncNow
		time.Sleep(20 * time.Millisecond)
		done <- nil
		close(done)
		close(doneSignal)
	}()

	start := time.Now()
	if err := requestClusterSync(true); err != nil {
		t.Fatalf("requestClusterSync(true) error = %v", err)
	}

	if time.Since(start) < 20*time.Millisecond {
		t.Fatalf("requestClusterSync(true) returned before sync completed")
	}

	select {
	case <-doneSignal:
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("sync worker did not finish")
	}
}

func TestRequestClusterSyncAsyncReturnsImmediately(t *testing.T) {
	originalSyncNow := syncNow
	syncNow = make(chan chan error, 1)
	defer func() {
		syncNow = originalSyncNow
	}()

	start := time.Now()
	if err := requestClusterSync(false); err != nil {
		t.Fatalf("requestClusterSync(false) error = %v", err)
	}

	if time.Since(start) > 10*time.Millisecond {
		t.Fatalf("requestClusterSync(false) did not return immediately")
	}

	select {
	case done := <-syncNow:
		if done != nil {
			t.Fatalf("async sync request should not provide a completion channel")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("sync request was not enqueued")
	}
}
