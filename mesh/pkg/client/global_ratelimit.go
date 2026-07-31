package client

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// PranorCacheStoreClient defines the interface for interacting with Pranor Cache storage.
type PranorCacheStoreClient interface {
	Get(key string) (interface{}, bool, error)
	Set(key string, value interface{}, ttl time.Duration) error
}

// GlobalRateLimiter enforces distributed token-bucket rate limits synchronized via Pranor Cache.
type GlobalRateLimiter struct {
	mu           sync.RWMutex
	cache        PranorCacheStoreClient
	maxTokens    int
	refillRateMs int
}

// NewGlobalRateLimiter creates a GlobalRateLimiter instance.
func NewGlobalRateLimiter(c PranorCacheStoreClient, maxTokens int, refillRateMs int) *GlobalRateLimiter {
	if maxTokens <= 0 {
		maxTokens = 100
	}
	if refillRateMs <= 0 {
		refillRateMs = 1000 // 1 second
	}
	return &GlobalRateLimiter{
		cache:        c,
		maxTokens:    maxTokens,
		refillRateMs: refillRateMs,
	}
}

// AllowChecks token bucket capacity for a key across distributed mesh nodes.
func (grl *GlobalRateLimiter) Allow(ctx context.Context, pairKey string) (bool, error) {
	if grl.cache == nil {
		return true, nil // Fallback to allow if cache unconfigured
	}

	grl.mu.Lock()
	defer grl.mu.Unlock()

	cacheKey := fmt.Sprintf("ratelimit:%s", pairKey)
	val, found, err := grl.cache.Get(cacheKey)
	if err != nil {
		return true, nil // Soft fail: allow traffic on cache query error
	}

	var currentTokens int
	if !found {
		currentTokens = grl.maxTokens
	} else {
		switch v := val.(type) {
		case int:
			currentTokens = v
		case int64:
			currentTokens = int(v)
		case string:
			currentTokens, _ = strconv.Atoi(v)
		default:
			currentTokens = grl.maxTokens
		}
	}

	if currentTokens <= 0 {
		return false, nil // Limit exceeded
	}

	// Consume 1 token and update Pranor Cache with TTL
	newTokenCount := currentTokens - 1
	ttl := time.Duration(grl.refillRateMs) * time.Millisecond
	_ = grl.cache.Set(cacheKey, newTokenCount, ttl)

	return true, nil
}

// Middleware returns an HTTP middleware enforcing global rate limiting on inter-service traffic.
func (grl *GlobalRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		caller := r.Header.Get("X-Caller-Id")
		if caller == "" {
			caller = "anonymous"
		}
		pairKey := fmt.Sprintf("%s->%s", caller, r.URL.Path)

		allowed, err := grl.Allow(r.Context(), pairKey)
		if err == nil && !allowed {
			http.Error(w, "global rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
