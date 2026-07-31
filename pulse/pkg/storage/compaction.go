package import (
	"sync"
	"time"
)

// CompactLogEntry represents a key-value record in a topic log segment.
type CompactLogEntry struct {
	Offset    int64     `json:"offset"`
	Key       string    `json:"key"`
	Value     []byte    `json:"value"` // Empty / nil value represents a tombstone deletion
	Timestamp time.Time `json:"timestamp"`
}

// LogCompactionEngine evaluates key-based log compaction, keeping only the latest entry per key.
type LogCompactionEngine struct {
	mu           sync.RWMutex
	retentionTTL time.Duration
	tombstoneTTL time.Duration
}

// NewLogCompactionEngine creates a LogCompactionEngine instance.
func NewLogCompactionEngine(retentionTTL, tombstoneTTL time.Duration) *LogCompactionEngine {
	if retentionTTL <= 0 {
		retentionTTL = 24 * time.Hour
	}
	if tombstoneTTL <= 0 {
		tombstoneTTL = 1 * time.Hour
	}
	return &LogCompactionEngine{
		retentionTTL: retentionTTL,
		tombstoneTTL: tombstoneTTL,
	}
}

// CompactSegment performs key-deduplication log compaction on log entries.
func (lce *LogCompactionEngine) CompactSegment(entries []CompactLogEntry, now time.Time) []CompactLogEntry {
	lce.mu.RLock()
	defer lce.mu.RUnlock()

	latestByKey := make(map[string]CompactLogEntry)

	for _, entry := range entries {
		if entry.Key == "" {
			continue
		}

		// Keep highest offset for given key
		existing, found := latestByKey[entry.Key]
		if !found || entry.Offset > existing.Offset {
			latestByKey[entry.Key] = entry
		}
	}

	var compacted []CompactLogEntry
	for _, entry := range latestByKey {
		// Filter out tombstones older than tombstoneTTL
		if len(entry.Value) == 0 {
			if now.Sub(entry.Timestamp) > lce.tombstoneTTL {
				continue // Purge tombstone
			}
		}

		// Filter out expired retention entries
		if lce.retentionTTL > 0 && now.Sub(entry.Timestamp) > lce.retentionTTL {
			continue
		}

		compacted = append(compacted, entry)
	}

	return compacted
}
