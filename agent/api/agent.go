package api

import (
	"context"
	"errors"
	"time"

	"github.com/vyuvaraj/pranor/core/pkg/execctx"
)

// AgentState tracks the runtime lifecycle state of an agent session.
type AgentState int

const (
	StateIdle AgentState = iota
	StateRunning
	StateWaitingTool
	StateWaitingHITL
	StateSuspended
	StateDone
	StateFailed
)

func (s AgentState) String() string {
	switch s {
	case StateIdle:
		return "IDLE"
	case StateRunning:
		return "RUNNING"
	case StateWaitingTool:
		return "WAITING_TOOL"
	case StateWaitingHITL:
		return "WAITING_HITL"
	case StateSuspended:
		return "SUSPENDED"
	case StateDone:
		return "DONE"
	case StateFailed:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}

// MemoryConfig configures agent memory allocation.
type MemoryConfig struct {
	EnableWorkingMemory  bool
	EnableEpisodicMemory bool
	EnableRAG            bool
	MaxEpisodicTurns     int
}

// BudgetConfig defines execution limits for the agent.
type BudgetConfig struct {
	MaxTokensPerRequest int
	MaxCostUSDPerDay    float64
	MaxSagaSteps        int
}

// AgentSpec is the declarative specification for an AI agent.
type AgentSpec struct {
	ID           string
	Name         string
	Version      string
	Description  string
	Capabilities []string // capability IDs from capability.Registry
	Memory       MemoryConfig
	Budget       BudgetConfig
	CreatedAt    time.Time
}

// AgentHandle represents an active agent instance bound to an ExecutionContext.
type AgentHandle struct {
	Spec      AgentSpec
	State     AgentState
	SessionID string
	ExecCtx   *execctx.ExecutionContext
	UpdatedAt time.Time
}

// AgentRegistry manages registration, lookup, and lifecycle of AgentSpecs and AgentHandles.
type AgentRegistry interface {
	Register(spec AgentSpec) error
	Lookup(agentID string) (AgentSpec, error)
	ListAll() []AgentSpec
	Spawn(ctx context.Context, ec *execctx.ExecutionContext, sessionID string) (*AgentHandle, error)
	UpdateState(handle *AgentHandle, state AgentState) error
	Suspend(handle *AgentHandle) error
	Resume(handle *AgentHandle) error
	Terminate(handle *AgentHandle, state AgentState) error
}

// Sentinel errors
var (
	ErrAgentNotFound      = errors.New("pranor/agent: agent spec not found")
	ErrAgentAlreadyExists = errors.New("pranor/agent: agent spec already registered")
	ErrInvalidAgentSpec   = errors.New("pranor/agent: invalid agent spec")
	ErrInvalidStateChange = errors.New("pranor/agent: invalid lifecycle state transition")
	ErrEERequired         = errors.New("pranor/agent: durable cross-cluster suspension requires Enterprise Edition")
)
