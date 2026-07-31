package import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// StampedeFetcher is a function provided by the caller to recompute or fetch an origin value.
type StampedeFetcher func() (interface{}, time.Duration, error)

// StampedeEntry holds a value alongside compute duration and absolute expiration.
type StampedeEntry struct {
	Value      interface{}
	Delta      time.Duration // Time taken to compute the value
	Expiration time.Time     // Expiration timestamp
}

// ProbabilisticEarlyExpiry (PER) protects against cache stampedes (thundering herd problem).
// It implements the XFetch algorithm (Vattani et al.):
//
//	early_expire = currentTime - delta * beta * ln(rnd()) > expiration
//
// Where:
//   - delta: duration taken to compute the item
//   - beta: aggression constant > 0 (default 1.0)
//   - rnd(): uniform random float in (0, 1]
type ProbabilisticEarlyExpiry struct {
	mu     sync.RWMutex
	cache  Cache
	entries map[string]StampedeEntry
	beta   float64
	sf     Group
}

// NewProbabilisticEarlyExpiry creates a new stampede protection wrapper around an existing Cache.
func NewProbabilisticEarlyExpiry(c Cache, beta float64) *ProbabilisticEarlyExpiry {
	if beta <= 0 {
		beta = 1.0
	}
	return &ProbabilisticEarlyExpiry{
		cache:   c,
		entries: make(map[string]StampedeEntry),
		beta:    beta,
	}
}

// ShouldRecompute calculates the XFetch probabilistic threshold.
// Returns true if the key should be recomputed early before hard expiration.
func (p *ProbabilisticEarlyExpiry) ShouldRecompute(entry StampedeEntry) bool {
	if entry.Expiration.IsZero() {
		return false
	}

	now := time.Now()
	if now.After(entry.Expiration) {
		return true // Hard expired
	}

	deltaSec := entry.Delta.Seconds()
	if deltaSec <= 0 {
		deltaSec = 0.01 // Minimum fallback computation time
	}

	rnd := rand.Float64()
	if rnd <= 0 {
		rnd = 0.0001
	}

	// XFetch formula: now - delta * beta * ln(rnd) > Expiration
	// Equivalent to: Expiration - now < -delta * beta * ln(rnd)
	timeToExpirySec := entry.Expiration.Sub(now).Seconds()
	earlyThresholdSec := -deltaSec * p.beta * math.Log(rnd)

	return timeToExpirySec < earlyThresholdSec
}

// GetOrFetch gets the value from cache, or recomputes it early if the XFetch check triggers.
// Uses singleflight coalescing so concurrent requests for an expired/early-recomputing key share one fetch call.
func (p *ProbabilisticEarlyExpiry) GetOrFetch(key string, fetcher StampedeFetcher) (interface{}, error) {
	p.mu.RLock()
	stEntry, hasMeta := p.entries[key]
	p.mu.RUnlock()

	val, found, err := p.cache.Get(key)
	if err == nil && found && hasMeta {
		// Key exists in cache, run PER probabilistic check
		if !p.ShouldRecompute(stEntry) {
			return val, nil
		}
	}

	// Key is either missing, hard expired, or selected for early recomputation.
	// Coalesce concurrent fetches using SingleFlight Group.
	res, fetchErr := p.sf.Do(key, func() (interface{}, error) {
		start := time.Now()
		newVal, ttl, err := fetcher()
		if err != nil {
			return nil, err
		}
		duration := time.Since(start)

		exp := time.Time{}
		if ttl > 0 {
			exp = time.Now().Add(ttl)
		}

		if err := p.cache.Set(key, newVal, ttl); err != nil {
			return nil, err
		}

		p.mu.Lock()
		p.entries[key] = StampedeEntry{
			Value:      newVal,
			Delta:      duration,
			Expiration: exp,
		}
		p.mu.Unlock()

		return newVal, nil
	})

	return res, fetchErr
}
