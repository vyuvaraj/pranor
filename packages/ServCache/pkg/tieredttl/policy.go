package tieredttl

import (
	"path"
	"strings"
	"sync"
	"time"

	"github.com/vyuvaraj/serv/packages/ServCache/pkg/cache"
)

// Tier represents the classification of a cache item based on its TTL.
type Tier int

const (
	TierHot Tier = iota
	TierWarm
	TierCold
)

// TierPolicy defines classification rules for TTLs.
type TierPolicy struct{}

// NewTierPolicy constructs a new TierPolicy.
func NewTierPolicy() *TierPolicy {
	return &TierPolicy{}
}

// Classify categorizes a TTL into Hot (<=1s), Warm (<=5m), or Cold (>5m).
func (p *TierPolicy) Classify(ttl time.Duration) Tier {
	if ttl <= 1*time.Second {
		return TierHot
	}
	if ttl <= 5*time.Minute {
		return TierWarm
	}
	return TierCold
}

// TierName returns the human-readable string representation of a Tier.
func (p *TierPolicy) TierName(t Tier) string {
	switch t {
	case TierHot:
		return "Hot"
	case TierWarm:
		return "Warm"
	case TierCold:
		return "Cold"
	default:
		return "Unknown"
	}
}

// Classify is a package-level helper for classifying TTL.
func Classify(ttl time.Duration) Tier {
	return (&TierPolicy{}).Classify(ttl)
}

// TierName is a package-level helper for getting tier names.
func TierName(t Tier) string {
	return (&TierPolicy{}).TierName(t)
}

// TierStats holds hit and miss metrics broken down by tier.
type TierStats struct {
	HotHits    int64 `json:"hot_hits"`
	HotMisses  int64 `json:"hot_misses"`
	WarmHits   int64 `json:"warm_hits"`
	WarmMisses int64 `json:"warm_misses"`
	ColdHits   int64 `json:"cold_hits"`
	ColdMisses int64 `json:"cold_misses"`
}

// TieredCache wraps a cache instance and routes items based on TTL policy, tracking per-tier metrics.
type TieredCache struct {
	mu         sync.RWMutex
	underlying cache.Cache
	policy     *TierPolicy
	keyTiers   map[string]Tier
	stats      TierStats
}

// NewTieredCache wraps an existing Cache with TierPolicy tracking.
func NewTieredCache(underlying cache.Cache, policy *TierPolicy) *TieredCache {
	if underlying == nil {
		underlying = cache.NewInMemoryCache(100 * time.Millisecond)
	}
	if policy == nil {
		policy = NewTierPolicy()
	}
	return &TieredCache{
		underlying: underlying,
		policy:     policy,
		keyTiers:   make(map[string]Tier),
	}
}

// Set classifies the key's TTL, records its tier, and delegates to the underlying cache.
func (c *TieredCache) Set(key string, value interface{}, ttl time.Duration) error {
	tier := c.policy.Classify(ttl)

	c.mu.Lock()
	c.keyTiers[key] = tier
	c.mu.Unlock()

	return c.underlying.Set(key, value, ttl)
}

// Get fetches the key from the underlying cache and updates hit/miss stats per tier.
func (c *TieredCache) Get(key string) (interface{}, bool, error) {
	val, found, err := c.underlying.Get(key)

	c.mu.Lock()
	defer c.mu.Unlock()

	tier, exists := c.keyTiers[key]
	if !exists {
		tier = TierCold
	}

	if found {
		switch tier {
		case TierHot:
			c.stats.HotHits++
		case TierWarm:
			c.stats.WarmHits++
		case TierCold:
			c.stats.ColdHits++
		}
	} else {
		switch tier {
		case TierHot:
			c.stats.HotMisses++
		case TierWarm:
			c.stats.WarmMisses++
		case TierCold:
			c.stats.ColdMisses++
		}
		if exists {
			delete(c.keyTiers, key)
		}
	}

	return val, found, err
}

// Delete removes a key from keyTiers mapping and delegates to the underlying cache.
func (c *TieredCache) Delete(key string) error {
	c.mu.Lock()
	delete(c.keyTiers, key)
	c.mu.Unlock()

	return c.underlying.Delete(key)
}

// Clear clears all keyTiers mappings and delegates to the underlying cache.
func (c *TieredCache) Clear() error {
	c.mu.Lock()
	c.keyTiers = make(map[string]Tier)
	c.mu.Unlock()

	return c.underlying.Clear()
}

// DeletePattern deletes keys matching the pattern and delegates to the underlying cache.
func (c *TieredCache) DeletePattern(pattern string) error {
	c.mu.Lock()
	for k := range c.keyTiers {
		matched, err := path.Match(pattern, k)
		if (err == nil && matched) ||
			(strings.HasSuffix(pattern, "*") && strings.HasPrefix(k, strings.TrimSuffix(pattern, "*"))) ||
			k == pattern {
			delete(c.keyTiers, k)
		}
	}
	c.mu.Unlock()

	return c.underlying.DeletePattern(pattern)
}

// Stats returns a snapshot of current per-tier hit and miss statistics.
func (c *TieredCache) Stats() TierStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}

// GetTier returns the classified tier of a key if present in keyTiers.
func (c *TieredCache) GetTier(key string) (Tier, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	t, ok := c.keyTiers[key]
	return t, ok
}

// Underlying returns the wrapped Cache instance.
func (c *TieredCache) Underlying() cache.Cache {
	return c.underlying
}
