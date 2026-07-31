package security

import (
	"sync"
	"time"
)

// VelocityLimiter is a sliding-window rate limiter tracking failed login attempts
// per key (IP / username) to protect against credential stuffing attacks.
type VelocityLimiter struct {
	mu             sync.RWMutex
	windowDuration time.Duration
	maxAttempts    int
	blockDuration  time.Duration
	failures       map[string][]time.Time
	blockedUntil   map[string]time.Time
}

// NewVelocityLimiter creates a new VelocityLimiter. If windowDuration, maxAttempts, or blockDuration
// are <= 0, sensible defaults (1 min window, 5 attempts max, 15 min block) are applied.
func NewVelocityLimiter(windowDuration time.Duration, maxAttempts int, blockDuration time.Duration) *VelocityLimiter {
	if windowDuration <= 0 {
		windowDuration = 1 * time.Minute
	}
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	if blockDuration <= 0 {
		blockDuration = 15 * time.Minute
	}

	return &VelocityLimiter{
		windowDuration: windowDuration,
		maxAttempts:    maxAttempts,
		blockDuration:  blockDuration,
		failures:       make(map[string][]time.Time),
		blockedUntil:   make(map[string]time.Time),
	}
}

// RecordFailure registers a failed authentication attempt for the given key (IP or username).
func (vl *VelocityLimiter) RecordFailure(key string) {
	if key == "" {
		return
	}

	vl.mu.Lock()
	defer vl.mu.Unlock()

	now := time.Now()

	// If currently blocked, check if block duration has expired
	if blockedAt, ok := vl.blockedUntil[key]; ok {
		if now.Before(blockedAt) {
			// Already blocked
			return
		}
		// Block expired
		delete(vl.blockedUntil, key)
	}

	// Filter failures within sliding window
	cutoff := now.Add(-vl.windowDuration)
	var recent []time.Time
	for _, t := range vl.failures[key] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}

	recent = append(recent, now)

	if len(recent) >= vl.maxAttempts {
		vl.blockedUntil[key] = now.Add(vl.blockDuration)
		delete(vl.failures, key)
	} else {
		vl.failures[key] = recent
	}
}

// IsBlocked returns true if the key is currently blocked due to exceeding max attempts in the sliding window.
func (vl *VelocityLimiter) IsBlocked(key string) bool {
	if key == "" {
		return false
	}

	vl.mu.RLock()
	defer vl.mu.RUnlock()

	now := time.Now()

	if blockedAt, ok := vl.blockedUntil[key]; ok {
		if now.Before(blockedAt) {
			return true
		}
	}

	cutoff := now.Add(-vl.windowDuration)
	count := 0
	for _, t := range vl.failures[key] {
		if t.After(cutoff) {
			count++
		}
	}

	return count >= vl.maxAttempts
}

// Reset clears all recorded failures and unblocks the key.
func (vl *VelocityLimiter) Reset(key string) {
	if key == "" {
		return
	}

	vl.mu.Lock()
	defer vl.mu.Unlock()

	delete(vl.failures, key)
	delete(vl.blockedUntil, key)
}

// GetWindowDuration returns the configured sliding-window duration.
func (vl *VelocityLimiter) GetWindowDuration() time.Duration {
	return vl.windowDuration
}

// GetMaxAttempts returns the max attempts threshold.
func (vl *VelocityLimiter) GetMaxAttempts() int {
	return vl.maxAttempts
}

// GetBlockDuration returns the configured block duration.
func (vl *VelocityLimiter) GetBlockDuration() time.Duration {
	return vl.blockDuration
}
