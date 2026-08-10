package agentgov

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Decision represents the Firewall evaluation outcome.
type Decision string

const (
	DecisionAllow     Decision = "ALLOW"
	DecisionDeny      Decision = "DENY"
	DecisionApprove   Decision = "APPROVE"
	DecisionTransform Decision = "TRANSFORM"
)

// AgentSecurityChain encapsulates first-class Agent -> User -> Tenant -> Capability identity. (EE.88.2)
type AgentSecurityChain struct {
	AgentID      string            `json:"agent_id"`
	UserID       string            `json:"user_id"`
	TenantID     string            `json:"tenant_id"`
	CapabilityID string            `json:"capability_id"`
	SessionID    string            `json:"session_id,omitempty"`
	DelegatedBy  string            `json:"delegated_by,omitempty"`
	Scopes       []string          `json:"scopes,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ToolCallPayload represents an intercepted tool call or AI capability invocation.
type ToolCallPayload struct {
	ToolName  string                 `json:"tool_name"`
	Arguments map[string]interface{} `json:"arguments"`
	RawIntent string                 `json:"raw_intent,omitempty"`
}

// FirewallEvaluationResult is the output of the Security Firewall inspection. (EE.88.1)
type FirewallEvaluationResult struct {
	Decision             Decision               `json:"decision"`
	RiskScore            float64                `json:"risk_score"`
	Reason               string                 `json:"reason"`
	TransformedArguments map[string]interface{} `json:"transformed_arguments,omitempty"`
	ApprovalID           string                 `json:"approval_id,omitempty"`
	ExecutionTrace       *ExecutionTrace        `json:"execution_trace,omitempty"`
}

// ApprovalStatus represents the state of a Human-in-the-Loop request. (EE.88.3)
type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "PENDING"
	ApprovalStatusApproved ApprovalStatus = "APPROVED"
	ApprovalStatusRejected ApprovalStatus = "REJECTED"
	ApprovalStatusExpired  ApprovalStatus = "EXPIRED"
)

// HITLApprovalRecord stores an asynchronous approval request.
type HITLApprovalRecord struct {
	ApprovalID  string             `json:"approval_id"`
	Status      ApprovalStatus     `json:"status"`
	Chain       AgentSecurityChain `json:"chain"`
	ToolCall    ToolCallPayload    `json:"tool_call"`
	Reason      string             `json:"reason"`
	RequestedAt time.Time          `json:"requested_at"`
	ExpiresAt   time.Time          `json:"expires_at"`
	ApprovedBy  string             `json:"approved_by,omitempty"`
	ApprovedAt  *time.Time         `json:"approved_at,omitempty"`
}

// TrajectoryStep records an agent execution step for observability & simulation replay. (EE.88.4)
type TrajectoryStep struct {
	Timestamp      time.Time                `json:"timestamp"`
	Chain          AgentSecurityChain       `json:"chain"`
	ToolCall       ToolCallPayload          `json:"tool_call"`
	ResultDecision FirewallEvaluationResult `json:"result_decision"`
	ExecutionMs    int64                    `json:"execution_ms"`
}

// AgentBudgetTracks session & agent tool invocation limits. (EE.88.6)
type AgentBudget struct {
	MaxToolCallsPerSession int `json:"max_tool_calls_per_session"`
	CurrentToolCalls       int `json:"current_tool_calls"`
}

// ExecutionTrace visualizes the end-to-end agent execution tree. (EE.88.10)
type ExecutionTrace struct {
	User               string   `json:"user"`
	Agent              string   `json:"agent"`
	Intent             string   `json:"intent"`
	CapabilitySelected string   `json:"capability_selected"`
	PolicyDecision     Decision `json:"policy_decision"`
	Result             string   `json:"result"`
}

// SecurityFirewall Engine manages all Phase 88 Next-Gen Agent Governance modules.
type SecurityFirewall struct {
	mu            sync.RWMutex
	approvals     map[string]*HITLApprovalRecord
	trajectories  []TrajectoryStep
	highRiskTools map[string]float64
	agentBudgets  map[string]*AgentBudget // EE.88.6
}

var (
	globalFirewall *SecurityFirewall
	firewallOnce   sync.Once
)

// GetFirewall returns the singleton instance of SecurityFirewall.
func GetFirewall() *SecurityFirewall {
	firewallOnce.Do(func() {
		globalFirewall = &SecurityFirewall{
			approvals: make(map[string]*HITLApprovalRecord),
			highRiskTools: map[string]float64{
				"payment_refund":    0.9,
				"database_drop":     0.95,
				"user_delete":       0.85,
				"system_reboot":     0.99,
				"permissions_grant": 0.8,
			},
			agentBudgets: make(map[string]*AgentBudget),
		}
	})
	return globalFirewall
}

// ExtractSecurityChain inspects HTTP headers to assemble the Agent -> User -> Tenant -> Capability chain (EE.88.2 & EE.88.8).
func (f *SecurityFirewall) ExtractSecurityChain(req *http.Request) AgentSecurityChain {
	chain := AgentSecurityChain{
		AgentID:      req.Header.Get("X-Agent-ID"),
		UserID:       req.Header.Get("X-User-ID"),
		TenantID:     req.Header.Get("X-Tenant-ID"),
		CapabilityID: req.Header.Get("X-Capability-ID"),
		SessionID:    req.Header.Get("X-Agent-Session-ID"),
		DelegatedBy:  req.Header.Get("X-Agent-Delegated-By"),
	}

	if chain.AgentID == "" {
		chain.AgentID = "agent-anonymous"
	}
	if chain.UserID == "" {
		chain.UserID = "user-default"
	}
	if chain.TenantID == "" {
		chain.TenantID = "tenant-default"
	}
	return chain
}

// EvaluateToolCall evaluates an incoming tool call across all EE.88 governance modules.
func (f *SecurityFirewall) EvaluateToolCall(chain AgentSecurityChain, toolCall ToolCallPayload) FirewallEvaluationResult {
	f.mu.Lock()
	defer f.mu.Unlock()

	// EE.88.5: AI Capability Risk & Trust Engine
	riskScore := 0.1
	if score, exists := f.highRiskTools[toolCall.ToolName]; exists {
		riskScore = score
	}

	// EE.88.6: Agent Blast-Radius & Tool Budget Enforcer
	budgetKey := chain.AgentID + ":" + chain.SessionID
	if budget, exists := f.agentBudgets[budgetKey]; exists {
		if budget.MaxToolCallsPerSession > 0 && budget.CurrentToolCalls >= budget.MaxToolCallsPerSession {
			return FirewallEvaluationResult{
				Decision:  DecisionDeny,
				RiskScore: riskScore,
				Reason:    fmt.Sprintf("Agent tool call budget exceeded (%d/%d tool calls)", budget.CurrentToolCalls, budget.MaxToolCallsPerSession),
			}
		}
		budget.CurrentToolCalls++
	} else {
		f.agentBudgets[budgetKey] = &AgentBudget{
			MaxToolCallsPerSession: 100,
			CurrentToolCalls:       1,
		}
	}

	// EE.88.8: Agent-to-Agent Delegation Check
	if chain.DelegatedBy != "" {
		// Enforce strict delegation policy check
		if strings.HasPrefix(toolCall.ToolName, "admin_") {
			return FirewallEvaluationResult{
				Decision:  DecisionDeny,
				RiskScore: riskScore,
				Reason:    fmt.Sprintf("Delegated agent (%s -> %s) denied execution of admin capability %s", chain.DelegatedBy, chain.AgentID, toolCall.ToolName),
			}
		}
	}

	// EE.88.3: Human-in-the-Loop Execution Engine
	if riskScore >= 0.85 {
		approvalID := generateApprovalID(chain, toolCall)
		if record, ok := f.approvals[approvalID]; ok {
			if record.Status == ApprovalStatusApproved && time.Now().Before(record.ExpiresAt) {
				return FirewallEvaluationResult{
					Decision:  DecisionAllow,
					RiskScore: riskScore,
					Reason:    fmt.Sprintf("HITL Pre-approved by %s", record.ApprovedBy),
				}
			} else if record.Status == ApprovalStatusRejected {
				return FirewallEvaluationResult{
					Decision:  DecisionDeny,
					RiskScore: riskScore,
					Reason:    "HITL Approval Explicitly Rejected",
				}
			}
		}

		newRecord := &HITLApprovalRecord{
			ApprovalID:  approvalID,
			Status:      ApprovalStatusPending,
			Chain:       chain,
			ToolCall:    toolCall,
			Reason:      fmt.Sprintf("High risk tool execution (%s risk=%.2f) requires human approval", toolCall.ToolName, riskScore),
			RequestedAt: time.Now(),
			ExpiresAt:   time.Now().Add(15 * time.Minute),
		}
		f.approvals[approvalID] = newRecord

		return FirewallEvaluationResult{
			Decision:   DecisionApprove,
			RiskScore:  riskScore,
			Reason:     newRecord.Reason,
			ApprovalID: approvalID,
		}
	}

	// EE.88.7: Agent Memory & Context Governance (PII & Context Filtering)
	transformed := make(map[string]interface{})
	needsTransform := false
	for k, v := range toolCall.Arguments {
		if k == "ssn" || k == "credit_card" || k == "api_secret" {
			transformed[k] = "[REDACTED-BY-GATE-FIREWALL]"
			needsTransform = true
		} else {
			transformed[k] = v
		}
	}

	// EE.88.10: End-to-End Agent Execution Trace Construction
	trace := &ExecutionTrace{
		User:               chain.UserID,
		Agent:              chain.AgentID,
		Intent:             toolCall.RawIntent,
		CapabilitySelected: toolCall.ToolName,
		PolicyDecision:     DecisionAllow,
		Result:             "PENDING_EXECUTION",
	}

	if needsTransform {
		trace.PolicyDecision = DecisionTransform
		return FirewallEvaluationResult{
			Decision:             DecisionTransform,
			RiskScore:            riskScore,
			Reason:               "Transformed sensitive arguments prior to execution",
			TransformedArguments: transformed,
			ExecutionTrace:       trace,
		}
	}

	return FirewallEvaluationResult{
		Decision:       DecisionAllow,
		RiskScore:      riskScore,
		Reason:         "Passed Agent Security Firewall policy evaluation",
		ExecutionTrace: trace,
	}
}

// ApprovePendingToolCall updates a pending approval record.
func (f *SecurityFirewall) ApprovePendingToolCall(approvalID, approver string, approve bool) (*HITLApprovalRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	rec, ok := f.approvals[approvalID]
	if !ok {
		return nil, fmt.Errorf("approval request %s not found", approvalID)
	}

	now := time.Now()
	rec.ApprovedBy = approver
	rec.ApprovedAt = &now
	if approve {
		rec.Status = ApprovalStatusApproved
	} else {
		rec.Status = ApprovalStatusRejected
	}

	return rec, nil
}

// GetPendingApprovals returns all currently pending approval requests.
func (f *SecurityFirewall) GetPendingApprovals() []*HITLApprovalRecord {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var pending []*HITLApprovalRecord
	for _, rec := range f.approvals {
		if rec.Status == ApprovalStatusPending && time.Now().Before(rec.ExpiresAt) {
			pending = append(pending, rec)
		}
	}
	return pending
}

// RecordTrajectoryStep appends an agent execution step.
func (f *SecurityFirewall) RecordTrajectoryStep(step TrajectoryStep) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.trajectories = append(f.trajectories, step)
}

// GetTrajectories returns recorded execution trajectories for replay and simulation diffs.
func (f *SecurityFirewall) GetTrajectories() []TrajectoryStep {
	f.mu.RLock()
	defer f.mu.RUnlock()

	result := make([]TrajectoryStep, len(f.trajectories))
	copy(result, f.trajectories)
	return result
}

// SimulateTrajectoryReplay runs recorded agent trajectories against current policies (EE.88.4).
func (f *SecurityFirewall) SimulateTrajectoryReplay() map[string]interface{} {
	f.mu.RLock()
	trajectories := make([]TrajectoryStep, len(f.trajectories))
	copy(trajectories, f.trajectories)
	f.mu.RUnlock()

	total := len(trajectories)
	allowed, denied, approved, transformed := 0, 0, 0, 0
	diffs := []map[string]interface{}{}

	for _, step := range trajectories {
		simRes := f.EvaluateToolCall(step.Chain, step.ToolCall)
		if simRes.Decision != step.ResultDecision.Decision {
			diffs = append(diffs, map[string]interface{}{
				"agent_id":        step.Chain.AgentID,
				"tool":            step.ToolCall.ToolName,
				"original_action": step.ResultDecision.Decision,
				"simulated_action": simRes.Decision,
				"reason":          simRes.Reason,
			})
		}

		switch simRes.Decision {
		case DecisionAllow:
			allowed++
		case DecisionDeny:
			denied++
		case DecisionApprove:
			approved++
		case DecisionTransform:
			transformed++
		}
	}

	return map[string]interface{}{
		"total_trajectories_replayed": total,
		"simulated_summary": map[string]int{
			"ALLOW":     allowed,
			"DENY":      denied,
			"APPROVE":   approved,
			"TRANSFORM": transformed,
		},
		"policy_diff_count": len(diffs),
		"diffs":             diffs,
	}
}

// EE.88.9: ProtocolAgnosticExposer exposes registered capabilities across MCP, gRPC, and HTTP protocols.
func (f *SecurityFirewall) ProtocolAgnosticExposer() map[string]interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()

	capabilities := []map[string]interface{}{}
	for name, risk := range f.highRiskTools {
		capabilities = append(capabilities, map[string]interface{}{
			"capability_name": name,
			"risk_score":      risk,
			"adapters":        []string{"MCP", "gRPC", "HTTP/REST", "WASM"},
		})
	}

	return map[string]interface{}{
		"status":                "active",
		"total_capabilities":    len(capabilities),
		"protocol_adapters":     []string{"MCP", "gRPC", "HTTP/REST", "WASM"},
		"registered_capabilities": capabilities,
	}
}

func generateApprovalID(chain AgentSecurityChain, toolCall ToolCallPayload) string {
	raw := fmt.Sprintf("%s:%s:%s:%s", chain.AgentID, chain.UserID, toolCall.ToolName, time.Now().Truncate(time.Minute).String())
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:8])
}

// InterceptAndGovernToolCall wraps HTTP request body parsing for MCP / Tool call payloads.
func InterceptAndGovernToolCall(r *http.Request) ([]byte, AgentSecurityChain, FirewallEvaluationResult, error) {
	fw := GetFirewall()
	chain := fw.ExtractSecurityChain(r)

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, chain, FirewallEvaluationResult{}, err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var payload ToolCallPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil || payload.ToolName == "" {
		var jsonrpc struct {
			Method string `json:"method"`
			Params struct {
				Name      string                 `json:"name"`
				Arguments map[string]interface{} `json:"arguments"`
			} `json:"params"`
		}
		if jsonErr := json.Unmarshal(bodyBytes, &jsonrpc); jsonErr == nil && jsonrpc.Params.Name != "" {
			payload.ToolName = jsonrpc.Params.Name
			payload.Arguments = jsonrpc.Params.Arguments
		} else {
			payload.ToolName = r.URL.Path
		}
	}

	start := time.Now()
	eval := fw.EvaluateToolCall(chain, payload)
	execMs := time.Since(start).Milliseconds()

	fw.RecordTrajectoryStep(TrajectoryStep{
		Timestamp:      time.Now(),
		Chain:          chain,
		ToolCall:       payload,
		ResultDecision: eval,
		ExecutionMs:    execMs,
	})

	return bodyBytes, chain, eval, nil
}
