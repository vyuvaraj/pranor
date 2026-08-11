package registry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vyuvaraj/pranor/agent/api"
	"github.com/vyuvaraj/pranor/core/pkg/execctx"
)

type ossRegistry struct {
	mu      sync.RWMutex
	specs   map[string]api.AgentSpec
	handles map[string]*api.AgentHandle
}

// NewOSSRegistry returns an in-memory AgentRegistry implementation for OSS.
func NewOSSRegistry() api.AgentRegistry {
	return &ossRegistry{
		specs:   make(map[string]api.AgentSpec),
		handles: make(map[string]*api.AgentHandle),
	}
}

func (r *ossRegistry) Register(spec api.AgentSpec) error {
	if spec.ID == "" {
		return fmt.Errorf("%w: ID is required", api.ErrInvalidAgentSpec)
	}
	if spec.Name == "" {
		return fmt.Errorf("%w: Name is required", api.ErrInvalidAgentSpec)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.specs[spec.ID]; exists {
		return api.ErrAgentAlreadyExists
	}

	spec.CreatedAt = time.Now().UTC()
	r.specs[spec.ID] = spec
	return nil
}

func (r *ossRegistry) Lookup(agentID string) (api.AgentSpec, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	spec, exists := r.specs[agentID]
	if !exists {
		return api.AgentSpec{}, api.ErrAgentNotFound
	}
	return spec, nil
}

func (r *ossRegistry) ListAll() []api.AgentSpec {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]api.AgentSpec, 0, len(r.specs))
	for _, s := range r.specs {
		list = append(list, s)
	}
	return list
}

func (r *ossRegistry) Spawn(ctx context.Context, ec *execctx.ExecutionContext, sessionID string) (*api.AgentHandle, error) {
	if err := ec.Validate(); err != nil {
		return nil, err
	}

	spec, err := r.Lookup(ec.AgentID)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	handle := &api.AgentHandle{
		Spec:      spec,
		State:     api.StateRunning,
		SessionID: sessionID,
		ExecCtx:   ec,
		UpdatedAt: now,
	}

	r.handles[sessionID] = handle
	return handle, nil
}

func (r *ossRegistry) UpdateState(handle *api.AgentHandle, newState api.AgentState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if handle.State == api.StateDone || handle.State == api.StateFailed {
		return api.ErrInvalidStateChange
	}

	handle.State = newState
	handle.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *ossRegistry) Suspend(handle *api.AgentHandle) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if handle.State != api.StateRunning && handle.State != api.StateWaitingTool && handle.State != api.StateWaitingHITL {
		return api.ErrInvalidStateChange
	}

	handle.State = api.StateSuspended
	handle.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *ossRegistry) Resume(handle *api.AgentHandle) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if handle.State != api.StateSuspended {
		return api.ErrInvalidStateChange
	}

	handle.State = api.StateRunning
	handle.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *ossRegistry) Terminate(handle *api.AgentHandle, state api.AgentState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if state != api.StateDone && state != api.StateFailed {
		return api.ErrInvalidStateChange
	}

	handle.State = state
	handle.UpdatedAt = time.Now().UTC()
	return nil
}
