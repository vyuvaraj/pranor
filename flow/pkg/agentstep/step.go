package agentstep

import (
	"context"
	"errors"
	"time"
)

type StepInput struct {
	StepName   string
	Payload    map[string]any
	RequestID  string
	AttemptNum int
}

type StepOutput struct {
	StepName string
	Result   map[string]any
	LatencyMs int64
}

type AgentStep interface {
	Execute(ctx context.Context, input StepInput) (StepOutput, error)
	Compensate(ctx context.Context, input StepInput) error
	Name() string
}

type LimitPolicy int

const (
	LimitPolicyCompensate LimitPolicy = iota
	LimitPolicyPauseForHITL
)

type SagaConfig struct {
	MaxSteps       int
	StepTimeout    time.Duration
	TotalTimeout   time.Duration
	OnStepLimitHit LimitPolicy
}

func DefaultSagaConfig() SagaConfig {
	return SagaConfig{
		MaxSteps:       25,
		StepTimeout:    30 * time.Second,
		TotalTimeout:   10 * time.Minute,
		OnStepLimitHit: LimitPolicyCompensate,
	}
}

var (
	ErrSagaStepLimitExceeded = errors.New("pranor/flow: saga step limit exceeded")
	ErrSagaTimeout           = errors.New("pranor/flow: saga total timeout exceeded")
	ErrStepTimeout           = errors.New("pranor/flow: step execution timeout")
	ErrCompensationFailed    = errors.New("pranor/flow: step compensation failed")
	ErrEERequired            = errors.New("pranor/flow: this capability requires Pranor Enterprise Edition")
)
