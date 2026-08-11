# Pranor Decision — AI Governance Engine

**Version:** 2.0.0-dev  
**Module Path:** `github.com/vyuvaraj/pranor/decision`  
**License:** AGPL-3.0 (OSS) / EE

---

## Overview

Pranor Decision provides a Governed AI execution decision layer with a 6-level veto ladder. It ensures safe and predictable AI operations.

---

## Key Features

- **6-Level Priority Veto Ladder**
- **SIMULATION Mode:** Counterfactual evaluation without committing state
- **Fault Contracts**

---

## 6-Level Priority Veto Ladder

| Level | Name | Module | Hard/Soft | Effect |
|-------|------|--------|-----------|--------|
| 1 | Auth | decision | Hard | DENY blocks all subsequent levels |
| 2 | Budget | decision | Hard | DENY on cost/token overflow |
| 3 | Risk | decision | Soft | APPROVE/DENY from risk signals |
| 4 | Rules | decision | Soft | APPROVE/DENY/TRANSFORM policy rules |
| 5 | Learn | learn | Soft | Advisory from ML predictor (skip on timeout) |
| 6 | Default | decision | Hard | Final ALLOW fallback |

---

## Fault Contract

- Returns `DENY` if graph context is unavailable.
- Learn level is skipped on `ErrSidecarTimeout`.

---

## Types

**DecisionRequest**
Input parameters containing context, agent info, and action intent.

**DecisionResult**
Output containing the veto outcome, priority level hit, and metadata.

---

## Quick Start

```go
engine := decision.NewEngine(cfg)
ctx := context.Background()

// Standard execution
res, err := engine.Evaluate(ctx, decision.DecisionRequest{
    AgentID: "agent_88",
    Action:  "transfer_funds",
})

// Simulation mode
simRes, err := engine.Evaluate(ctx, decision.DecisionRequest{
    AgentID:   "agent_88",
    Action:    "transfer_funds",
    Simulate:  true, // Do not commit state
})
```

---

## Enterprise Edition

| Feature | OSS | EE |
|---------|:---:|:--:|
| Basic 6-level ladder | ✓ | ✓ |
| Simulation mode | ✓ | ✓ |
| Advanced Risk Models | — | ✓ |
| Custom Rules Engine UI | — | ✓ |
