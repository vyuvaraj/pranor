package import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// FaultKind specifies the type of chaos fault to inject.
type FaultKind string

const (
	FaultNetwork  FaultKind = "network"
	FaultCPU      FaultKind = "cpu"
	FaultMemory   FaultKind = "memory"
	FaultDisk     FaultKind = "disk"
	FaultClockSkew FaultKind = "clock_skew"
)

// ChaosFault defines a chaos injection specification.
type ChaosFault struct {
	ID          string        `json:"id"`
	Kind        FaultKind     `json:"kind"`
	TargetNode  string        `json:"target_node"`
	Intensity   float64       `json:"intensity"` // 0.0-1.0 e.g. 0.5 = 50% packet loss or 50% CPU spike
	Duration    time.Duration `json:"duration"`
	Active      bool          `json:"active"`
	InjectedAt  time.Time     `json:"injected_at"`
}

// UnifiedChaosEngine manages cross-platform fault injection across network, CPU, memory, disk, and clock.
type UnifiedChaosEngine struct {
	mu     sync.RWMutex
	faults map[string]*ChaosFault
}

// NewUnifiedChaosEngine creates a UnifiedChaosEngine instance.
func NewUnifiedChaosEngine() *UnifiedChaosEngine {
	return &UnifiedChaosEngine{
		faults: make(map[string]*ChaosFault),
	}
}

// InjectFault registers and activates a chaos fault.
func (uce *UnifiedChaosEngine) InjectFault(fault ChaosFault) (*ChaosFault, error) {
	if fault.TargetNode == "" {
		return nil, fmt.Errorf("target_node is required")
	}
	if fault.Kind == "" {
		return nil, fmt.Errorf("fault kind is required")
	}
	if fault.Intensity < 0 || fault.Intensity > 1 {
		return nil, fmt.Errorf("intensity must be between 0.0 and 1.0")
	}

	uce.mu.Lock()
	defer uce.mu.Unlock()

	id := fmt.Sprintf("fault-%d", len(uce.faults)+1)
	fault.ID = id
	fault.Active = true
	fault.InjectedAt = time.Now()
	uce.faults[id] = &fault
	return &fault, nil
}

// AbortFault deactivates an active chaos fault.
func (uce *UnifiedChaosEngine) AbortFault(faultID string) error {
	uce.mu.Lock()
	defer uce.mu.Unlock()
	f, ok := uce.faults[faultID]
	if !ok {
		return fmt.Errorf("fault '%s' not found", faultID)
	}
	f.Active = false
	return nil
}

// ListFaults returns all registered chaos faults.
func (uce *UnifiedChaosEngine) ListFaults() []ChaosFault {
	uce.mu.RLock()
	defer uce.mu.RUnlock()
	res := make([]ChaosFault, 0, len(uce.faults))
	for _, f := range uce.faults {
		res = append(res, *f)
	}
	return res
}

// HTTPHandler exposes `/api/v1/platform/chaos/faults`.
func (uce *UnifiedChaosEngine) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			var fault ChaosFault
			if err := json.NewDecoder(r.Body).Decode(&fault); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			injected, err := uce.InjectFault(fault)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(injected)
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"faults": uce.ListFaults()})
	})
}
