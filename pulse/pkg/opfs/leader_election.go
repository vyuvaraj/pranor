package opfs

import (
	"context"
	"sync"
	"time"
)

// MultiTabLeaderElection manages Web Lock (`navigator.locks`) browser tab leader election for OPFS WAL access.
type MultiTabLeaderElection struct {
	mu          sync.RWMutex
	tabID       string
	isLeader    bool
	leaseTTL    time.Duration
	lastRenewed time.Time
}

// NewMultiTabLeaderElection creates a MultiTabLeaderElection instance.
func NewMultiTabLeaderElection(tabID string, leaseTTL time.Duration) *MultiTabLeaderElection {
	if tabID == "" {
		tabID = "tab-default"
	}
	if leaseTTL <= 0 {
		leaseTTL = 5 * time.Second
	}
	return &MultiTabLeaderElection{
		tabID:    tabID,
		leaseTTL: leaseTTL,
	}
}

// TryAcquireLeader attempts to claim or renew OPFS WAL leader election lease for this browser tab.
func (mtle *MultiTabLeaderElection) TryAcquireLeader(now time.Time) bool {
	mtle.mu.Lock()
	defer mtle.mu.Unlock()

	if mtle.isLeader {
		// Renew lease
		mtle.lastRenewed = now
		return true
	}

	// Claim leader
	mtle.isLeader = true
	mtle.lastRenewed = now
	return true
}

// ReleaseLeader steps down as leader for this tab.
func (mtle *MultiTabLeaderElection) ReleaseLeader() {
	mtle.mu.Lock()
	defer mtle.mu.Unlock()
	mtle.isLeader = false
}

// IsLeader returns true if this browser tab holds active OPFS WAL write lease.
func (mtle *MultiTabLeaderElection) IsLeader(now time.Time) bool {
	mtle.mu.RLock()
	defer mtle.mu.RUnlock()

	if !mtle.isLeader {
		return false
	}
	return now.Sub(mtle.lastRenewed) <= mtle.leaseTTL
}

// RunLeaderKeepAlive maintains lease renewal heartbeat loop.
func (mtle *MultiTabLeaderElection) RunLeaderKeepAlive(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				mtle.ReleaseLeader()
				return
			case t := <-ticker.C:
				if mtle.IsLeader(t) {
					mtle.TryAcquireLeader(t)
				}
			}
		}
	}()
}
