package api

import (
	"context"
	"errors"
)

type DecisionAction string

const (
	ActionDeny      DecisionAction = "DENY"
	ActionApprove   DecisionAction = "APPROVE"
	ActionTransform DecisionAction = "TRANSFORM"
)

type DecisionResult struct {
	Action        DecisionAction
	Reason        string
	PriorityLevel int
	LatencyMs     int64
}

type DecisionRequest struct {
	AgentID    string
	UserID     string
	TenantID   string
	RequestID  string
	Capability string
	Parameters map[string]any
	Context    map[string]any
}

const (
	PriorityAuth    = 1 // Authorization check - Hard DENY
	PriorityBudget  = 2 // Cost & Blast Radius - Hard DENY
	PriorityRisk    = 3 // Risk Engine - APPROVE/DENY
	PriorityRules   = 4 // Custom Rules - APPROVE/DENY/TRANSFORM
	PriorityLearn   = 5 // ML Advice - Soft Advisory
	PriorityDefault = 6 // Default policy
)

type DecisionEngine interface {
	Evaluate(ctx context.Context, req DecisionRequest) (DecisionResult, error)
}

var (
	ErrDecisionDenied     = errors.New("decision denied")
	ErrContextUnavailable = errors.New("context unavailable")
	ErrEERequired         = errors.New("enterprise edition required")
)
