# ExecutionContext (`core/pkg/execctx`)

**Package:** `github.com/vyuvaraj/pranor/core/pkg/execctx`  
**Introduced:** Phase 91 (Sprint V2.91.1)

---

## Overview

`ExecutionContext` is the canonical propagation structure passed through all HTTP routes, WASM plugins, database queries, and background tasks in Pranor v2.x. It unifies identity, policy context, budget circuit breakers, and correlation IDs into a single struct embedding `context.Context`.

Every request boundary across Gate, Graph, Decision, Flow, Learn, and Tools MUST accept and pass `*execctx.ExecutionContext`.

---

## Type Definition

```go
type ExecutionContext struct {
    context.Context

    // Identity & Context Propagation
    TenantID      string `json:"tenant_id"`       // mandatory tenant isolation ID
    AgentID       string `json:"agent_id"`        // executing agent ID
    UserID        string `json:"user_id"`         // authenticated user ID
    TraceID       string `json:"trace_id"`        // OTLP trace ID
    RequestID     string `json:"request_id"`      // request correlation ID
    ParentAgentID string `json:"parent_agent_id"` // parent agent ID if spawned in A2A delegation

    // Capability & Policy
    Capabilities  []string          `json:"capabilities"`   // authorized capability IDs
    PolicyContext map[string]string `json:"policy_context"` // arbitrary key-value policy tags

    // Budget Limits & Circuit Breakers
    RiskBudget   float64 `json:"risk_budget"`    // 0.0-1.0 (0.0 = zero risk allowed)
    TokenBudget  int     `json:"token_budget"`   // max LLM tokens allowed
    CostBudgetUS float64 `json:"cost_budget_us"` // max USD cost allowed

    // Metadata
    Metadata  map[string]string `json:"metadata"`
    CreatedAt time.Time         `json:"created_at"`
}
```

---

## Key Functions & Builders

| Function | Description |
|----------|-------------|
| `New(ctx, tenantID, agentID, userID)` | Creates a new `ExecutionContext`. `TenantID` is required. |
| `FromHTTP(ctx, r)` | Extracts `ExecutionContext` from `X-Pranor-*` HTTP headers. Fails closed (`ErrMissingTenantID`) if missing. |
| `ec.WithAgent(agentID)` | Returns a shallow copy with `AgentID` set to `agentID` and `ParentAgentID` set to old `AgentID`. |
| `ec.WithCapability(capID)` | Returns a shallow copy with `capID` appended to `Capabilities`. |
| `ec.WithPolicy(key, value)` | Returns a shallow copy with updated `PolicyContext`. |
| `ec.WithBudget(risk, tokens, cost)` | Returns a shallow copy with updated budget limits. |
| `ec.Validate()` | Returns `ErrMissingTenantID` if `TenantID` is empty. |
| `ec.HasCapability(capID)` | Returns `true` if `capID` is in the authorized capabilities list. |
| `ec.InjectHTTP(r)` | Writes `X-Pranor-*` headers to an outgoing HTTP request. |

---

## Propagation Protocol

```text
HTTP Request (X-Pranor-Tenant-ID, X-Pranor-Agent-ID)
  ↓
Gate (execctx.FromHTTP)
  ↓
Capability Registry (ec.HasCapability)
  ↓
Decision & Graph (ec.RiskBudget, ec.TenantID)
  ↓
Flow & Tools (ec.InjectHTTP for downstream calls)
```

---

## HTTP Propagation Headers

- `X-Pranor-Tenant-ID`: Tenant isolation ID (Required)
- `X-Pranor-Agent-ID`: Executing Agent ID
- `X-Pranor-User-ID`: Authenticated User ID
- `X-Pranor-Trace-ID`: Distributed Trace ID
- `X-Pranor-Request-ID`: Correlation Request ID
- `X-Pranor-Parent-Agent-ID`: Parent Agent ID for A2A delegation
