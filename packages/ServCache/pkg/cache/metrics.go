package cache

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

// MetricsCollector tracks real-time cache hit, miss, eviction, and memory telemetry.
type MetricsCollector struct {
	hits       uint64
	misses     uint64
	evictions  uint64
	sets       uint64
	deletes    uint64
	startTime  time.Time
}

// MetricsSnapshot is a serializable view of ServCache telemetry.
type MetricsSnapshot struct {
	Hits           uint64  `json:"hits"`
	Misses         uint64  `json:"misses"`
	Evictions      uint64  `json:"evictions"`
	Sets           uint64  `json:"sets"`
	Deletes        uint64  `json:"deletes"`
	TotalRequests  uint64  `json:"total_requests"`
	HitRate        float64 `json:"hit_rate"`
	UptimeSeconds  float64 `json:"uptime_seconds"`
}

// NewMetricsCollector initializes a telemetry collector.
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		startTime: time.Now(),
	}
}

// RecordHit increments the hit counter.
func (m *MetricsCollector) RecordHit() {
	atomic.AddUint64(&m.hits, 1)
}

// RecordMiss increments the miss counter.
func (m *MetricsCollector) RecordMiss() {
	atomic.AddUint64(&m.misses, 1)
}

// RecordEviction increments the eviction counter.
func (m *MetricsCollector) RecordEviction() {
	atomic.AddUint64(&m.evictions, 1)
}

// RecordSet increments the set counter.
func (m *MetricsCollector) RecordSet() {
	atomic.AddUint64(&m.sets, 1)
}

// RecordDelete increments the delete counter.
func (m *MetricsCollector) RecordDelete() {
	atomic.AddUint64(&m.deletes, 1)
}

// Snapshot returns the current metrics values.
func (m *MetricsCollector) Snapshot() MetricsSnapshot {
	hits := atomic.LoadUint64(&m.hits)
	misses := atomic.LoadUint64(&m.misses)
	evictions := atomic.LoadUint64(&m.evictions)
	sets := atomic.LoadUint64(&m.sets)
	deletes := atomic.LoadUint64(&m.deletes)

	total := hits + misses
	var hitRate float64
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}

	uptime := time.Since(m.startTime).Seconds()

	return MetricsSnapshot{
		Hits:          hits,
		Misses:        misses,
		Evictions:     evictions,
		Sets:          sets,
		Deletes:       deletes,
		TotalRequests: total,
		HitRate:       hitRate,
		UptimeSeconds: uptime,
	}
}

// HTTPHandler returns an http.Handler that outputs live JSON metrics for ServConsole integration.
func (m *MetricsCollector) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(m.Snapshot())
	})
}
