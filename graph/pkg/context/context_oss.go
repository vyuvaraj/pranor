//go:build !enterprise

package graphctx

import (
	"context"
	"sync"
	"log"

	"github.com/vyuvaraj/pranor/graph/api"
)

type ThreeTierAssembler struct {
	mu           sync.RWMutex
	hotCache     map[string]api.ContextResult
	missSpanLog  []string
}

func NewThreeTierAssembler() *ThreeTierAssembler {
	return &ThreeTierAssembler{
		hotCache:    make(map[string]api.ContextResult),
		missSpanLog: make([]string, 0),
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

	// Record cache miss span / telemetry when hot cache misses
	a.recordMissTelemetry("hot_cache_miss", q)

	// Try Warm Tier
	warmRes, err := a.fetchWarmTier(ctx, q)
	if err == nil {
		warmRes.Tier = api.TierWarm
		warmRes.CacheHit = false
		return warmRes, nil
	}

	a.recordMissTelemetry("warm_cache_miss", q)

	// Try Cold Tier
	coldRes, err := a.fetchColdTier(ctx, q)
	if err == nil {
		coldRes.Tier = api.TierCold
		coldRes.CacheHit = false
		return coldRes, nil
	}

	a.recordMissTelemetry("cold_cache_miss", q)

	// Fail-closed: ensure api.ErrGraphContextUnavailable is strictly returned when all tiers are exhausted
	return api.ContextResult{}, api.ErrGraphContextUnavailable
}

func (a *ThreeTierAssembler) fetchWarmTier(ctx context.Context, q api.ContextQuery) (api.ContextResult, error) {
	if q.RequestedTier == api.TierWarm {
		return api.ContextResult{
			EntityID: q.EntityID,
			TenantID: q.TenantID,
			Tier:     api.TierWarm,
			Data:     make(map[string]any),
		}, nil
	}
	return api.ContextResult{}, api.ErrGraphContextUnavailable
}

func (a *ThreeTierAssembler) fetchColdTier(ctx context.Context, q api.ContextQuery) (api.ContextResult, error) {
	// OSS stub for cold tier
	return api.ContextResult{}, api.ErrGraphContextUnavailable
}

func (a *ThreeTierAssembler) recordMissTelemetry(event string, q api.ContextQuery) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.missSpanLog = append(a.missSpanLog, event+":"+q.EntityID+":"+q.TenantID)
	log.Printf("Telemetry: %s for entity %s, tenant %s", event, q.EntityID, q.TenantID)
}

// Helper method for testing hot tier
func (a *ThreeTierAssembler) SetHot(entityID, tenantID string, res api.ContextResult) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.hotCache[entityID+":"+tenantID] = res
}

// GetMissSpanLog returns recorded telemetry spans for testing
func (a *ThreeTierAssembler) GetMissSpanLog() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	logCopy := make([]string, len(a.missSpanLog))
	copy(logCopy, a.missSpanLog)
	return logCopy
}
