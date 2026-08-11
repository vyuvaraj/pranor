package api

import (
	"context"
	"time"
)

// TrajectorySpan represents a single recorded span in an agent execution trajectory.
type TrajectorySpan struct {
	SpanName    string            // e.g. "pranor.agent_execution", "pranor.gate.inspect"
	Module      string            // e.g. "gate", "graph", "decision"
	Outcome     string            // "ALLOW", "DENY", "APPROVE", "TRANSFORM", "ERROR"
	AgentID     string
	UserID      string
	TenantID    string
	RequestID   string
	StartedAt   time.Time
	DurationMs  int64
	Attrs       map[string]string
	Error       error
}

// Trajectory is a complete ordered sequence of spans for one agent execution.
type Trajectory struct {
	ID        string
	AgentID   string
	TenantID  string
	RequestID string
	RecordedAt time.Time
	Spans     []TrajectorySpan
}

// EvalScore holds the result of one evaluator pass over a trajectory.
type EvalScore struct {
	Evaluator   string  // "accuracy", "latency", "cost", "safety"
	Score       float64 // 0.0 – 1.0
	MaxScore    float64
	Pass        bool
	Detail      string
}

// EvalResult holds all scores for a single trajectory evaluation.
type EvalResult struct {
	TrajectoryID string
	Scores       []EvalScore
	OverallPass  bool
	EvaluatedAt  time.Time
}

// Evaluator is the interface each score evaluator must implement.
type Evaluator interface {
	Name() string
	Evaluate(ctx context.Context, t Trajectory) (EvalScore, error)
}

// EvalEngine runs all registered evaluators over a trajectory.
type EvalEngine interface {
	Register(e Evaluator)
	Run(ctx context.Context, t Trajectory) (EvalResult, error)
	Replay(ctx context.Context, t Trajectory) (Trajectory, error)
}

// Sentinel errors
var (
	ErrNoEvaluators    = errorf("pranor/eval: no evaluators registered")
	ErrInvalidTrajectory = errorf("pranor/eval: invalid or empty trajectory")
	ErrEERequired      = errorf("pranor/eval: this capability requires Pranor Enterprise Edition")
)

type evalError string

func (e evalError) Error() string { return string(e) }
func errorf(s string) error       { return evalError(s) }
