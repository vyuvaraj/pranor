package import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// ComponentHealth represents live health status of a Pranor platform component.
type ComponentHealth struct {
	Name      string        `json:"name"`
	Healthy   bool          `json:"healthy"`
	Latency   time.Duration `json:"latency_ns"`
	Message   string        `json:"message,omitempty"`
	CheckedAt time.Time     `json:"checked_at"`
}

// HealthRollup aggregates health across all components.
type HealthRollup struct {
	Healthy    bool              `json:"healthy"`
	Components []ComponentHealth `json:"components"`
	Total      int               `json:"total"`
	Passing    int               `json:"passing"`
	Failing    int               `json:"failing"`
}

// UnifiedHealthAPI provides platform-wide health, readiness, and rollup metrics endpoints.
type UnifiedHealthAPI struct {
	mu      sync.RWMutex
	checks  map[string]*ComponentHealth
}

// NewUnifiedHealthAPI creates a UnifiedHealthAPI instance.
func NewUnifiedHealthAPI() *UnifiedHealthAPI {
	return &UnifiedHealthAPI{
		checks: make(map[string]*ComponentHealth),
	}
}

// ReportHealth records a component health check result.
func (uha *UnifiedHealthAPI) ReportHealth(h ComponentHealth) {
	uha.mu.Lock()
	defer uha.mu.Unlock()
	if h.CheckedAt.IsZero() {
		h.CheckedAt = time.Now()
	}
	uha.checks[h.Name] = &h
}

// GetRollup returns the aggregated health rollup across all components.
func (uha *UnifiedHealthAPI) GetRollup() HealthRollup {
	uha.mu.RLock()
	defer uha.mu.RUnlock()

	var components []ComponentHealth
	passing, failing := 0, 0
	for _, h := range uha.checks {
		components = append(components, *h)
		if h.Healthy {
			passing++
		} else {
			failing++
		}
	}

	return HealthRollup{
		Healthy:    failing == 0,
		Components: components,
		Total:      len(components),
		Passing:    passing,
		Failing:    failing,
	}
}

// HTTPHandler exposes `/health`, `/ready`, and `/api/v1/platform/health/rollup`.
func (uha *UnifiedHealthAPI) HTTPHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		rollup := uha.GetRollup()
		w.Header().Set("Content-Type", "application/json")
		if rollup.Healthy {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": rollup.Healthy})
	})

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		rollup := uha.GetRollup()
		w.Header().Set("Content-Type", "application/json")
		if rollup.Total > 0 && rollup.Healthy {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]bool{"ready": true})
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]bool{"ready": false})
		}
	})

	mux.HandleFunc("/api/v1/platform/health/rollup", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(uha.GetRollup())
	})

	return mux
}
