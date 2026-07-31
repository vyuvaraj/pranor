package import (
	"fmt"
	"sync"
)

type QuotaPolicy struct {
	MaxBytes        int64 `json:"max_bytes"`
	MaxSyncedRecords int   `json:"max_synced_records"`
}

type Compactor struct {
	mu     sync.Mutex
	policy QuotaPolicy
}

func NewCompactor(policy QuotaPolicy) *Compactor {
	if policy.MaxBytes <= 0 {
		policy.MaxBytes = 50 * 1024 * 1024 // 50MB default browser limit
	}
	if policy.MaxSyncedRecords <= 0 {
		policy.MaxSyncedRecords = 10000
	}
	return &Compactor{
		policy: policy,
	}
}

// CompactPurgeSynced filters out acknowledged records that exceed quota limits.
func (c *Compactor) CompactPurgeSynced(entries []LogEntry) ([]LogEntry, int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var retained []LogEntry
	purgedCount := 0

	var totalBytes int64
	for _, entry := range entries {
		totalBytes += int64(len(entry.Payload) + len(entry.Topic))
	}

	for _, entry := range entries {
		// Purge synced records if storage budget exceeded
		if entry.Synced && (totalBytes > c.policy.MaxBytes || len(retained) > c.policy.MaxSyncedRecords) {
			totalBytes -= int64(len(entry.Payload) + len(entry.Topic))
			purgedCount++
			continue
		}
		retained = append(retained, entry)
	}

	return retained, purgedCount
}

// Compact triggers storage driver compaction via Engine
func (e *Engine) Compact(policy QuotaPolicy) (int, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	compactor := NewCompactor(policy)
	allEntries, err := e.driver.Recover()
	if err != nil {
		return 0, fmt.Errorf("compaction recover failed: %w", err)
	}

	retained, purged := compactor.CompactPurgeSynced(allEntries)
	if purged > 0 {
		// Re-initialize synced offsets with retained slice
		var syncedOffsets []uint64
		for _, r := range retained {
			if r.Synced {
				syncedOffsets = append(syncedOffsets, r.Offset)
			}
		}
		_ = e.driver.MarkSynced(syncedOffsets)
	}

	return purged, nil
}
