//go:build !enterprise

package graphctx

import (
	"context"
	"sync"

	"github.com/vyuvaraj/pranor/graph/api"
)

type ThreeTierAssembler struct {
	mu       sync.RWMutex
	hotCache map[string]api.ContextResult
}

func NewThreeTierAssembler() *ThreeTierAssembler {
	return &ThreeTierAssembler{
		hotCache: make(map[string]api.ContextResult),
	}
}

func (a *ThreeTierAssembler) Assemble(ctx context.Context, q api.ContextQuery) (api.ContextResult, error) {
	// Try Hot Tier
	a.mu.RLock()
	res, ok := a.hotCache[q.EntityID+":"+q.TenantID]
	a.mu.RUnlock()

	if ok {
		res.CacheHit = true
		res.Tier = api.TierHot
		res.LatencyMs = 0
		return res, nil
	}

	// Try Warm Tier (Stub)
	if q.RequestedTier == api.TierHot {
		return api.ContextResult{}, api.ErrGraphContextUnavailable
	}

	if q.RequestedTier == api.TierWarm {
		return api.ContextResult{
			EntityID: q.EntityID,
			TenantID: q.TenantID,
			Tier:     api.TierWarm,
			Data:     make(map[string]any),
		}, nil
	}

	// Cold tier / Exhausted
	return api.ContextResult{}, api.ErrGraphContextUnavailable
}

// Helper method for testing hot tier
func (a *ThreeTierAssembler) SetHot(entityID, tenantID string, res api.ContextResult) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.hotCache[entityID+":"+tenantID] = res
}
