//go:build !enterprise

package graphctx

import (
	"context"
	"encoding/json"
	"log"
	"sort"
	"strings"
	"sync"

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
		return a.finalize(res, q), nil
	}

	// Record cache miss span / telemetry when hot cache misses
	a.recordMissTelemetry("hot_cache_miss", q)

	// Try Warm Tier
	warmRes, err := a.fetchWarmTier(ctx, q)
	if err == nil {
		warmRes.Tier = api.TierWarm
		warmRes.CacheHit = false
		return a.finalize(warmRes, q), nil
	}

	a.recordMissTelemetry("warm_cache_miss", q)

	// Try Cold Tier
	coldRes, err := a.fetchColdTier(ctx, q)
	if err == nil {
		coldRes.Tier = api.TierCold
		coldRes.CacheHit = false
		return a.finalize(coldRes, q), nil
	}

	a.recordMissTelemetry("cold_cache_miss", q)

	// Fail-closed: ensure api.ErrGraphContextUnavailable is strictly returned when all tiers are exhausted
	return api.ContextResult{}, api.ErrGraphContextUnavailable
}

func (a *ThreeTierAssembler) finalize(res api.ContextResult, q api.ContextQuery) api.ContextResult {
	if res.Data != nil {
		dataCopy := make(map[string]any)
		for k, v := range res.Data {
			dataCopy[k] = v
		}
		res.Data = dataCopy
	}

	if res.Data != nil {
		b, _ := json.Marshal(res.Data)
		res.TokenCount = len(b) / 4

		if q.MaxTokenBudget > 0 && res.TokenCount > q.MaxTokenBudget {
			keys := make([]string, 0, len(res.Data))
			for k := range res.Data {
				keys = append(keys, k)
			}

			if q.PruningStrategy == api.PruningStrategySemanticRelevance {
				rank := func(k string) int {
					kl := strings.ToLower(k)
					if strings.Contains(kl, "debug") || strings.Contains(kl, "raw") || strings.Contains(kl, "history") {
						return 0
					}
					if strings.Contains(kl, "user") || strings.Contains(kl, "intent") || strings.Contains(kl, "core") || strings.Contains(kl, "account") {
						return 2
					}
					return 1
				}
				sort.Slice(keys, func(i, j int) bool {
					ri := rank(keys[i])
					rj := rank(keys[j])
					if ri == rj {
						return keys[i] < keys[j]
					}
					return ri < rj
				})
			} else {
				sort.Strings(keys)
			}

			pruned := 0
			for _, k := range keys {
				delete(res.Data, k)
				pruned++

				b, _ := json.Marshal(res.Data)
				res.TokenCount = len(b) / 4
				if res.TokenCount <= q.MaxTokenBudget {
					break
				}
			}
			res.PrunedNodeCount = pruned
		}
	}
	return res
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
