package evaluators

import (
	"context"
	"fmt"

	"github.com/vyuvaraj/pranor/eval/api"
)

// AccuracyEvaluator checks that all spans ended with a non-error outcome.
type AccuracyEvaluator struct{}

func (e *AccuracyEvaluator) Name() string { return "accuracy" }

func (e *AccuracyEvaluator) Evaluate(_ context.Context, t api.Trajectory) (api.EvalScore, error) {
	total := len(t.Spans)
	if total == 0 {
		return api.EvalScore{Evaluator: e.Name()}, api.ErrInvalidTrajectory
	}
	errors := 0
	for _, s := range t.Spans {
		if s.Outcome == "ERROR" || s.Error != nil {
			errors++
		}
	}
	score := 1.0 - float64(errors)/float64(total)
	return api.EvalScore{
		Evaluator: e.Name(),
		Score:     score,
		MaxScore:  1.0,
		Pass:      score >= 0.8,
		Detail:    fmt.Sprintf("%d/%d spans without error", total-errors, total),
	}, nil
}

// LatencyEvaluator checks that the sum of span durations is within a budget.
type LatencyEvaluator struct {
	BudgetMs int64 // fail if total latency exceeds this
}

func (e *LatencyEvaluator) Name() string { return "latency" }

func (e *LatencyEvaluator) Evaluate(_ context.Context, t api.Trajectory) (api.EvalScore, error) {
	if len(t.Spans) == 0 {
		return api.EvalScore{Evaluator: e.Name()}, api.ErrInvalidTrajectory
	}
	var total int64
	for _, s := range t.Spans {
		total += s.DurationMs
	}
	budget := e.BudgetMs
	if budget == 0 {
		budget = 5000 // default 5s
	}
	ratio := float64(total) / float64(budget)
	score := 1.0 - ratio
	if score < 0 {
		score = 0
	}
	return api.EvalScore{
		Evaluator: e.Name(),
		Score:     score,
		MaxScore:  1.0,
		Pass:      total <= budget,
		Detail:    fmt.Sprintf("total %dms, budget %dms", total, budget),
	}, nil
}

// CostEvaluator checks that the number of spans (proxy for cost) is within budget.
type CostEvaluator struct {
	MaxSpans int
}

func (e *CostEvaluator) Name() string { return "cost" }

func (e *CostEvaluator) Evaluate(_ context.Context, t api.Trajectory) (api.EvalScore, error) {
	if len(t.Spans) == 0 {
		return api.EvalScore{Evaluator: e.Name()}, api.ErrInvalidTrajectory
	}
	max := e.MaxSpans
	if max == 0 {
		max = 20
	}
	count := len(t.Spans)
	score := 1.0 - float64(count)/float64(max)
	if score < 0 {
		score = 0
	}
	return api.EvalScore{
		Evaluator: e.Name(),
		Score:     score,
		MaxScore:  1.0,
		Pass:      count <= max,
		Detail:    fmt.Sprintf("%d spans, max %d", count, max),
	}, nil
}

// SafetyEvaluator checks that no span has a DENY outcome on safety-critical modules.
type SafetyEvaluator struct {
	CriticalModules []string
}

func (e *SafetyEvaluator) Name() string { return "safety" }

func (e *SafetyEvaluator) Evaluate(_ context.Context, t api.Trajectory) (api.EvalScore, error) {
	if len(t.Spans) == 0 {
		return api.EvalScore{Evaluator: e.Name()}, api.ErrInvalidTrajectory
	}
	critical := e.CriticalModules
	if len(critical) == 0 {
		critical = []string{"gate", "decision"}
	}
	criticalMap := make(map[string]bool, len(critical))
	for _, m := range critical {
		criticalMap[m] = true
	}
	violations := 0
	for _, s := range t.Spans {
		if criticalMap[s.Module] && s.Outcome == "DENY" && s.Error != nil {
			violations++
		}
	}
	score := 1.0
	if violations > 0 {
		score = 0.0
	}
	return api.EvalScore{
		Evaluator: e.Name(),
		Score:     score,
		MaxScore:  1.0,
		Pass:      violations == 0,
		Detail:    fmt.Sprintf("%d safety violation(s) on critical modules", violations),
	}, nil
}
