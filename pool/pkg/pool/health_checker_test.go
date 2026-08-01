package pool

import (
	"context"
	"sync"
	"testing"
	"time"

	)

func TestHealthChecker_HealthyConnPasses(t *testing.T) {
	inner := NewConnectionPool(5, "postgres")
	defer inner.Shutdown(context.Background())

	// Default validation (nil validateFn)
	hc := NewHealthChecker(inner, nil)

	conn, err := hc.Acquire()
	if err != nil {
		t.Fatalf("expected acquire success, got err: %v", err)
	}
	if conn == nil {
		t.Fatalf("expected non-nil connection")
	}

	stats := hc.HealthStats()
	if stats.HealthyAcquires != 1 {
		t.Errorf("expected HealthyAcquires == 1, got %d", stats.HealthyAcquires)
	}
	if stats.StaleDiscarded != 0 {
		t.Errorf("expected StaleDiscarded == 0, got %d", stats.StaleDiscarded)
	}

	hc.Release(conn)
}

func TestHealthChecker_DiscardAndRetry(t *testing.T) {
	inner := NewConnectionPool(5, "postgres")
	defer inner.Shutdown(context.Background())

	attempts := 0
	// Reject the first 2 connections, pass the 3rd
	validateFn := func(conn *DbConn) bool {
		attempts++
		return attempts >= 3
	}

	hc := NewHealthChecker(inner, validateFn)

	conn, err := hc.Acquire()
	if err != nil {
		t.Fatalf("expected eventual acquire success, got err: %v", err)
	}
	if conn == nil {
		t.Fatalf("expected non-nil connection")
	}

	stats := hc.HealthStats()
	if stats.HealthyAcquires != 1 {
		t.Errorf("expected HealthyAcquires == 1, got %d", stats.HealthyAcquires)
	}
	if stats.StaleDiscarded != 2 {
		t.Errorf("expected StaleDiscarded == 2, got %d", stats.StaleDiscarded)
	}

	hc.Release(conn)
}

func TestHealthChecker_AllUnhealthyReturnsError(t *testing.T) {
	inner := NewConnectionPool(5, "postgres")
	defer inner.Shutdown(context.Background())

	// Always fail validation
	validateFn := func(conn *DbConn) bool {
		return false
	}

	hc := NewHealthChecker(inner, validateFn)

	conn, err := hc.Acquire()
	if err == nil {
		t.Fatalf("expected error when all connections are unhealthy, got nil error")
	}
	if conn != nil {
		t.Fatalf("expected nil connection on error")
	}

	stats := hc.HealthStats()
	if stats.HealthyAcquires != 0 {
		t.Errorf("expected HealthyAcquires == 0, got %d", stats.HealthyAcquires)
	}
	if stats.StaleDiscarded != 3 {
		t.Errorf("expected StaleDiscarded == 3 after 3 failed attempts, got %d", stats.StaleDiscarded)
	}
}

func TestHealthChecker_DelegatedMethods(t *testing.T) {
	inner := NewConnectionPool(5, "postgres")
	hc := NewHealthChecker(inner, nil)

	if hc.Dialect() != "postgres" {
		t.Errorf("Dialect() = %s, expected postgres", hc.Dialect())
	}

	hc.IncrementQueries()
	pStats := hc.Stats()
	if pStats.TotalQueries != 1 {
		t.Errorf("TotalQueries = %d, expected 1", pStats.TotalQueries)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := hc.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown error: %v", err)
	}
}

func TestHealthChecker_InnerAcquireError(t *testing.T) {
	inner := NewConnectionPool(5, "postgres")
	hc := NewHealthChecker(inner, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_ = hc.Shutdown(ctx)

	conn, err := hc.Acquire()
	if err == nil {
		t.Fatalf("expected error on Acquire after Shutdown, got nil")
	}
	if conn != nil {
		t.Fatalf("expected nil connection on acquire error")
	}
}

func TestHealthChecker_DynamicValidateFn(t *testing.T) {
	inner := NewConnectionPool(5, "postgres")
	defer inner.Shutdown(context.Background())

	hc := NewHealthChecker(inner, nil)
	conn1, err := hc.Acquire()
	if err != nil || conn1 == nil {
		t.Fatalf("expected acquire success")
	}
	hc.Release(conn1)

	// Dynamically change ValidateFn to fail validation
	hc.ValidateFn = func(c *DbConn) bool { return false }

	conn2, err := hc.Acquire()
	if err == nil {
		t.Fatalf("expected error after changing ValidateFn to reject connections")
	}
	if conn2 != nil {
		t.Fatalf("expected nil conn on failure")
	}
}

func TestHealthChecker_ConcurrentAcquire(t *testing.T) {
	inner := NewConnectionPool(10, "postgres")
	defer inner.Shutdown(context.Background())

	hc := NewHealthChecker(inner, func(c *DbConn) bool {
		return c.ID%2 == 0 // Reject odd IDs
	})

	const numGoroutines = 10
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := hc.Acquire()
			if err == nil && conn != nil {
				if conn.ID%2 != 0 {
					t.Errorf("expected even conn ID, got odd: %d", conn.ID)
				}
				hc.Release(conn)
			}
		}()
	}
	wg.Wait()

	hStats := hc.HealthStats()
	if hStats.HealthyAcquires+hStats.StaleDiscarded == 0 {
		t.Errorf("expected non-zero health stats")
	}
}

