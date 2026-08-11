# Agent-to-Agent Delegation Protocol (`agent/pkg/a2a`)

**Package:** `github.com/vyuvaraj/pranor/agent/pkg/a2a`  
**Introduced:** Phase 93 (Sprint V2.93.3)

---

## Overview

The A2A Delegation Protocol (`agent/pkg/a2a`) enables secure inter-agent task delegation. It enforces **capability escalation prevention** (child agents cannot inherit permissions beyond what the parent possesses) and handles automatic identity propagation.

---

## Data Structures

```go
type DelegationRequest struct {
	ChildAgentID          string         `json:"child_agent_id"`
	SubTaskPayload        map[string]any `json:"subtask_payload"`
	RequestedCapabilities []string       `json:"requested_capabilities"`
	RiskBudget            float64        `json:"risk_budget"`
	TokenBudget           int            `json:"token_budget"`
}

type DelegationResult struct {
	SessionID     string                   `json:"session_id"`
	Status        string                   `json:"status"` // "SUCCESS", "FAILED"
	OutputPayload map[string]any           `json:"output_payload"`
	TokensUsed    int                      `json:"tokens_used"`
	CostUSD       float64                  `json:"cost_usd"`
	ChildExecCtx  *execctx.ExecutionContext `json:"child_exec_ctx"`
}
```

---

## Delegation Sequence

```text
Parent Agent (AgentID: parent-1, Capabilities: [pool.query, notify.send])
  │
  ├── Delegate(child-1, RequestedCapabilities: [pool.query])
  │     ├── Check Escalation: pool.query ∈ parent capabilities? -> YES
  │     ├── Create Child ExecCtx: parentEC.WithAgent("child-1")
  │     │     (ParentAgentID: "parent-1", AgentID: "child-1")
  │     └── Execute Subtask -> SUCCESS
  │
  └── Delegate(child-2, RequestedCapabilities: [secret.delete])
        └── Check Escalation: secret.delete ∈ parent capabilities? -> NO
              └── Returns ErrCapabilityEscalationDenied (Fail-Closed)
```

---

## Code Example

```go
import "github.com/vyuvaraj/pranor/agent/pkg/a2a"

delegator := a2a.NewOSSDelegator()

res, err := delegator.Delegate(ctx, parentEC, a2a.DelegationRequest{
	ChildAgentID:          "sub-analyst",
	RequestedCapabilities: []string{"pool.query"},
	SubTaskPayload:        map[string]any{"query": "SELECT count(*) FROM orders"},
})
if err == a2a.ErrCapabilityEscalationDenied {
	// Child attempted to escalate permissions beyond parent
}
```
