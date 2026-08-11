package engine

import (
	"context"
	"sync"
	"time"

	"github.com/vyuvaraj/pranor/eval/api"
)

// ossEvalEngine implements EvalEngine for OSS builds.
type ossEvalEngine struct {
	mu         sync.RWMutex
	evaluators []api.Evaluator
}

// NewOSSEngine returns an EvalEngine for OSS (non-enterprise) use.
func NewOSSEngine() api.EvalEngine {
	return &ossEvalEngine{}
}

func (e *ossEvalEngine) Register(ev api.Evaluator) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.evaluators = append(e.evaluators, ev)
}

func (e *ossEvalEngine) Run(ctx context.Context, t api.Trajectory) (api.EvalResult, error) {
	e.mu.RLock()
	evs := make([]api.Evaluator, len(e.evaluators))
	copy(evs, e.evaluators)
	e.mu.RUnlock()

	if len(evs) == 0 {
		return api.EvalResult{}, api.ErrNoEvaluators
	}
	if len(t.Spans) == 0 {
		return api.EvalResult{}, api.ErrInvalidTrajectory
	}

	result := api.EvalResult{
		TrajectoryID: t.ID,
		EvaluatedAt:  time.Now().UTC(),
		OverallPass:  true,
	}

	for _, ev := range evs {
		score, err := ev.Evaluate(ctx, t)
		if err != nil {
			// Soft failure: record a failing score, don't abort the run.
			score = api.EvalScore{
				Evaluator: ev.Name(),
				Score:     0,
				MaxScore:  1.0,
				Pass:      false,
				Detail:    err.Error(),
			}
		}
		result.Scores = append(result.Scores, score)
		if !score.Pass {
			result.OverallPass = false
		}
	}

	return result, nil
}

// Replay re-emits a trajectory's spans in order, returning a new annotated Trajectory.
// OSS: replays locally without archiving. EE: persists to trace archive for Eval query API.
func (e *ossEvalEngine) Replay(ctx context.Context, t api.Trajectory) (api.Trajectory, error) {
	if len(t.Spans) == 0 {
		return api.Trajectory{}, api.ErrInvalidTrajectory
	}
	// OSS replay: return a copy of the trajectory with RecordedAt set to now.
	replayed := api.Trajectory{
		ID:         t.ID + "-replay",
		AgentID:    t.AgentID,
		TenantID:   t.TenantID,
		RequestID:  t.RequestID,
		RecordedAt: time.Now().UTC(),
		Spans:      make([]api.TrajectorySpan, len(t.Spans)),
	}
	copy(replayed.Spans, t.Spans)
	return replayed, nil
}
