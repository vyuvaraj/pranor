package agent

import (
	"context"

	"github.com/vyuvaraj/pranor/agent/api"
	"github.com/vyuvaraj/pranor/agent/pkg/registry"
	"github.com/vyuvaraj/pranor/core/pkg/execctx"
)

// DefaultRegistry is the package-level AgentRegistry.
var DefaultRegistry api.AgentRegistry = registry.NewOSSRegistry()

// Register registers an AgentSpec in DefaultRegistry.
func Register(spec api.AgentSpec) error {
	return DefaultRegistry.Register(spec)
}

// Lookup retrieves an AgentSpec from DefaultRegistry.
func Lookup(agentID string) (api.AgentSpec, error) {
	return DefaultRegistry.Lookup(agentID)
}

// ListAll lists all registered AgentSpecs in DefaultRegistry.
func ListAll() []api.AgentSpec {
	return DefaultRegistry.ListAll()
}

// Spawn spawns an AgentHandle bound to an ExecutionContext.
func Spawn(ctx context.Context, ec *execctx.ExecutionContext, sessionID string) (*api.AgentHandle, error) {
	return DefaultRegistry.Spawn(ctx, ec, sessionID)
}

// Suspend suspends an active AgentHandle.
func Suspend(handle *api.AgentHandle) error {
	return DefaultRegistry.Suspend(handle)
}

// Resume resumes a suspended AgentHandle.
func Resume(handle *api.AgentHandle) error {
	return DefaultRegistry.Resume(handle)
}

// Terminate terminates an AgentHandle with a final state (DONE or FAILED).
func Terminate(handle *api.AgentHandle, state api.AgentState) error {
	return DefaultRegistry.Terminate(handle, state)
}
