# Capability Registry (`core/pkg/capability`)

**Package:** `github.com/vyuvaraj/pranor/core/pkg/capability`  
**Introduced:** Phase 91 (Sprint V2.91.2)

---

## Overview

Capabilities in Pranor v2.x are **first-class governed resources** rather than opaque tool names. Each capability defines its schema, risk classification, required permissions, rate limits, blast radius, HITL approval requirements, and protocol binding.

The `Capability Registry` acts as the single source of truth for tool resolution and authorization before execution at the Gate.

---

## Capability Schema

```go
type RiskClass int

const (
    RiskLow      RiskClass = iota // Read-only, internal state
    RiskMedium                    // Writes to internal state
    RiskHigh                      // External API calls, financial actions
    RiskCritical                  // Destructive ops, PII, payments
)

type Protocol int

const (
    ProtocolMCP   Protocol = iota // Model Context Protocol
    ProtocolGRPC                  // gRPC sidecar
    ProtocolREST                  // HTTP REST API
    ProtocolWASM                  // WASM sandbox via wazero
    ProtocolNative                // Native Go in-process call
)

type Capability struct {
    ID             string           `json:"id"`              // e.g. "pool.query", "notify.send"
    Version        string           `json:"version"`         // semver e.g. "1.0.0"
    Name           string           `json:"name"`
    Description    string           `json:"description"`
    Schema         CapabilitySchema `json:"schema"`          // JSON schema input/output
    Risk           RiskClass        `json:"risk"`            // LOW, MEDIUM, HIGH, CRITICAL
    RequiredPerms  []string         `json:"required_perms"`
    AllowedAgents  []string         `json:"allowed_agents"`  // empty = all allowed
    AllowedTenants []string         `json:"allowed_tenants"` // empty = all allowed
    RateLimit      RateLimit        `json:"rate_limit"`      // reqs/min, burst
    BlastRadius    BlastRadius      `json:"blast_radius"`    // external API, DB writes, notifications
    RequiresHITL   bool             `json:"requires_hitl"`   // requires Human-In-The-Loop approval
    Protocol       Protocol         `json:"protocol"`        // MCP, GRPC, REST, WASM, NATIVE
    Endpoint       string           `json:"endpoint"`        // URI for remote/sidecar calls
}
```

---

## Registry API

```go
type Registry interface {
    Register(c Capability) error
    Lookup(id string) (Capability, error)
    ListAll() []Capability
    ListByAgent(agentID string) []Capability
    ListByTenant(tenantID string) []Capability
    Authorize(tenantID, agentID, capID string) error
    Unregister(id string) error
}
```

- **OSS Implementation:** `InMemoryRegistry` (thread-safe `sync.RWMutex`, wildcard `*` matching for agent/tenant).
- **EE Implementation:** Persistent registry backed by `Pranor Vault` with cross-datacenter synchronization.

---

## Usage Example

```go
import "github.com/vyuvaraj/pranor/core/pkg/capability"

// Register a capability
capability.Register(capability.Capability{
    ID:          "pool.query",
    Version:     "1.0.0",
    Name:        "Database Query",
    Risk:        capability.RiskLow,
    Protocol:    capability.ProtocolNative,
    BlastRadius: capability.BlastRadius{WritesDB: false},
})

// Authorize before execution
err := capability.Authorize("tenant-acme", "agent-analyst", "pool.query")
if err != nil {
    // Fails closed if unauthorized
}
```
