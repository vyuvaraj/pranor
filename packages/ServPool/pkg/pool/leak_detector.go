package pool

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// CheckedOutConnection tracks metadata for an active connection checkout.
type CheckedOutConnection struct {
	ConnID          string    `json:"conn_id"`
	CheckedOutAt    time.Time `json:"checked_out_at"`
	MaxAllowedHold  time.Duration `json:"max_allowed_hold"`
	AllocationStack string    `json:"allocation_stack"`
}

// ConnectionLeakDetector monitors long-held connections and forcefully reclaims leaked connection slots.
type ConnectionLeakDetector struct {
	mu               sync.RWMutex
	checkouts        map[string]*CheckedOutConnection // connID -> checkout info
	defaultMaxHold   time.Duration
	leakedCount      int64
	forcedReclaimFn  func(connID string)
}

// NewConnectionLeakDetector creates a ConnectionLeakDetector instance.
func NewConnectionLeakDetector(defaultMaxHold time.Duration, forcedReclaimFn func(connID string)) *ConnectionLeakDetector {
	if defaultMaxHold <= 0 {
		defaultMaxHold = 30 * time.Second
	}
	return &ConnectionLeakDetector{
		checkouts:       make(map[string]*CheckedOutConnection),
		defaultMaxHold:  defaultMaxHold,
		forcedReclaimFn: forcedReclaimFn,
	}
}

// RecordCheckout registers a connection checkout with allocation stack capture.
func (cld *ConnectionLeakDetector) RecordCheckout(connID string) {
	cld.mu.Lock()
	defer cld.mu.Unlock()

	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	stack := string(buf[:n])

	cld.checkouts[connID] = &CheckedOutConnection{
		ConnID:          connID,
		CheckedOutAt:    time.Now(),
		MaxAllowedHold:  cld.defaultMaxHold,
		AllocationStack: stack,
	}
}

// RecordCheckin unregisters a connection upon clean return to the pool.
func (cld *ConnectionLeakDetector) RecordCheckin(connID string) {
	cld.mu.Lock()
	defer cld.mu.Unlock()
	delete(cld.checkouts, connID)
}

// SweepLeaks scans active checkouts and forcefully reclaims any connections exceeding max allowed hold.
func (cld *ConnectionLeakDetector) SweepLeaks(now time.Time) []string {
	cld.mu.Lock()
	var leakedIDs []string

	for connID, info := range cld.checkouts {
		if now.Sub(info.CheckedOutAt) > info.MaxAllowedHold {
			leakedIDs = append(leakedIDs, connID)
			delete(cld.checkouts, connID)
			cld.leakedCount++
			if cld.forcedReclaimFn != nil {
				cld.forcedReclaimFn(connID)
			}
		}
	}
	cld.mu.Unlock()

	return leakedIDs
}

// GetLeakedCount returns total number of forcefully reclaimed leaked connections.
func (cld *ConnectionLeakDetector) GetLeakedCount() int64 {
	cld.mu.RLock()
	defer cld.mu.RUnlock()
	return cld.leakedCount
}

// StartBackgroundSweeper starts periodic leak detection background scan.
func (cld *ConnectionLeakDetector) StartBackgroundSweeper(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				leaks := cld.SweepLeaks(t)
				if len(leaks) > 0 {
					fmt.Printf("[ServPool LeakDetector] Reclaimed %d leaked connections: %v\n", len(leaks), leaks)
				}
			}
		}
	}()
}
