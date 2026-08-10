package agentgov_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vyuvaraj/pranor/gate/pkg/agentgov"
	"github.com/vyuvaraj/pranor/gate/pkg/mcp"
)

func TestSecurityFirewall_EvaluateToolCall(t *testing.T) {
	fw := agentgov.GetFirewall()

	chain := agentgov.AgentSecurityChain{
		AgentID:      "agent-007",
		UserID:       "user-alice",
		TenantID:     "tenant-acme",
		CapabilityID: "tool-test",
		SessionID:    "sess-100",
	}

	// Test 1: Low-risk tool call ALLOW
	res1 := fw.EvaluateToolCall(chain, agentgov.ToolCallPayload{
		ToolName:  "player_read",
		Arguments: map[string]interface{}{"player_id": "123"},
	})
	if res1.Decision != agentgov.DecisionAllow {
		t.Fatalf("expected ALLOW for low-risk tool, got %s", res1.Decision)
	}

	// Test 2: High-risk tool call APPROVE (HITL)
	res2 := fw.EvaluateToolCall(chain, agentgov.ToolCallPayload{
		ToolName:  "payment_refund",
		Arguments: map[string]interface{}{"amount": 5000},
	})
	if res2.Decision != agentgov.DecisionApprove {
		t.Fatalf("expected APPROVE for high-risk tool, got %s", res2.Decision)
	}
	if res2.ApprovalID == "" {
		t.Fatalf("expected valid ApprovalID for HITL workflow")
	}

	// Test 3: HITL Approve Action
	record, err := fw.ApprovePendingToolCall(res2.ApprovalID, "admin-bob", true)
	if err != nil || record.Status != agentgov.ApprovalStatusApproved {
		t.Fatalf("failed to approve HITL request: %v", err)
	}

	// Re-evaluate approved tool call -> ALLOW
	res2Retry := fw.EvaluateToolCall(chain, agentgov.ToolCallPayload{
		ToolName:  "payment_refund",
		Arguments: map[string]interface{}{"amount": 5000},
	})
	if res2Retry.Decision != agentgov.DecisionAllow {
		t.Fatalf("expected ALLOW after HITL approval, got %s", res2Retry.Decision)
	}

	// Test 4: Argument Transformation & Context Governance (TRANSFORM)
	res3 := fw.EvaluateToolCall(chain, agentgov.ToolCallPayload{
		ToolName:  "user_update",
		Arguments: map[string]interface{}{"username": "alice", "ssn": "123-45-6789"},
	})
	if res3.Decision != agentgov.DecisionTransform {
		t.Fatalf("expected TRANSFORM for sensitive SSN argument, got %s", res3.Decision)
	}
	if res3.TransformedArguments["ssn"] != "[REDACTED-BY-GATE-FIREWALL]" {
		t.Fatalf("expected ssn argument to be redacted, got %v", res3.TransformedArguments["ssn"])
	}

	// Test 5: Agent-to-Agent Delegation DENY for admin tool
	delegatedChain := chain
	delegatedChain.DelegatedBy = "agent-parent"
	res4 := fw.EvaluateToolCall(delegatedChain, agentgov.ToolCallPayload{
		ToolName:  "admin_system_shutdown",
		Arguments: map[string]interface{}{},
	})
	if res4.Decision != agentgov.DecisionDeny {
		t.Fatalf("expected DENY for delegated agent calling admin tool, got %s", res4.Decision)
	}
}

func TestTrajectorySimulationAndTrace(t *testing.T) {
	fw := agentgov.GetFirewall()
	report := fw.SimulateTrajectoryReplay()

	if _, exists := report["total_trajectories_replayed"]; !exists {
		t.Fatalf("invalid simulation report structure")
	}

	caps := fw.ProtocolAgnosticExposer()
	if caps["status"] != "active" {
		t.Fatalf("expected active status from ProtocolAgnosticExposer")
	}
}

func TestMCP_SecurityFirewallIntegration(t *testing.T) {
	mcpServer := mcp.NewMCPServer("pranor-gate-test", "1.0.0")
	mcpServer.RegisterTool(mcp.MCPTool{
		Name:        "system_info",
		Description: "Read system information",
	}, func(ctx context.Context, params map[string]interface{}) (mcp.MCPToolResult, error) {
		return mcp.MCPToolResult{Content: []mcp.MCPContent{{Type: "text", Text: "ok"}}}, nil
	})

	reqBody := `{"jsonrpc":"2.0","method":"tools/call","params":{"name":"system_info","arguments":{}},"id":1}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	mcpServer.Handler().ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for MCP request, got %d", resp.StatusCode)
	}

	var jsonResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&jsonResp)
	if jsonResp["result"] == nil {
		t.Fatalf("expected valid JSON-RPC result, got %v", jsonResp)
	}
}
