package capability

import (
    "errors"
    "fmt"
    "sync"
    "time"
)

// RiskClass defines the risk level of a capability.
type RiskClass int

const (
    RiskLow      RiskClass = iota // read-only, no external side effects
    RiskMedium                    // writes to internal state
    RiskHigh                      // external API calls, financial operations
    RiskCritical                  // destructive operations, PII access, payment
)

func (r RiskClass) String() string {
    switch r {
    case RiskLow:      return "LOW"
    case RiskMedium:   return "MEDIUM"
    case RiskHigh:     return "HIGH"
    case RiskCritical: return "CRITICAL"
    default:           return "UNKNOWN"
    }
}

// Protocol defines how a capability is invoked.
type Protocol int

const (
    ProtocolMCP   Protocol = iota // Model Context Protocol
    ProtocolGRPC                  // gRPC sidecar
    ProtocolREST                  // HTTP REST endpoint
    ProtocolWASM                  // WASM module via wazero
    ProtocolNative                // native Go call within same process
)

func (p Protocol) String() string {
    switch p {
    case ProtocolMCP:    return "MCP"
    case ProtocolGRPC:   return "GRPC"
    case ProtocolREST:   return "REST"
    case ProtocolWASM:   return "WASM"
    case ProtocolNative: return "NATIVE"
    default:             return "UNKNOWN"
    }
}

// RateLimit defines request-rate constraints on a capability.
type RateLimit struct {
    RequestsPerMinute int
    BurstSize         int
}

// BlastRadius defines the scope of side effects a capability can cause.
type BlastRadius struct {
    ExternalAPICalls bool // touches systems outside Pranor
    WritesDB         bool // modifies persistent state
    SendsNotification bool // fires emails, SMS, webhooks
    MaxAffectedRows  int  // 0 = unlimited
}

// CapabilitySchema holds input/output schema descriptors.
type CapabilitySchema struct {
    InputSchema  string // JSON schema string
    OutputSchema string // JSON schema string
    Description  string
}

// Capability is a first-class governed resource describing a tool or action an agent can invoke.
type Capability struct {
    ID             string           // stable dot-namespaced ID: e.g. "pool.query", "notify.send"
    Version        string           // semver: "1.0.0"
    Name           string           // human-readable name
    Description    string
    Schema         CapabilitySchema
    Risk           RiskClass
    RequiredPerms  []string         // permission identifiers required to invoke
    AllowedAgents  []string         // empty = all agents allowed
    AllowedTenants []string         // empty = all tenants allowed
    RateLimit      RateLimit
    BlastRadius    BlastRadius
    RequiresHITL   bool             // true = execution must pass HITL approval
    Protocol       Protocol
    Endpoint       string           // endpoint URI for GRPC/REST/WASM
    RegisteredAt   time.Time
}

// Sentinel errors
var (
    ErrCapabilityNotFound      = errors.New("pranor/core/capability: capability not found")
    ErrCapabilityUnauthorized  = errors.New("pranor/core/capability: agent not authorized for this capability")
    ErrCapabilityAlreadyExists = errors.New("pranor/core/capability: capability already registered")
    ErrInvalidCapability       = errors.New("pranor/core/capability: invalid capability definition")
)

// Registry is the interface for a Capability Registry.
type Registry interface {
    Register(c Capability) error
    Lookup(id string) (Capability, error)
    ListAll() []Capability
    ListByAgent(agentID string) []Capability
    ListByTenant(tenantID string) []Capability
    Authorize(tenantID, agentID, capID string) error
    Unregister(id string) error
}

// InMemoryRegistry is the OSS in-memory implementation of Registry.
type InMemoryRegistry struct {
    mu   sync.RWMutex
    caps map[string]Capability
}

// NewInMemoryRegistry returns a new empty registry.
func NewInMemoryRegistry() *InMemoryRegistry {
    return &InMemoryRegistry{caps: make(map[string]Capability)}
}

func (r *InMemoryRegistry) Register(c Capability) error {
    if c.ID == "" {
        return fmt.Errorf("%w: ID is required", ErrInvalidCapability)
    }
    if c.Version == "" {
        return fmt.Errorf("%w: Version is required", ErrInvalidCapability)
    }
    r.mu.Lock()
    defer r.mu.Unlock()
    if _, exists := r.caps[c.ID]; exists {
        return fmt.Errorf("%w: %s", ErrCapabilityAlreadyExists, c.ID)
    }
    c.RegisteredAt = time.Now().UTC()
    r.caps[c.ID] = c
    return nil
}

func (r *InMemoryRegistry) Lookup(id string) (Capability, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    c, ok := r.caps[id]
    if !ok {
        return Capability{}, fmt.Errorf("%w: %s", ErrCapabilityNotFound, id)
    }
    return c, nil
}

func (r *InMemoryRegistry) ListAll() []Capability {
    r.mu.RLock()
    defer r.mu.RUnlock()
    out := make([]Capability, 0, len(r.caps))
    for _, c := range r.caps {
        out = append(out, c)
    }
    return out
}

func (r *InMemoryRegistry) ListByAgent(agentID string) []Capability {
    r.mu.RLock()
    defer r.mu.RUnlock()
    var out []Capability
    for _, c := range r.caps {
        if isAllowed(c.AllowedAgents, agentID) {
            out = append(out, c)
        }
    }
    return out
}

func (r *InMemoryRegistry) ListByTenant(tenantID string) []Capability {
    r.mu.RLock()
    defer r.mu.RUnlock()
    var out []Capability
    for _, c := range r.caps {
        if isAllowed(c.AllowedTenants, tenantID) {
            out = append(out, c)
        }
    }
    return out
}

func (r *InMemoryRegistry) Authorize(tenantID, agentID, capID string) error {
    cap, err := r.Lookup(capID)
    if err != nil {
        return err
    }
    if !isAllowed(cap.AllowedTenants, tenantID) {
        return fmt.Errorf("%w: tenant %q not allowed for capability %q", ErrCapabilityUnauthorized, tenantID, capID)
    }
    if !isAllowed(cap.AllowedAgents, agentID) {
        return fmt.Errorf("%w: agent %q not allowed for capability %q", ErrCapabilityUnauthorized, agentID, capID)
    }
    return nil
}

func (r *InMemoryRegistry) Unregister(id string) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    if _, ok := r.caps[id]; !ok {
        return fmt.Errorf("%w: %s", ErrCapabilityNotFound, id)
    }
    delete(r.caps, id)
    return nil
}

// isAllowed returns true if list is empty (wildcard) or contains value.
func isAllowed(list []string, value string) bool {
    if len(list) == 0 {
        return true
    }
    for _, v := range list {
        if v == value || v == "*" {
            return true
        }
    }
    return false
}

// DefaultRegistry is the package-level registry. Safe to use from init().
var DefaultRegistry Registry = NewInMemoryRegistry()

// Register registers a capability in DefaultRegistry.
func Register(c Capability) error { return DefaultRegistry.Register(c) }

// Lookup looks up a capability in DefaultRegistry.
func Lookup(id string) (Capability, error) { return DefaultRegistry.Lookup(id) }

// Authorize checks authorization in DefaultRegistry.
func Authorize(tenantID, agentID, capID string) error {
    return DefaultRegistry.Authorize(tenantID, agentID, capID)
}
