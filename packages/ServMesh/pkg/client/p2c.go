package client

import (
	"math/rand"
	"sync"
	"time"

	"github.com/vyuvaraj/serv/packages/ServMesh/pkg/registry"
)

// P2CLoadBalancer implements Power-of-Two-Choices (P2C) latency-weighted load balancing with locality preference.
type P2CLoadBalancer struct {
	mu            sync.RWMutex
	latencies     map[string]time.Duration // address -> EWMA latency
	localRegion   string
	localZone     string
	rng           *rand.Rand
}

// NewP2CLoadBalancer creates a P2CLoadBalancer with optional local region/zone awareness.
func NewP2CLoadBalancer(region, zone string) *P2CLoadBalancer {
	return &P2CLoadBalancer{
		latencies:   make(map[string]time.Duration),
		localRegion: region,
		localZone:   zone,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// RecordLatency updates the EWMA latency for a target address.
func (p *P2CLoadBalancer) RecordLatency(addr string, duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	old, ok := p.latencies[addr]
	if !ok || old == 0 {
		p.latencies[addr] = duration
	} else {
		// EWMA formula: 0.8 * old + 0.2 * sample
		p.latencies[addr] = time.Duration(0.8*float64(old) + 0.2*float64(duration))
	}
}

// SelectTarget picks the optimal target address using Power-of-Two-Choices.
// Preference hierarchy: same region > lowest EWMA latency.
func (p *P2CLoadBalancer) SelectTarget(instances []registry.Instance) string {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(instances) == 0 {
		return ""
	}
	if len(instances) == 1 {
		return instances[0].Address
	}

	// Filter for same-region matches
	var regionMatches []registry.Instance
	for _, inst := range instances {
		if p.localRegion != "" && inst.Region == p.localRegion {
			regionMatches = append(regionMatches, inst)
		}
	}

	candidates := instances
	if len(regionMatches) >= 2 {
		candidates = regionMatches
	}

	// 2. Pick two distinct random candidates from candidate pool
	idx1 := p.rng.Intn(len(candidates))
	idx2 := p.rng.Intn(len(candidates))
	if idx1 == idx2 {
		idx2 = (idx1 + 1) % len(candidates)
	}

	cand1 := candidates[idx1]
	cand2 := candidates[idx2]

	lat1 := p.latencies[cand1.Address]
	lat2 := p.latencies[cand2.Address]

	// Select candidate with lower EWMA latency
	if lat1 <= lat2 {
		return cand1.Address
	}
	return cand2.Address
}
