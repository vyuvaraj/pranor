# Gate Shadow Execution (`gate/pkg/shadow`)

**Package:** `github.com/vyuvaraj/pranor/gate/pkg/shadow`  
**Introduced:** Phase 92 (Sprint V2.92.4)

---

## Overview

Gate Shadow Execution provides side-effect isolation when evaluating agents or policies in `SIMULATION` / shadow mode.

When a request contains header `X-Shadow-Mode: true` or `ec.PolicyContext["mode"] == "SIMULATION"`, the `shadow.Interceptor` at the Gate boundary:
- Allows read-only capability calls to execute normally.
- Intercepts write/destructive capability calls (database mutations, external API calls, notification triggers) and converts them into no-op mock responses with annotation `[SHADOW_MODE_NOOP]`.
- Emits `pranor.gate.shadow_execution` OTLP telemetry.

---

## Key Interface

```go
type Interceptor interface {
	IsShadowMode(ec *execctx.ExecutionContext) bool
	InterceptCapability(ctx context.Context, ec *execctx.ExecutionContext, capID string, input map[string]any) (map[string]any, bool, error)
}
```

---

## Behavior Matrix

| Operation Type | Real Mode (`REAL`) | Shadow Mode (`SIMULATION`) |
|----------------|-------------------|-----------------------------|
| Read (`RiskLow`, no DB writes) | Execute backend query | Execute backend query (Passthrough) |
| DB Write (`WritesDB = true`) | Execute database write | **`[SHADOW_MODE_NOOP]`** |
| External API (`ExternalAPICalls = true`) | HTTP POST / gRPC call | **`[SHADOW_MODE_NOOP]`** |
| Notification (`SendsNotification = true`) | Send Email / SMS / Webhook | **`[SHADOW_MODE_NOOP]`** |
