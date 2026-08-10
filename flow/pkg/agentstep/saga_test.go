package agentstep

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockStep struct {
	name          string
	executeErr    error
	compensateErr error
	sleepDur      time.Duration
}

func (m *mockStep) Name() string { return m.name }

func (m *mockStep) Execute(ctx context.Context, input StepInput) (StepOutput, error) {
	if m.sleepDur > 0 {
		select {
		case <-ctx.Done():
			return StepOutput{}, ctx.Err()
		case <-time.After(m.sleepDur):
		}
	}
	return StepOutput{StepName: m.name}, m.executeErr
}

func (m *mockStep) Compensate(ctx context.Context, input StepInput) error {
	return m.compensateErr
}

func TestDefaultSagaConfig(t *testing.T) {
	cfg := DefaultSagaConfig()
	if cfg.MaxSteps != 25 {
		t.Errorf("expected MaxSteps=25, got %d", cfg.MaxSteps)
	}
	if cfg.StepTimeout != 30*time.Second {
		t.Errorf("expected StepTimeout=30s, got %v", cfg.StepTimeout)
	}
	if cfg.TotalTimeout != 10*time.Minute {
		t.Errorf("expected TotalTimeout=10m, got %v", cfg.TotalTimeout)
	}
}

func TestSagaRunSuccessful(t *testing.T) {
	cfg := DefaultSagaConfig()
	steps := []AgentStep{
		&mockStep{name: "step1"},
		&mockStep{name: "step2"},
	}
	saga := NewSaga(cfg, steps)
	res, err := saga.Run(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if res.CompletedSteps != len(steps) {
		t.Errorf("expected %d completed steps, got %d", len(steps), res.CompletedSteps)
	}
}

func TestSagaRunWithStepFailure(t *testing.T) {
	cfg := DefaultSagaConfig()
	expectedErr := errors.New("step2 failed")
	steps := []AgentStep{
		&mockStep{name: "step1"},
		&mockStep{name: "step2", executeErr: expectedErr},
		&mockStep{name: "step3"},
	}
	saga := NewSaga(cfg, steps)
	res, err := saga.Run(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if res.CompletedSteps != 1 {
		t.Errorf("expected 1 completed step, got %d", res.CompletedSteps)
	}
	if res.CompensatedSteps != 1 {
		t.Errorf("expected 1 compensated step, got %d", res.CompensatedSteps)
	}
}

func TestSagaStepLimitExceeded(t *testing.T) {
	cfg := DefaultSagaConfig()
	cfg.MaxSteps = 2
	steps := []AgentStep{
		&mockStep{name: "step1"},
		&mockStep{name: "step2"},
		&mockStep{name: "step3"},
	}
	saga := NewSaga(cfg, steps)
	res, err := saga.Run(context.Background())
	if !errors.Is(err, ErrSagaStepLimitExceeded) {
		t.Fatalf("expected ErrSagaStepLimitExceeded, got %v", err)
	}
	if res.CompletedSteps != 2 {
		t.Errorf("expected 2 completed steps, got %d", res.CompletedSteps)
	}
	if res.CompensatedSteps != 2 {
		t.Errorf("expected 2 compensated steps, got %d", res.CompensatedSteps)
	}
}

func TestSagaStepTimeout(t *testing.T) {
	cfg := DefaultSagaConfig()
	cfg.StepTimeout = 10 * time.Millisecond
	steps := []AgentStep{
		&mockStep{name: "step1"},
		&mockStep{name: "step2", sleepDur: 50 * time.Millisecond},
	}
	saga := NewSaga(cfg, steps)
	res, err := saga.Run(context.Background())
	if !errors.Is(err, ErrStepTimeout) {
		t.Fatalf("expected ErrStepTimeout, got %v", err)
	}
	if res.CompletedSteps != 1 {
		t.Errorf("expected 1 completed steps, got %d", res.CompletedSteps)
	}
	if res.CompensatedSteps != 1 {
		t.Errorf("expected 1 compensated steps, got %d", res.CompensatedSteps)
	}
}

func TestSentinelErrors(t *testing.T) {
	if ErrSagaStepLimitExceeded == nil ||
		ErrSagaTimeout == nil ||
		ErrStepTimeout == nil ||
		ErrCompensationFailed == nil ||
		ErrEERequired == nil {
		t.Fatalf("expected sentinel errors to be non-nil")
	}
}
