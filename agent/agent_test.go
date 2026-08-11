package agent_test

import (
	"context"
	"testing"

	"github.com/vyuvaraj/pranor/agent"
	"github.com/vyuvaraj/pranor/agent/api"
	"github.com/vyuvaraj/pranor/core/pkg/execctx"
)

func TestRegisterAndLookup(t *testing.T) {
	spec := api.AgentSpec{
		ID:          "agent-test-1",
		Name:        "Test Agent",
		Version:     "1.0.0",
		Description: "A test agent spec",
		Capabilities: []string{"pool.query"},
	}

	err := agent.Register(spec)
	if err != nil {
		t.Fatalf("unexpected error registering spec: %v", err)
	}

	found, err := agent.Lookup("agent-test-1")
	if err != nil {
		t.Fatalf("unexpected error looking up spec: %v", err)
	}
	if found.Name != spec.Name {
		t.Errorf("expected name %s, got %s", spec.Name, found.Name)
	}
}

func TestRegister_Duplicate(t *testing.T) {
	spec := api.AgentSpec{
		ID:   "agent-dup-1",
		Name: "Dup Agent",
	}

	_ = agent.Register(spec)
	err := agent.Register(spec)
	if err != api.ErrAgentAlreadyExists {
		t.Errorf("expected ErrAgentAlreadyExists, got %v", err)
	}
}

func TestSpawn(t *testing.T) {
	spec := api.AgentSpec{
		ID:   "agent-spawn-1",
		Name: "Spawn Agent",
	}
	_ = agent.Register(spec)

	ec := execctx.New(context.Background(), "tenant-1", "agent-spawn-1", "user-1")
	handle, err := agent.Spawn(context.Background(), ec, "session-123")
	if err != nil {
		t.Fatalf("unexpected error spawning agent: %v", err)
	}

	if handle.State != api.StateRunning {
		t.Errorf("expected state RUNNING, got %s", handle.State)
	}
	if handle.SessionID != "session-123" {
		t.Errorf("expected session-123, got %s", handle.SessionID)
	}
}

func TestLifecycleTransitions(t *testing.T) {
	spec := api.AgentSpec{
		ID:   "agent-life-1",
		Name: "Lifecycle Agent",
	}
	_ = agent.Register(spec)

	ec := execctx.New(context.Background(), "tenant-1", "agent-life-1", "user-1")
	handle, _ := agent.Spawn(context.Background(), ec, "session-life")

	// RUNNING -> WAITING_TOOL
	err := agent.DefaultRegistry.UpdateState(handle, api.StateWaitingTool)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// WAITING_TOOL -> RUNNING
	err = agent.DefaultRegistry.UpdateState(handle, api.StateRunning)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// RUNNING -> SUSPENDED via helper
	err = agent.Suspend(handle)
	if err != nil {
		t.Fatalf("unexpected error suspending: %v", err)
	}
	if handle.State != api.StateSuspended {
		t.Errorf("expected SUSPENDED, got %s", handle.State)
	}

	// SUSPENDED -> RUNNING via helper
	err = agent.Resume(handle)
	if err != nil {
		t.Fatalf("unexpected error resuming: %v", err)
	}

	// RUNNING -> DONE via Terminate
	err = agent.Terminate(handle, api.StateDone)
	if err != nil {
		t.Fatalf("unexpected error terminating: %v", err)
	}
	if handle.State != api.StateDone {
		t.Errorf("expected DONE, got %s", handle.State)
	}
}

func TestInvalidStateTransition(t *testing.T) {
	spec := api.AgentSpec{
		ID:   "agent-invalid-1",
		Name: "Invalid Agent",
	}
	_ = agent.Register(spec)

	ec := execctx.New(context.Background(), "tenant-1", "agent-invalid-1", "user-1")
	handle, _ := agent.Spawn(context.Background(), ec, "session-inv")
	_ = agent.Terminate(handle, api.StateDone)

	// Attempting transition from DONE -> RUNNING
	err := agent.DefaultRegistry.UpdateState(handle, api.StateRunning)
	if err != api.ErrInvalidStateChange {
		t.Errorf("expected ErrInvalidStateChange, got %v", err)
	}
}

func TestState_String(t *testing.T) {
	tests := []struct {
		state    api.AgentState
		expected string
	}{
		{api.StateIdle, "IDLE"},
		{api.StateRunning, "RUNNING"},
		{api.StateWaitingTool, "WAITING_TOOL"},
		{api.StateWaitingHITL, "WAITING_HITL"},
		{api.StateSuspended, "SUSPENDED"},
		{api.StateDone, "DONE"},
		{api.StateFailed, "FAILED"},
	}

	for _, tc := range tests {
		if tc.state.String() != tc.expected {
			t.Errorf("state %d: expected %s, got %s", tc.state, tc.expected, tc.state.String())
		}
	}
}
