package import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// TimelineEntry represents a single job execution event for Gantt chart rendering.
type TimelineEntry struct {
	JobID        string    `json:"job_id"`
	TargetURL    string    `json:"target_url"`
	StartTime    time.Time `json:"start_time"`
	DurationMs   int64     `json:"duration_ms"`
	StatusCode   int       `json:"status_code"`
	Outcome      string    `json:"outcome"` // "success" or "failed"
	Depth        int       `json:"depth"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// TimelineRecorder stores a circular buffer of job execution timeline entries.
type TimelineRecorder struct {
	mu       sync.RWMutex
	capacity int
	entries  []TimelineEntry
}

// NewTimelineRecorder creates a circular buffer timeline recorder with given capacity.
func NewTimelineRecorder(capacity int) *TimelineRecorder {
	if capacity <= 0 {
		capacity = 500
	}
	return &TimelineRecorder{
		capacity: capacity,
		entries:  make([]TimelineEntry, 0, capacity),
	}
}

// Record Execution adds an entry to the ring buffer.
func (tr *TimelineRecorder) Record(entry TimelineEntry) {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if len(tr.entries) >= tr.capacity {
		tr.entries = tr.entries[1:] // Drop oldest
	}
	tr.entries = append(tr.entries, entry)
}

// GetEntries returns a snapshot of recorded timeline entries.
func (tr *TimelineRecorder) GetEntries() []TimelineEntry {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	out := make([]TimelineEntry, len(tr.entries))
	copy(out, tr.entries)
	return out
}

// HTTPHandler returns an http.Handler serving Gantt chart timeline metrics.
func (tr *TimelineRecorder) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		entries := tr.GetEntries()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"count":    len(entries),
			"capacity": tr.capacity,
			"timeline": entries,
		})
	})
}
