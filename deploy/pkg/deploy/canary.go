package import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CanaryState tracks canary rollout progression.
type CanaryState string

const (
	CanaryStatePending    CanaryState = "pending"
	CanaryStateInProgress CanaryState = "in_progress"
	CanaryStatePromoted   CanaryState = "promoted"
	CanaryStateRolledBack CanaryState = "rolled_back"
)

// CanaryConfig defines traffic weights and error rate threshold for canary rollouts.
type CanaryConfig struct {
	ServiceName           string    `json:"service_name"`
	Steps                 []int     `json:"steps"`                  // e.g. [5, 25, 50, 100]
	MaxErrorRateThreshold float64   `json:"max_error_rate_threshold"` // e.g. 0.05 (5%)
	StepInterval          time.Duration `json:"step_interval"`
}

// CanaryManager handles automated canary traffic stepping and instant automatic rollback on SLO violations.
type CanaryManager struct {
	mu            sync.RWMutex
	cfg           CanaryConfig
	state         CanaryState
	currentWeight int
}

// NewCanaryManager creates a CanaryManager instance.
func NewCanaryManager(cfg CanaryConfig) *CanaryManager {
	if len(cfg.Steps) == 0 {
		cfg.Steps = []int{5, 25, 50, 100}
	}
	if cfg.MaxErrorRateThreshold <= 0 {
		cfg.MaxErrorRateThreshold = 0.05
	}
	if cfg.StepInterval <= 0 {
		cfg.StepInterval = 100 * time.Millisecond
	}
	return &CanaryManager{
		cfg:           cfg,
		state:         CanaryStatePending,
		currentWeight: 0,
	}
}

// Status returns current canary rollout state and traffic weight.
func (cm *CanaryManager) Status() (CanaryState, int) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.state, cm.currentWeight
}

// ExecuteRollout steps through canary traffic weights, monitoring error rates continuously.
func (cm *CanaryManager) ExecuteRollout(ctx context.Context, errorRateFetcher func() float64) (CanaryState, error) {
	cm.mu.Lock()
	cm.state = CanaryStateInProgress
	cm.mu.Unlock()

	for _, weight := range cm.cfg.Steps {
		select {
		case <-ctx.Done():
			cm.rollback()
			return CanaryStateRolledBack, ctx.Err()
		default:
		}

		cm.mu.Lock()
		cm.currentWeight = weight
		cm.mu.Unlock()

		time.Sleep(cm.cfg.StepInterval)

		// Evaluate error rate SLO threshold
		currentErrorRate := errorRateFetcher()
		if currentErrorRate > cm.cfg.MaxErrorRateThreshold {
			cm.rollback()
			return CanaryStateRolledBack, fmt.Errorf("canary error rate %.2f exceeds threshold %.2f — auto-rolled back", currentErrorRate, cm.cfg.MaxErrorRateThreshold)
		}
	}

	cm.mu.Lock()
	cm.state = CanaryStatePromoted
	cm.currentWeight = 100
	cm.mu.Unlock()

	return CanaryStatePromoted, nil
}

func (cm *CanaryManager) rollback() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.state = CanaryStateRolledBack
	cm.currentWeight = 0
}
