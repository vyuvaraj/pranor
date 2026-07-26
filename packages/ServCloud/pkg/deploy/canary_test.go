package deploy

import (
	"context"
	"testing"
	"time"
)

func TestCanaryManager_SuccessfulPromotion(t *testing.T) {
	cfg := CanaryConfig{
		ServiceName:           "api-gateway",
		Steps:                 []int{10, 50, 100},
		MaxErrorRateThreshold: 0.05,
		StepInterval:          10 * time.Millisecond,
	}

	cm := NewCanaryManager(cfg)

	// Low error rate (1%) -> should promote smoothly to 100%
	mockErrorFetcher := func() float64 {
		return 0.01
	}

	state, err := cm.ExecuteRollout(context.Background(), mockErrorFetcher)
	if err != nil {
		t.Fatalf("ExecuteRollout failed: %v", err)
	}

	if state != CanaryStatePromoted {
		t.Errorf("expected CanaryStatePromoted, got %s", state)
	}

	finalState, weight := cm.Status()
	if finalState != CanaryStatePromoted || weight != 100 {
		t.Errorf("unexpected status: state=%s weight=%d", finalState, weight)
	}
}

func TestCanaryManager_AutomaticRollback(t *testing.T) {
	cfg := CanaryConfig{
		ServiceName:           "cart-service",
		Steps:                 []int{10, 50, 100},
		MaxErrorRateThreshold: 0.05,
		StepInterval:          10 * time.Millisecond,
	}

	cm := NewCanaryManager(cfg)

	// Error rate spikes to 12% on step 2
	stepCount := 0
	mockSpikingErrorFetcher := func() float64 {
		stepCount++
		if stepCount >= 2 {
			return 0.12 // Exceeds 5% threshold
		}
		return 0.01
	}

	state, err := cm.ExecuteRollout(context.Background(), mockSpikingErrorFetcher)
	if err == nil {
		t.Error("expected error triggering automatic rollback")
	}

	if state != CanaryStateRolledBack {
		t.Errorf("expected CanaryStateRolledBack, got %s", state)
	}

	finalState, weight := cm.Status()
	if finalState != CanaryStateRolledBack || weight != 0 {
		t.Errorf("expected weight 0 after rollback, got weight=%d", weight)
	}
}
