package pool

import (
	"context"
	"errors"
	"sync/atomic"
)

// HealthStats tracks connection health validation metrics.
type HealthStats struct {
	HealthyAcquires int64 `json:"healthy_acquires"`
	StaleDiscarded  int64 `json:"stale_discarded"`
}

// HealthChecker wraps a pool.Manager to validate connection health prior to checkout.
type HealthChecker struct {
	inner           Manager
	ValidateFn      func(*DbConn) bool
	healthyAcquires int64
	staleDiscarded  int64
}

// Verify HealthChecker implements Manager.
var _ Manager = (*HealthChecker)(nil)

// NewHealthChecker creates a HealthChecker wrapping the given pool Manager.
// If validateFn is nil, a default validation function returning true is used.
func NewHealthChecker(inner Manager, validateFn func(*DbConn) bool) *HealthChecker {
	if validateFn == nil {
		validateFn = func(conn *DbConn) bool { return true }
	}
	return &HealthChecker{
		inner:      inner,
		ValidateFn: validateFn,
	}
}

// Acquire retrieves a connection from the underlying pool and runs ValidateFn.
// If validation fails, the connection is discarded (released back) and retried up to 3 times total.
// Returns an error if all 3 attempts fail or if underlying Acquire fails.
func (hc *HealthChecker) Acquire() (*DbConn, error) {
	validate := hc.ValidateFn
	if validate == nil {
		validate = func(conn *DbConn) bool { return true }
	}

	const maxAttempts = 3
	for attempt := 0; attempt < maxAttempts; attempt++ {
		conn, err := hc.inner.Acquire()
		if err != nil {
			return nil, err
		}

		if validate(conn) {
			atomic.AddInt64(&hc.healthyAcquires, 1)
			return conn, nil
		}

		// Connection is unhealthy: increment stale count, discard, and retry
		atomic.AddInt64(&hc.staleDiscarded, 1)
		hc.inner.Release(conn)
	}

	return nil, errors.New("health check failed after 3 attempts: connections invalid")
}

// Release releases a connection back to the underlying pool.
func (hc *HealthChecker) Release(conn *DbConn) {
	hc.inner.Release(conn)
}

// IncrementQueries delegates query count increment to the underlying pool.
func (hc *HealthChecker) IncrementQueries() {
	hc.inner.IncrementQueries()
}

// Stats returns runtime pool metrics from the underlying pool.
func (hc *HealthChecker) Stats() PoolStats {
	return hc.inner.Stats()
}

// HealthStats returns health check validation counters.
func (hc *HealthChecker) HealthStats() HealthStats {
	return HealthStats{
		HealthyAcquires: atomic.LoadInt64(&hc.healthyAcquires),
		StaleDiscarded:  atomic.LoadInt64(&hc.staleDiscarded),
	}
}

// Dialect returns the database dialect of the underlying pool.
func (hc *HealthChecker) Dialect() string {
	return hc.inner.Dialect()
}

// Shutdown drains and shuts down the underlying pool.
func (hc *HealthChecker) Shutdown(ctx context.Context) error {
	return hc.inner.Shutdown(ctx)
}
