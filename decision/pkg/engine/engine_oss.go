//go:build !enterprise

package engine

import (
	"context"

	"github.com/vyuvaraj/pranor/decision/api"
	graphapi "github.com/vyuvaraj/pranor/graph/api"
)

func (e *VetoLadderEngine) Evaluate(ctx context.Context, req api.DecisionRequest) (api.DecisionResult, error) {
	// 1. Context loading from Graph
	if e.graphProvider != nil {
		q := graphapi.ContextQuery{
			EntityID:      req.AgentID,
			TenantID:      req.TenantID,
			AgentID:       req.AgentID,
			UserID:        req.UserID,
			RequestedTier: graphapi.TierHot,
		}
		_, err := e.graphProvider.Query(ctx, q)
		if err != nil {
			if err == graphapi.ErrGraphContextUnavailable {
				return api.DecisionResult{
					Action:        api.ActionDeny,
					Reason:        "graph context unavailable",
					PriorityLevel: api.PriorityAuth,
				}, api.ErrContextUnavailable
			}
			return api.DecisionResult{
				Action:        api.ActionDeny,
				Reason:        "graph context error",
				PriorityLevel: api.PriorityAuth,
			}, err
		}
	}

	// Priority 1: Auth (Hard DENY)
	if req.Capability == "FORBIDDEN_AUTH" {
		res := api.DecisionResult{
			Action:        api.ActionDeny,
			Reason:        "auth denied",
			PriorityLevel: api.PriorityAuth,
		}
		if req.SimulationMode {
			res.Reason = "simulation: " + res.Reason
			return res, nil
		}
		return res, api.ErrDecisionDenied
	}

	// Priority 2: Budget (Hard DENY)
	if req.Capability == "EXCEEDS_BUDGET" {
		res := api.DecisionResult{
			Action:        api.ActionDeny,
			Reason:        "budget exceeded",
			PriorityLevel: api.PriorityBudget,
		}
		if req.SimulationMode {
			res.Reason = "simulation: " + res.Reason
			return res, nil
		}
		return res, api.ErrDecisionDenied
	}

	// Priority 3: Risk Engine (APPROVE/DENY)
	if req.Capability == "HIGH_RISK" {
		res := api.DecisionResult{
			Action:        api.ActionDeny,
			Reason:        "risk too high",
			PriorityLevel: api.PriorityRisk,
		}
		if req.SimulationMode {
			res.Reason = "simulation: " + res.Reason
			return res, nil
		}
		return res, api.ErrDecisionDenied
	}

	// Priority 4: Custom Rules
	if req.Capability == "NEEDS_TRANSFORM" {
		res := api.DecisionResult{
			Action:        api.ActionTransform,
			Reason:        "transformed by rules",
			PriorityLevel: api.PriorityRules,
		}
		if req.SimulationMode {
			res.Reason = "simulation: " + res.Reason
		}
		return res, nil
	}

	// Priority 5: Learn (ML Advice - Soft Advisory)
	// Soft advisory: if it fails or is unavailable, we skip it.
	// In OSS, we don't do anything for ML Advice, it just passes through.

	// Priority 6: Default Policy
	res := api.DecisionResult{
		Action:        api.ActionApprove,
		Reason:        "default approve",
		PriorityLevel: api.PriorityDefault,
	}
	if req.SimulationMode {
		res.Reason = "simulation: " + res.Reason
	}
	return res, nil
}
