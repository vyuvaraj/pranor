package pool

import (
	"sync"
	"testing"
	"time"
)

func TestConnectionLeakDetector_SweepLeaks(t *testing.T) {
	reclaimed := make(map[string]bool)
	var mu sync.Mutex

	reclaimFn := func(connID string) {
		mu.Lock()
		defer mu.Unlock()
		reclaimed[connID] = true
	}

	detector := NewConnectionLeakDetector(100*time.Millisecond, reclaimFn)

	// Normal connection checkout & checkin
	detector.RecordCheckout("conn-1")
	detector.RecordCheckin("conn-1")

	// Leaked connection (not checked in)
	detector.RecordCheckout("conn-leaked-2")

	// Sweep immediately -> nothing leaked yet
	leaked := detector.SweepLeaks(time.Now())
	if len(leaked) > 0 {
		t.Errorf("expected no leaks immediately after checkout")
	}

	// Sweep after threshold
	futureTime := time.Now().Add(200 * time.Millisecond)
	leaked = detector.SweepLeaks(futureTime)

	if len(leaked) != 1 || leaked[0] != "conn-leaked-2" {
		t.Fatalf("expected conn-leaked-2 to be reclaimed, got %v", leaked)
	}

	mu.Lock()
	reclaimedOk := reclaimed["conn-leaked-2"]
	mu.Unlock()

	if !reclaimedOk {
		t.Error("expected reclaim function callback to be invoked for leaked connection")
	}

	if detector.GetLeakedCount() != 1 {
		t.Errorf("expected leaked count 1, got %d", detector.GetLeakedCount())
	}
}
