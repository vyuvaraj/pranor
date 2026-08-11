# `agentctl` Developer CLI (`tools/agentctl`)

**Package:** `github.com/vyuvaraj/pranor/tools/agentctl`  
**Introduced:** Phase 92 (Sprint V2.92.3)

---

## Overview

`agentctl` is the official developer CLI tool for inspecting, debugging, replaying, and simulating Pranor agent executions locally.

---

## Commands & Usage

```bash
agentctl — Pranor Agent Developer CLI Tool

Commands:
  trace <session-id>           Print span waterfall trace summary
  replay <trajectory.json>     Replay trajectory & run quality evaluators
  budget [agent-id]            Display token & cost budget status
  policy simulate <req.json>   Dry-run Decision Engine policy simulation
```

---

## Command Details

### 1. `agentctl trace <session-id>`
Prints formatted OTLP span waterfall telemetry for an active or recorded agent session:
```bash
$ agentctl trace sess-8910
=== Agent Execution Trace: sess-8910 ===
Span: pranor.agent_execution [ALLOW] 12ms
Span: pranor.gate.inspect      [ALLOW] 2ms
Span: pranor.decision.evaluate [APPROVE] 4ms
```

### 2. `agentctl replay <trajectory.json>`
Loads a recorded trajectory JSON file, re-emits its spans through `eval.Replay`, and executes registered quality evaluators (`AccuracyEvaluator`, `LatencyEvaluator`, `CostEvaluator`, `SafetyEvaluator`):
```bash
$ agentctl replay trajectory_prod.json
✓ Trajectory replayed: tr-001-replay (spans: 4)
Evaluation Result: OverallPass=true
  - accuracy: score=1.00 pass=true (4/4 spans without error)
  - latency: score=0.98 pass=true (90ms, budget 5000ms)
```

### 3. `agentctl budget [agent-id]`
Displays token and cost quota consumption:
```bash
$ agentctl budget support-bot
=== Budget Status for Agent: support-bot ===
Token Quotas   : 45,000 / 100,000 tokens (45% used)
Daily Cost     : $0.14 / $5.00 USD
Status         : OK
```

### 4. `agentctl policy simulate <request.json>`
Performs counterfactual policy evaluation using `decision.Simulate` without executing side effects or mutating backend state:
```bash
$ agentctl policy simulate req.json
=== Decision Engine Policy Simulation ===
Request        : AgentID=support-bot TenantID=acme-corp
Evaluated      : Priority 1 (Auth) -> PASS, Priority 2 (Budget) -> PASS
Outcome        : APPROVE (Simulation Mode - No Side Effects Committed)
```
