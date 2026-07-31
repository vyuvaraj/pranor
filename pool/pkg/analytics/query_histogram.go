package analytics

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// SlowQueryLog records query string, duration, and timestamp.
type SlowQueryLog struct {
	Query      string        `json:"query"`
	DurationMs float64       `json:"duration_ms"`
	Timestamp  time.Time     `json:"timestamp"`
}

// QueryHistogram tracks database query latency distributions and captures slow queries.
type QueryHistogram struct {
	mu             sync.RWMutex
	buckets        map[string]int64 // e.g. "<5ms", "<50ms", "<200ms", ">200ms"
	slowQueries    []SlowQueryLog
	slowThreshold  time.Duration
	maxSlowRecords int
}

// NewQueryHistogram creates a QueryHistogram instance.
func NewQueryHistogram(slowThreshold time.Duration) *QueryHistogram {
	if slowThreshold <= 0 {
		slowThreshold = 200 * time.Millisecond
	}
	return &QueryHistogram{
		buckets: map[string]int64{
			"le_5ms":   0,
			"le_50ms":  0,
			"le_200ms": 0,
			"gt_200ms": 0,
		},
		slowQueries:    make([]SlowQueryLog, 0),
		slowThreshold:  slowThreshold,
		maxSlowRecords: 100,
	}
}

// RecordQuery records query execution duration and logs if slow threshold is breached.
func (qh *QueryHistogram) RecordQuery(query string, duration time.Duration) {
	qh.mu.Lock()
	defer qh.mu.Unlock()

	durMs := float64(duration.Microseconds()) / 1000.0

	if duration <= 5*time.Millisecond {
		qh.buckets["le_5ms"]++
	} else if duration <= 50*time.Millisecond {
		qh.buckets["le_50ms"]++
	} else if duration <= 200*time.Millisecond {
		qh.buckets["le_200ms"]++
	} else {
		qh.buckets["gt_200ms"]++
	}

	if duration >= qh.slowThreshold {
		log := SlowQueryLog{
			Query:      query,
			DurationMs: durMs,
			Timestamp:  time.Now(),
		}
		if len(qh.slowQueries) >= qh.maxSlowRecords {
			qh.slowQueries = qh.slowQueries[1:]
		}
		qh.slowQueries = append(qh.slowQueries, log)
	}
}

// GetSnapshot returns current histogram counts and slow query log records.
func (qh *QueryHistogram) GetSnapshot() (map[string]int64, []SlowQueryLog) {
	qh.mu.RLock()
	defer qh.mu.RUnlock()

	bCopy := make(map[string]int64, len(qh.buckets))
	for k, v := range qh.buckets {
		bCopy[k] = v
	}
	sqCopy := make([]SlowQueryLog, len(qh.slowQueries))
	copy(sqCopy, qh.slowQueries)

	return bCopy, sqCopy
}

// HTTPHandler exposes `/api/v1/pool/query-stats` for Pranor Console visual display.
func (qh *QueryHistogram) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		buckets, slowLogs := qh.GetSnapshot()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"buckets":      buckets,
			"slow_queries": slowLogs,
		})
	})
}
