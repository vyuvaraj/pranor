package execctx

import (
    "context"
    "errors"
    "net/http"
    "strings"
    "time"
)

// Key constants for HTTP header extraction
const (
    HeaderTenantID      = "X-Pranor-Tenant-ID"
    HeaderAgentID       = "X-Pranor-Agent-ID"
    HeaderUserID        = "X-Pranor-User-ID"
    HeaderTraceID       = "X-Pranor-Trace-ID"
    HeaderParentAgentID = "X-Pranor-Parent-Agent-ID"
    HeaderRequestID     = "X-Pranor-Request-ID"
)

// Sentinel errors
var (
    ErrMissingTenantID = errors.New("pranor/core/execctx: TenantID is required")
    ErrMissingAgentID  = errors.New("pranor/core/execctx: AgentID is required")
)

// ExecutionContext is the canonical propagation struct for all Pranor v2.x module boundaries.
// It carries identity, policy, budget, and observability context across Gate → Graph → Decision → Flow → Learn → Tool.
type ExecutionContext struct {
    context.Context

    // Identity
    TenantID      string
    AgentID       string
    UserID        string
    TraceID       string
    RequestID     string
    ParentAgentID string // set when a child agent is spawned by a parent

    // Capability & Policy
    Capabilities  []string          // list of capability IDs this execution is authorized for
    PolicyContext map[string]string // arbitrary key-value policy annotations

    // Budget
    RiskBudget   float64 // 0.0–1.0; 0.0 = no risk allowed, 1.0 = any risk allowed
    TokenBudget  int     // max LLM tokens for this execution; 0 = unlimited
    CostBudgetUS float64 // max USD cost; 0 = unlimited

    // Metadata
    Metadata  map[string]string
    CreatedAt time.Time
}

// New creates an ExecutionContext with required fields.
func New(ctx context.Context, tenantID, agentID, userID string) *ExecutionContext {
    return &ExecutionContext{
        Context:       ctx,
        TenantID:      tenantID,
        AgentID:       agentID,
        UserID:        userID,
        Capabilities:  []string{},
        PolicyContext: map[string]string{},
        Metadata:      map[string]string{},
        CreatedAt:     time.Now().UTC(),
    }
}

// FromHTTP extracts an ExecutionContext from standard Pranor HTTP headers.
// Returns ErrMissingTenantID if the X-Pranor-Tenant-ID header is absent.
func FromHTTP(ctx context.Context, r *http.Request) (*ExecutionContext, error) {
    tenantID := r.Header.Get(HeaderTenantID)
    if strings.TrimSpace(tenantID) == "" {
        return nil, ErrMissingTenantID
    }
    ec := New(ctx, tenantID, r.Header.Get(HeaderAgentID), r.Header.Get(HeaderUserID))
    ec.TraceID       = r.Header.Get(HeaderTraceID)
    ec.RequestID     = r.Header.Get(HeaderRequestID)
    ec.ParentAgentID = r.Header.Get(HeaderParentAgentID)
    return ec, nil
}

// WithAgent returns a shallow copy with a new AgentID.
func (ec *ExecutionContext) WithAgent(agentID string) *ExecutionContext {
    cp := ec.clone()
    cp.ParentAgentID = ec.AgentID
    cp.AgentID = agentID
    return cp
}

// WithCapability returns a shallow copy with the capability appended.
func (ec *ExecutionContext) WithCapability(capID string) *ExecutionContext {
    cp := ec.clone()
    caps := make([]string, len(ec.Capabilities)+1)
    copy(caps, ec.Capabilities)
    caps[len(ec.Capabilities)] = capID
    cp.Capabilities = caps
    return cp
}

// WithPolicy returns a shallow copy with the policy key-value set.
func (ec *ExecutionContext) WithPolicy(key, value string) *ExecutionContext {
    cp := ec.clone()
    cp.PolicyContext = copyMap(ec.PolicyContext)
    cp.PolicyContext[key] = value
    return cp
}

// WithBudget returns a shallow copy with budget fields set.
func (ec *ExecutionContext) WithBudget(riskBudget float64, tokenBudget int, costBudgetUS float64) *ExecutionContext {
    cp := ec.clone()
    cp.RiskBudget   = riskBudget
    cp.TokenBudget  = tokenBudget
    cp.CostBudgetUS = costBudgetUS
    return cp
}

// Validate ensures required fields are populated.
func (ec *ExecutionContext) Validate() error {
    if strings.TrimSpace(ec.TenantID) == "" {
        return ErrMissingTenantID
    }
    return nil
}

// HasCapability reports whether capID is in the authorized capabilities list.
func (ec *ExecutionContext) HasCapability(capID string) bool {
    for _, c := range ec.Capabilities {
        if c == capID {
            return true
        }
    }
    return false
}

// InjectHTTP writes Pranor propagation headers onto an outgoing request.
func (ec *ExecutionContext) InjectHTTP(r *http.Request) {
    r.Header.Set(HeaderTenantID,      ec.TenantID)
    r.Header.Set(HeaderAgentID,       ec.AgentID)
    r.Header.Set(HeaderUserID,        ec.UserID)
    r.Header.Set(HeaderTraceID,       ec.TraceID)
    r.Header.Set(HeaderRequestID,     ec.RequestID)
    r.Header.Set(HeaderParentAgentID, ec.ParentAgentID)
}

func (ec *ExecutionContext) clone() *ExecutionContext {
    cp := *ec
    return &cp
}

func copyMap(m map[string]string) map[string]string {
    out := make(map[string]string, len(m))
    for k, v := range m {
        out[k] = v
    }
    return out
}
