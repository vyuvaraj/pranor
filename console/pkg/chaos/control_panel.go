package chaos

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// ChaosExperiment defines a chaos injection trial run.
type ChaosExperiment struct {
	ID            string    `json:"id"`
	TargetService string    `json:"target_service"`
	FaultType     string    `json:"fault_type"` // "latency", "error_500", "partition"
	Duration      time.Duration `json:"duration"`
	Status        string    `json:"status"` // "scheduled", "running", "completed", "aborted"
	CreatedAt     time.Time `json:"created_at"`
}

// ChaosControlPanel orchestrates runtime chaos engineering experiment triggers in Pranor Console.
type ChaosControlPanel struct {
	mu          sync.RWMutex
	experiments map[string]*ChaosExperiment // experimentID -> experiment
}

// NewChaosControlPanel creates a ChaosControlPanel instance.
func NewChaosControlPanel() *ChaosControlPanel {
	return &ChaosControlPanel{
		experiments: make(map[string]*ChaosExperiment),
	}
}

// TriggerExperiment schedules and starts a new chaos experiment.
func (ccp *ChaosControlPanel) TriggerExperiment(exp ChaosExperiment) (*ChaosExperiment, error) {
	if exp.TargetService == "" || exp.FaultType == "" {
		return nil, fmt.Errorf("target_service and fault_type are required")
	}

	ccp.mu.Lock()
	defer ccp.mu.Unlock()

	expID := fmt.Sprintf("exp-%d", len(ccp.experiments)+1)
	exp.ID = expID
	exp.Status = "running"
	exp.CreatedAt = time.Now()

	ccp.experiments[expID] = &exp
	return &exp, nil
}

// AbortExperiment cancels an active chaos experiment immediately.
func (ccp *ChaosControlPanel) AbortExperiment(expID string) error {
	ccp.mu.Lock()
	defer ccp.mu.Unlock()

	exp, ok := ccp.experiments[expID]
	if !ok {
		return fmt.Errorf("experiment ID '%s' not found", expID)
	}

	exp.Status = "aborted"
	return nil
}

// GetExperiments returns active and historical chaos experiments.
func (ccp *ChaosControlPanel) GetExperiments() []ChaosExperiment {
	ccp.mu.RLock()
	defer ccp.mu.RUnlock()

	res := make([]ChaosExperiment, 0, len(ccp.experiments))
	for _, exp := range ccp.experiments {
		res = append(res, *exp)
	}
	return res
}

// HTTPHandler exposes `/api/v1/console/chaos/experiments` for Pranor Console visual controls.
func (ccp *ChaosControlPanel) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodPost {
			var exp ChaosExperiment
			if err := json.NewDecoder(r.Body).Decode(&exp); err != nil {
				http.Error(w, "invalid JSON payload", http.StatusBadRequest)
				return
			}
			created, err := ccp.TriggerExperiment(exp)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(created)
			return
		}

		w.WriteHeader(http.StatusOK)
		experiments := ccp.GetExperiments()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"count":       len(experiments),
			"experiments": experiments,
		})
	})
}
