package pool

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	)

// TestEmpirical_HealthChecker_Retries verifies checkout retries up to attempt 3.
func TestEmpirical_HealthChecker_Retries(t *testing.T) {
	t.Run("Succeeds on attempt 1", func(t *testing.T) {
		inner := NewConnectionPool(5, "postgres")
		defer inner.Shutdown(context.Background())

		calls := 0
		hc := NewHealthChecker(inner, func(conn *DbConn) bool {
			calls++
			return true
		})

		conn, err := hc.Acquire()
		if err != nil {
			t.Fatalf("expected Acquire success, got err: %v", err)
		}
		if conn == nil {
			t.Fatalf("expected non-nil conn")
		}
		hc.Release(conn)

		if calls != 1 {
			t.Errorf("expected 1 validate call, got %d", calls)
		}

		stats := hc.HealthStats()
		if stats.HealthyAcquires != 1 || stats.StaleDiscarded != 0 {
			t.Errorf("unexpected stats: %+v", stats)
		}
	})

	t.Run("Succeeds on attempt 2 (1 retry)", func(t *testing.T) {
		inner := NewConnectionPool(5, "postgres")
		defer inner.Shutdown(context.Background())

		calls := 0
		hc := NewHealthChecker(inner, func(conn *DbConn) bool {
			calls++
			return calls > 1 // Fail attempt 1, pass attempt 2
		})

		conn, err := hc.Acquire()
		if err != nil {
			t.Fatalf("expected Acquire success on retry, got err: %v", err)
		}
		if conn == nil {
			t.Fatalf("expected non-nil conn")
		}
		hc.Release(conn)

		if calls != 2 {
			t.Errorf("expected 2 validate calls, got %d", calls)
		}

		stats := hc.HealthStats()
		if stats.HealthyAcquires != 1 || stats.StaleDiscarded != 1 {
			t.Errorf("unexpected stats: %+v", stats)
		}
	})

	t.Run("Succeeds on attempt 3 (2 retries)", func(t *testing.T) {
		inner := NewConnectionPool(5, "postgres")
		defer inner.Shutdown(context.Background())

		calls := 0
		hc := NewHealthChecker(inner, func(conn *DbConn) bool {
			calls++
			return calls >= 3 // Fail attempts 1 & 2, pass attempt 3
		})

		conn, err := hc.Acquire()
		if err != nil {
			t.Fatalf("expected Acquire success on attempt 3, got err: %v", err)
		}
		if conn == nil {
			t.Fatalf("expected non-nil conn")
		}
		hc.Release(conn)

		if calls != 3 {
			t.Errorf("expected 3 validate calls, got %d", calls)
		}

		stats := hc.HealthStats()
		if stats.HealthyAcquires != 1 || stats.StaleDiscarded != 2 {
			t.Errorf("unexpected stats: %+v", stats)
		}
	})
}

// TestEmpirical_HealthChecker_FailureAfter3Attempts verifies error, counters, and connection cleanup when all 3 attempts fail.
func TestEmpirical_HealthChecker_FailureAfter3Attempts(t *testing.T) {
	inner := NewConnectionPool(5, "postgres")
	defer inner.Shutdown(context.Background())

	calls := 0
	hc := NewHealthChecker(inner, func(conn *DbConn) bool {
		calls++
		return false // Always fail validation
	})

	conn, err := hc.Acquire()
	if err == nil {
		t.Fatalf("expected error after 3 failed attempts, got nil error")
	}
	if conn != nil {
		t.Fatalf("expected nil conn on failure")
	}

	if !strings.Contains(err.Error(), "3 attempts") {
		t.Errorf("expected error message to mention 3 attempts, got %q", err.Error())
	}

	if calls != 3 {
		t.Errorf("expected exactly 3 validate calls before giving up, got %d", calls)
	}

	stats := hc.HealthStats()
	if stats.HealthyAcquires != 0 {
		t.Errorf("expected HealthyAcquires == 0, got %d", stats.HealthyAcquires)
	}
	if stats.StaleDiscarded != 3 {
		t.Errorf("expected StaleDiscarded == 3, got %d", stats.StaleDiscarded)
	}

	// Crucial check: verify all discarded connections were properly released back to the pool
	// and no active connection leak remains in inner 
	poolStats := inner.Stats()
	if poolStats.ActiveConnections != 0 {
		t.Errorf("expected 0 active connections after failure, got %d (connection leak!)", poolStats.ActiveConnections)
	}
}

// TestEmpirical_HealthChecker_NilValidateFn verifies behavior when ValidateFn is nil.
func TestEmpirical_HealthChecker_NilValidateFn(t *testing.T) {
	inner := NewConnectionPool(5, "postgres")
	defer inner.Shutdown(context.Background())

	// Explicitly pass nil validateFn
	hc := NewHealthChecker(inner, nil)
	conn, err := hc.Acquire()
	if err != nil || conn == nil {
		t.Fatalf("expected Acquire success with nil ValidateFn")
	}
	hc.Release(conn)

	// Set hc.ValidateFn to nil dynamically
	hc.ValidateFn = nil
	conn2, err := hc.Acquire()
	if err != nil || conn2 == nil {
		t.Fatalf("expected Acquire success with dynamically nil ValidateFn")
	}
	hc.Release(conn2)

	stats := hc.HealthStats()
	if stats.HealthyAcquires != 2 || stats.StaleDiscarded != 0 {
		t.Errorf("unexpected stats: %+v", stats)
	}
}

// TestEmpirical_HealthChecker_ConcurrentAcquireRelease verifies race safety under concurrent load.
func TestEmpirical_HealthChecker_ConcurrentAcquireRelease(t *testing.T) {
	inner := NewConnectionPool(20, "postgres")
	defer inner.Shutdown(context.Background())

	var validateCounter int64

	hc := NewHealthChecker(inner, func(conn *DbConn) bool {
		val := atomic.AddInt64(&validateCounter, 1)
		// Reject every 3rd connection attempt
		return val%3 != 0
	})

	const numGoroutines = 50
	const opsPerGoroutine = 20
	var wg sync.WaitGroup

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for op := 0; op < opsPerGoroutine; op++ {
				conn, err := hc.Acquire()
				if err == nil && conn != nil {
					// Simulate minor work
					time.Sleep(10 * time.Microsecond)
					hc.Release(conn)
				}
			}
		}()
	}
	wg.Wait()

	stats := hc.HealthStats()
	if stats.HealthyAcquires == 0 {
		t.Errorf("expected >0 healthy acquires in concurrent test")
	}
	if stats.StaleDiscarded == 0 {
		t.Errorf("expected >0 stale discarded in concurrent test")
	}

	// Verify no leaked active connections
	pStats := hc.Stats()
	if pStats.ActiveConnections != 0 {
		t.Errorf("expected 0 active connections after all goroutines finish, got %d", pStats.ActiveConnections)
	}
}
