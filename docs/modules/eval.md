# Pranor Eval — Agent Quality Scoring

**Version:** 2.0.0-dev  
**Module Path:** `github.com/vyuvaraj/pranor/eval`  
**License:** AGPL-3.0 (OSS) / EE

---

## Overview

Pranor Eval is a trajectory-based quality scoring and replay framework for AI agents, allowing offline and online evaluation of AI behavior.

---

## Key Features

- **4 Evaluators:** Accuracy, Latency, Cost, Safety
- **Soft-fail guarantee:** A single evaluator panic/error degrades the score but doesn't abort the run.

---

## Evaluators

| Name | Metric | Pass Threshold | Description |
|------|--------|----------------|-------------|
| AccuracyEvaluator | Error-free span rate | ≥80% | Validates agent output matches expected outcomes without internal errors. |
| LatencyEvaluator | Total DurationMs vs BudgetMs | Within budget | Ensures execution completes within SLA timeouts. |
| CostEvaluator | Span count vs MaxSpans | Within max | Bounds agent exploration steps and LLM token usage. |
| SafetyEvaluator | DENY outcomes on critical modules | 0 violations | Strictly checks for security or policy vetoes. |

---

## API Reference

### EvalEngine API

- `Register(evaluator Evaluator)`: Register a new evaluator.
- `Run(ctx context.Context, trajectory Trajectory) (EvalResult, error)`: Run evaluation on a trajectory.
- `Replay(ctx context.Context, id string) (Trajectory, error)`: Fetch and replay a previous run.

### Trajectory Types
- **TrajectorySpan:** Individual unit of execution.
- **Trajectory:** Collection of spans representing an execution path.
- **EvalScore:** Individual evaluator score.
- **EvalResult:** Final aggregated result.

---

## Quick Start

```go
engine := eval.NewEvalEngine()
engine.Register(eval.NewAccuracyEvaluator())
engine.Register(eval.NewSafetyEvaluator())

trajectory := getAgentTrajectory("exec_123")
result, _ := engine.Run(context.Background(), trajectory)
fmt.Println("Score:", result.TotalScore)
```

---

## Enterprise Edition

| Feature | OSS | EE |
|---------|:---:|:--:|
| Local replay | ✓ | ✓ |
| CI/CD quality gate | — | ✓ |
| Trace archive | — | ✓ |
