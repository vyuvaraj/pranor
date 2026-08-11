package a2a

import (
	"context"
	"errors"

	"github.com/vyuvaraj/pranor/core/pkg/execctx"
)

var ErrCapabilityEscalationDenied = errors.New("pranor/agent/a2a: capability escalation denied")

type DelegationRequest struct {
	ChildAgentID          string
	SubTaskPayload        map[string]any
	RequestedCapabilities []string
	RiskBudget            float64
	TokenBudget           int
}

type DelegationResult struct {
	SessionID     string
	Status        string // SUCCESS, FAILED
	OutputPayload map[string]any
	TokensUsed    int
	CostUSD       float64
	ChildExecCtx  *execctx.ExecutionContext
}

type Delegator interface {
	Delegate(ctx context.Context, parentEC *execctx.ExecutionContext, req DelegationRequest) (DelegationResult, error)
}
