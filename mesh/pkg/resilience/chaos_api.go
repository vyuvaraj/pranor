package import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// ChaosRule defines simulated fault parameters for a target service route.
type ChaosRule struct {
	TargetService   string        `json:"target_service"`
	LatencyDelay    time.Duration `json:"latency_delay_ms"`
	LatencyRatio    float64       `json:"latency_ratio"`    // 0.0 to 1.0 (probability)
	ErrorStatus     int           `json:"error_status"`      // e.g., 500, 503
	ErrorRatio      float64       `json:"error_ratio"`      // 0.0 to 1.0
	EnablePartition bool          `json:"enable_partition"` // Simulates complete network partition
}

// ChaosEngine manages chaos engineering fault rules for Pranor Mesh routes.
type ChaosEngine struct {
	mu    sync.RWMutex
	rules map[string]ChaosRule // targetService -> ChaosRule
	rng   *rand.Rand
}

// NewChaosEngine creates a ChaosEngine instance.
func NewChaosEngine() *ChaosEngine {
	return &ChaosEngine{
		rules: make(map[string]ChaosRule),
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// InjectRule registers or updates a chaos rule for a target service.
func (ce *ChaosEngine) InjectRule(rule ChaosRule) {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	ce.rules[rule.TargetService] = rule
}

// RemoveRule deletes any active chaos rule for a target service.
func (ce *ChaosEngine) RemoveRule(targetService string) {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	delete(ce.rules, targetService)
}

// Evaluate applies configured chaos faults (latency delay, error response, or partition drop).
// Returns non-nil http.Response if an artificial fault should short-circuit the request.
func (ce *ChaosEngine) Evaluate(targetService string, req *http.Request) (*http.Response, bool) {
	ce.mu.Lock()
	rule, ok := ce.rules[targetService]
	rndVal1 := ce.rng.Float64()
	rndVal2 := ce.rng.Float64()
	ce.mu.Unlock()

	if !ok {
		return nil, false
	}

	// 1. Simulate Network Partition (100% drop)
	if rule.EnablePartition {
		return &http.Response{
			StatusCode: http.StatusServiceUnavailable,
			Body:       http.NoBody,
			Header:     make(http.Header),
			Request:    req,
		}, true
	}

	// 2. Inject Artificial Latency Delay
	if rule.LatencyDelay > 0 && rndVal1 < rule.LatencyRatio {
		time.Sleep(rule.LatencyDelay * time.Millisecond)
	}

	// 3. Inject Artificial HTTP Error Status
	if rule.ErrorStatus > 0 && rndVal2 < rule.ErrorRatio {
		return &http.Response{
			StatusCode: rule.ErrorStatus,
			Body:       http.NoBody,
			Header:     make(http.Header),
			Request:    req,
		}, true
	}

	return nil, false
}

// HTTPHandler exposes REST endpoints `/api/v1/chaos/inject` and `/api/v1/chaos/rules`.
func (ce *ChaosEngine) HTTPHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/chaos/inject", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var rule ChaosRule
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
			return
		}
		if rule.TargetService == "" {
			http.Error(w, "target_service is required", http.StatusBadRequest)
			return
		}
		ce.InjectRule(rule)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	mux.HandleFunc("/api/v1/chaos/rules", func(w http.ResponseWriter, r *http.Request) {
		ce.mu.RLock()
		rules := ce.rules
		ce.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rules)
	})

	return mux
}
