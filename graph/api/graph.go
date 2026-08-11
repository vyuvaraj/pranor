package api

import (
	"context"
	"errors"
)

type ContextTier int

const (
	TierHot ContextTier = iota
	TierWarm
	TierCold
)

type PruningStrategy string

const (
	PruningStrategyFIFO              PruningStrategy = "fifo"
	PruningStrategySemanticRelevance PruningStrategy = "semantic_relevance"
)

type ContextQuery struct {
	EntityID        string
	TenantID        string
	AgentID         string
	UserID          string
	RequestedTier   ContextTier
	MaxAgeSecs      int
	MaxTokenBudget  int             // max token budget for assembled context; 0 = unlimited
	PruningStrategy PruningStrategy // strategy for context node pruning
}

type ContextResult struct {
	EntityID        string
	TenantID        string
	Tier            ContextTier
	Data            map[string]any
	AssembledAtNs   int64
	LatencyMs       int64
	CacheHit        bool
	TokenCount      int // total token count of assembled context
	PrunedNodeCount int // number of nodes pruned to fit MaxTokenBudget
}

var ErrGraphContextUnavailable = errors.New("graph context unavailable")
var ErrEERequired = errors.New("enterprise edition required")

type GraphProvider interface {
	Query(ctx context.Context, q ContextQuery) (ContextResult, error)
	Invalidate(ctx context.Context, entityID, tenantID string) error
	HealthCheck(ctx context.Context) error
}
