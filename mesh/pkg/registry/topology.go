package import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// TopologyEdge represents a directional traffic dependency link between two services.
type TopologyEdge struct {
	SourceService string        `json:"source_service"`
	TargetService string        `json:"target_service"`
	RPS           float64       `json:"rps"`
	P99LatencyMs  float64       `json:"p99_latency_ms"`
	ErrorRate     float64       `json:"error_rate"`
	BreakerState  string        `json:"breaker_state"` // "closed", "open", "half_open"
	LastSeen      time.Time     `json:"last_seen"`
}

// TopologyGraph maintains live inter-service dependency links for Pranor Console force-directed graph rendering.
type TopologyGraph struct {
	mu    sync.RWMutex
	edges map[string]*TopologyEdge // "source->target" -> edge
}

// NewTopologyGraph creates a TopologyGraph instance.
func NewTopologyGraph() *TopologyGraph {
	return &TopologyGraph{
		edges: make(map[string]*TopologyEdge),
	}
}

// UpdateEdge records or updates traffic metrics between a source and target service.
func (tg *TopologyGraph) UpdateEdge(source, target string, latencyMs float64, isError bool, breakerState string) {
	if source == "" || target == "" {
		return
	}
	tg.mu.Lock()
	defer tg.mu.Unlock()

	key := source + "->" + target
	edge, ok := tg.edges[key]
	if !ok {
		edge = &TopologyEdge{
			SourceService: source,
			TargetService: target,
			BreakerState:  "closed",
		}
		tg.edges[key] = edge
	}

	edge.LastSeen = time.Now()

	// EWMA latency update
	if edge.P99LatencyMs == 0 {
		edge.P99LatencyMs = latencyMs
	} else {
		edge.P99LatencyMs = 0.8*edge.P99LatencyMs + 0.2*latencyMs
	}

	// EWMA error rate update
	errVal := 0.0
	if isError {
		errVal = 1.0
	}
	edge.ErrorRate = 0.9*edge.ErrorRate + 0.1*errVal

	if breakerState != "" {
		edge.BreakerState = breakerState
	}
}

// GetEdges returns a snapshot of active inter-service edges (seen within last 5 minutes).
func (tg *TopologyGraph) GetEdges() []TopologyEdge {
	tg.mu.RLock()
	defer tg.mu.RUnlock()

	cutoff := time.Now().Add(-5 * time.Minute)
	list := make([]TopologyEdge, 0)
	for _, edge := range tg.edges {
		if edge.LastSeen.After(cutoff) {
			list = append(list, *edge)
		}
	}
	return list
}

// HTTPHandler exposes REST endpoint `/api/v1/topology` for Pranor Console visual topology rendering.
func (tg *TopologyGraph) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		edges := tg.GetEdges()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"count": len(edges),
			"edges": edges,
		})
	})
}
