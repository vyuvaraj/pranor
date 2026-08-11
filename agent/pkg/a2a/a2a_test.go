package a2a

import (
	"context"
	"testing"

	"github.com/vyuvaraj/pranor/agent/api"
	"github.com/vyuvaraj/pranor/core/pkg/execctx"
)

type mockAgentRegistry struct {
	api.AgentRegistry
	spawnCalled bool
}

func (m *mockAgentRegistry) Spawn(ctx context.Context, ec *execctx.ExecutionContext, sessionID string) (*api.AgentHandle, error) {
	m.spawnCalled = true
	return &api.AgentHandle{
		SessionID: sessionID,
		ExecCtx:   ec,
	}, nil
}

func TestOSSDelegator_CapabilityEscalation(t *testing.T) {
	reg := &mockAgentRegistry{}
	delegator := NewOSSDelegator(reg)

	parentEC := execctx.New(context.Background(), "t1", "parent-agent", "u1")
	parentEC = parentEC.WithCapability("read")

	req := DelegationRequest{
		ChildAgentID:          "child-agent",
		RequestedCapabilities: []string{"read", "write"},
	}

	_, err := delegator.Delegate(context.Background(), parentEC, req)
	if err != ErrCapabilityEscalationDenied {
		t.Fatalf("expected ErrCapabilityEscalationDenied, got %v", err)
	}
}

func TestOSSDelegator_Success(t *testing.T) {
	reg := &mockAgentRegistry{}
	delegator := NewOSSDelegator(reg)

	parentEC := execctx.New(context.Background(), "t1", "parent-agent", "u1")
	parentEC = parentEC.WithCapability("read")

	req := DelegationRequest{
		ChildAgentID:          "child-agent",
		RequestedCapabilities: []string{"read"},
		RiskBudget:            0.5,
		TokenBudget:           1000,
	}

	res, err := delegator.Delegate(context.Background(), parentEC, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !reg.spawnCalled {
		t.Fatal("expected registry Spawn to be called")
	}

	if res.ChildExecCtx == nil {
		t.Fatal("expected ChildExecCtx to be returned")
	}

	if res.ChildExecCtx.AgentID != "child-agent" {
		t.Errorf("expected child agent ID 'child-agent', got '%s'", res.ChildExecCtx.AgentID)
	}

	if res.ChildExecCtx.ParentAgentID != "parent-agent" {
		t.Errorf("expected parent agent ID 'parent-agent', got '%s'", res.ChildExecCtx.ParentAgentID)
	}
}
