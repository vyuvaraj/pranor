//go:build !enterprise

package a2a

import (
	"context"
	"fmt"
	"time"

	"github.com/vyuvaraj/pranor/agent/api"
	"github.com/vyuvaraj/pranor/core/pkg/execctx"
)

type ossDelegator struct {
	agentRegistry api.AgentRegistry
}

func NewOSSDelegator(reg api.AgentRegistry) Delegator {
	return &ossDelegator{agentRegistry: reg}
}

func (d *ossDelegator) Delegate(ctx context.Context, parentEC *execctx.ExecutionContext, req DelegationRequest) (DelegationResult, error) {
	for _, capID := range req.RequestedCapabilities {
		if !parentEC.HasCapability(capID) {
			return DelegationResult{Status: "FAILED"}, ErrCapabilityEscalationDenied
		}
	}

	childEC := parentEC.WithAgent(req.ChildAgentID)
	
	childEC.Capabilities = []string{}
	for _, capID := range req.RequestedCapabilities {
		childEC = childEC.WithCapability(capID)
	}

	childEC = childEC.WithBudget(req.RiskBudget, req.TokenBudget, parentEC.CostBudgetUS)

	sessionID := fmt.Sprintf("session-%d", time.Now().UnixNano())
	handle, err := d.agentRegistry.Spawn(ctx, childEC, sessionID)
	if err != nil {
		return DelegationResult{Status: "FAILED"}, err
	}

	return DelegationResult{
		SessionID:     handle.SessionID,
		Status:        "SUCCESS",
		OutputPayload: req.SubTaskPayload,
		TokensUsed:    0,
		CostUSD:       0.0,
		ChildExecCtx:  childEC,
	}, nil
}
