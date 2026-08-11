# Multi-Tenant Sandboxing & Rate Limiting (`core/pkg/tenant`)

**Package:** `github.com/vyuvaraj/pranor/core/pkg/tenant`  
**Introduced:** Phase 93 (Sprint V2.93.2)

---

## Overview

Pranor Tenant (`core/pkg/tenant`) enforces multi-tenant resource quotas, request rate limiting, and daily token/cost bounds to guarantee hard tenant isolation and prevent runaway billing or resource starvation.

---

## Data Structures

```go
type Quota struct {
	MaxRequestsPerMin   int     `json:"max_requests_per_min"`
	MaxConcurrentAgents int     `json:"max_concurrent_agents"`
	MaxTokensPerDay     int     `json:"max_tokens_per_day"`
	MaxCostUSDPerDay    float64 `json:"max_cost_usd_per_day"`
}

type UsageStats struct {
	RequestsThisMin int       `json:"requests_this_min"`
	ActiveAgents    int       `json:"active_agents"`
	TokensToday     int       `json:"tokens_today"`
	CostUSDToday    float64   `json:"cost_usd_today"`
	LastResetMinute time.Time `json:"last_reset_minute"`
	LastResetDay    time.Time `json:"last_reset_day"`
}
```

---

## Enforcer API

```go
type Enforcer interface {
	SetQuota(tenantID string, q Quota)
	GetQuota(tenantID string) (Quota, bool)
	Enforce(ec *execctx.ExecutionContext) error
	RecordUsage(ec *execctx.ExecutionContext, tokens int, costUSD float64) error
	ReleaseAgent(ec *execctx.ExecutionContext)
}
```

- **Enforce:** Called at Gate ingress. Returns `ErrTenantRateLimited` if request rate or active agents exceed quota, or `ErrTenantQuotaExceeded` if daily token or cost limits are hit.
- **RecordUsage:** Called post-execution to update daily token and USD cost counters.

---

## Code Example

```go
import "github.com/vyuvaraj/pranor/core/pkg/tenant"

enforcer := tenant.NewOSSEnforcer()
enforcer.SetQuota("tenant-acme", tenant.Quota{
	MaxRequestsPerMin:   60,
	MaxConcurrentAgents: 5,
	MaxTokensPerDay:     100000,
	MaxCostUSDPerDay:    10.00,
})

// Check quota before execution
if err := enforcer.Enforce(ec); err != nil {
	// Returns ErrTenantRateLimited or ErrTenantQuotaExceeded
}
```
