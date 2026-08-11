package agentstep

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type SagaResult struct {
	CompletedSteps   int
	CompensatedSteps int
	FinalError       error
	TotalLatencyMs   int64
}

type Saga struct {
	config    SagaConfig
	steps     []AgentStep
	completed []StepInput
}

func NewSaga(cfg SagaConfig, steps []AgentStep) *Saga {
	return &Saga{
		config:    cfg,
		steps:     steps,
		completed: make([]StepInput, 0),
	}
}

func (s *Saga) Run(ctx context.Context) (SagaResult, error) {
	start := time.Now()
	res := SagaResult{}

	totalCtx, totalCancel := context.WithTimeout(ctx, s.config.TotalTimeout)
	defer totalCancel()

	for i, step := range s.steps {
		if i >= s.config.MaxSteps {
			if s.config.OnStepLimitHit == LimitPolicyCompensate {
				res.CompensatedSteps = s.compensateAll(ctx)
				res.FinalError = ErrSagaStepLimitExceeded
				res.TotalLatencyMs = time.Since(start).Milliseconds()
				return res, ErrSagaStepLimitExceeded
			} else if s.config.OnStepLimitHit == LimitPolicyPauseForHITL {
				err := pauseForHITL(ctx, s)
				if err != nil {
					res.CompensatedSteps = s.compensateAll(ctx)
					res.FinalError = err
					res.TotalLatencyMs = time.Since(start).Milliseconds()
					return res, err
				}
				// Normally we'd wait, but for this implementation we just return or proceed based on logic.
				// Since we're beyond max steps, let's just error out for now after pause if it failed, or return limit exceeded.
				res.FinalError = ErrSagaStepLimitExceeded
				res.TotalLatencyMs = time.Since(start).Milliseconds()
				return res, ErrSagaStepLimitExceeded
			}
		}

		select {
		case <-totalCtx.Done():
			res.CompensatedSteps = s.compensateAll(ctx)
			res.FinalError = ErrSagaTimeout
			res.TotalLatencyMs = time.Since(start).Milliseconds()
			return res, ErrSagaTimeout
		default:
		}

		stepCtx, stepCancel := context.WithTimeout(totalCtx, s.config.StepTimeout)
		
		input := StepInput{
			StepName:   step.Name(),
			AttemptNum: 1,
		}

		_, err := step.Execute(stepCtx, input)
		stepCancel()

		if err != nil {
			errToReturn := err
			if errors.Is(stepCtx.Err(), context.DeadlineExceeded) {
				errToReturn = ErrStepTimeout
			}
			res.CompensatedSteps = s.compensateAll(ctx)
			res.FinalError = errToReturn
			res.TotalLatencyMs = time.Since(start).Milliseconds()
			return res, errToReturn
		}

		s.completed = append(s.completed, input)
		res.CompletedSteps++
	}

	res.TotalLatencyMs = time.Since(start).Milliseconds()
	return res, nil
}

func (s *Saga) compensateAll(ctx context.Context) int {
	count := 0
	for i := len(s.completed) - 1; i >= 0; i-- {
		input := s.completed[i]
		// Find the step corresponding to this input
		var stepToCompensate AgentStep
		for _, step := range s.steps {
			if step.Name() == input.StepName {
				stepToCompensate = step
				break
			}
		}
		if stepToCompensate != nil {
			err := stepToCompensate.Compensate(ctx, input)
			if err != nil {
				// Log critical, do not swallow. For now, fmt.Println.
				fmt.Printf("CRITICAL: compensation failed for step %s: %v\n", input.StepName, err)
			} else {
				count++
			}
		}
	}
	return count
}
