package eval

import (
	"context"
	"testing"
	"time"

	"github.com/vyuvaraj/pranor/eval/api"
	"github.com/vyuvaraj/pranor/eval/pkg/engine"
	"github.com/vyuvaraj/pranor/eval/pkg/evaluators"
)

func makeTrajectory(spans []api.TrajectorySpan) api.Trajectory {
	return api.Trajectory{
		ID:         "traj-001",
		AgentID:    "agent-1",
		TenantID:   "tenant-1",
		RequestID:  "req-1",
		RecordedAt: time.Now().UTC(),
		Spans:      spans,
	}
}

func TestEvalEngine_NoEvaluators(t *testing.T) {
	eng := engine.NewOSSEngine()
	traj := makeTrajectory([]api.TrajectorySpan{
		{SpanName: "pranor.agent_execution", Module: "gate", Outcome: "ALLOW", DurationMs: 10},
	})
	_, err := eng.Run(context.Background(), traj)
	if err != api.ErrNoEvaluators {
		t.Fatalf("expected ErrNoEvaluators, got %v", err)
	}
}

func TestEvalEngine_EmptyTrajectory(t *testing.T) {
	eng := engine.NewOSSEngine()
	eng.Register(&evaluators.AccuracyEvaluator{})
	_, err := eng.Run(context.Background(), api.Trajectory{ID: "empty"})
	if err != api.ErrInvalidTrajectory {
		t.Fatalf("expected ErrInvalidTrajectory, got %v", err)
	}
}

func TestAccuracyEvaluator_AllPass(t *testing.T) {
	eng := engine.NewOSSEngine()
	eng.Register(&evaluators.AccuracyEvaluator{})
	traj := makeTrajectory([]api.TrajectorySpan{
		{Module: "gate", Outcome: "ALLOW", DurationMs: 5},
		{Module: "decision", Outcome: "APPROVE", DurationMs: 3},
	})
	result, err := eng.Run(context.Background(), traj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.OverallPass {
		t.Error("expected overall pass")
	}
	if result.Scores[0].Evaluator != "accuracy" {
		t.Errorf("expected accuracy evaluator, got %s", result.Scores[0].Evaluator)
	}
}

func TestAccuracyEvaluator_WithErrors(t *testing.T) {
	eng := engine.NewOSSEngine()
	eng.Register(&evaluators.AccuracyEvaluator{})
	traj := makeTrajectory([]api.TrajectorySpan{
		{Module: "gate", Outcome: "ERROR", Error: api.ErrInvalidTrajectory},
		{Module: "gate", Outcome: "ERROR", Error: api.ErrInvalidTrajectory},
		{Module: "decision", Outcome: "APPROVE", DurationMs: 3},
	})
	result, err := eng.Run(context.Background(), traj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Scores[0].Pass { // 2/3 errors = score ~0.33, below 0.8 threshold
		t.Error("expected accuracy to fail with high error rate")
	}
}

func TestLatencyEvaluator_WithinBudget(t *testing.T) {
	eng := engine.NewOSSEngine()
	eng.Register(&evaluators.LatencyEvaluator{BudgetMs: 100})
	traj := makeTrajectory([]api.TrajectorySpan{
		{Module: "gate", Outcome: "ALLOW", DurationMs: 20},
		{Module: "decision", Outcome: "APPROVE", DurationMs: 30},
	})
	result, err := eng.Run(context.Background(), traj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Scores[0].Pass {
		t.Errorf("expected latency pass (50ms < 100ms budget), detail: %s", result.Scores[0].Detail)
	}
}

func TestLatencyEvaluator_ExceedsBudget(t *testing.T) {
	eng := engine.NewOSSEngine()
	eng.Register(&evaluators.LatencyEvaluator{BudgetMs: 10})
	traj := makeTrajectory([]api.TrajectorySpan{
		{Module: "gate", Outcome: "ALLOW", DurationMs: 50},
	})
	result, err := eng.Run(context.Background(), traj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Scores[0].Pass {
		t.Error("expected latency to fail (50ms > 10ms budget)")
	}
}

func TestCostEvaluator(t *testing.T) {
	eng := engine.NewOSSEngine()
	eng.Register(&evaluators.CostEvaluator{MaxSpans: 5})
	traj := makeTrajectory([]api.TrajectorySpan{
		{Module: "gate", Outcome: "ALLOW"},
		{Module: "graph", Outcome: "ALLOW"},
		{Module: "decision", Outcome: "APPROVE"},
	})
	result, err := eng.Run(context.Background(), traj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Scores[0].Pass {
		t.Error("expected cost pass (3 spans < max 5)")
	}
}

func TestSafetyEvaluator_NoViolations(t *testing.T) {
	eng := engine.NewOSSEngine()
	eng.Register(&evaluators.SafetyEvaluator{CriticalModules: []string{"gate", "decision"}})
	traj := makeTrajectory([]api.TrajectorySpan{
		{Module: "gate", Outcome: "ALLOW"},
		{Module: "decision", Outcome: "APPROVE"},
	})
	result, err := eng.Run(context.Background(), traj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Scores[0].Pass {
		t.Error("expected safety pass with no violations")
	}
}

func TestReplay(t *testing.T) {
	eng := engine.NewOSSEngine()
	traj := makeTrajectory([]api.TrajectorySpan{
		{SpanName: "pranor.agent_execution", Module: "gate", Outcome: "ALLOW", DurationMs: 10},
	})
	replayed, err := eng.Replay(context.Background(), traj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if replayed.ID != traj.ID+"-replay" {
		t.Errorf("expected replay ID suffix, got %s", replayed.ID)
	}
	if len(replayed.Spans) != len(traj.Spans) {
		t.Errorf("expected same number of spans in replay")
	}
}

func TestSentinelErrors(t *testing.T) {
	if api.ErrNoEvaluators == nil {
		t.Error("ErrNoEvaluators must be non-nil")
	}
	if api.ErrInvalidTrajectory == nil {
		t.Error("ErrInvalidTrajectory must be non-nil")
	}
	if api.ErrEERequired == nil {
		t.Error("ErrEERequired must be non-nil")
	}
}
