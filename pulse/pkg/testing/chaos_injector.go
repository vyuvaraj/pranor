package import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type ChaosMode string

const (
	ChaosModeLatency   ChaosMode = "latency"
	ChaosModePartition ChaosMode = "partition"
	ChaosModeCorruption ChaosMode = "corruption"
	ChaosModeCrash     ChaosMode = "crash"
)

type FaultRule struct {
	Mode         ChaosMode     `json:"mode"`
	TopicPattern string        `json:"topic_pattern"`
	Latency      time.Duration `json:"latency"`
	Probability  float64       `json:"probability"`
	Enabled      bool          `json:"enabled"`
}

type ChaosInjector struct {
	mu          sync.RWMutex
	rules       map[string]*FaultRule
	InjectedCount uint64
}

func NewChaosInjector() *ChaosInjector {
	return &ChaosInjector{
		rules: make(map[string]*FaultRule),
	}
}

func (c *ChaosInjector) AddRule(id string, rule *FaultRule) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if rule.Probability <= 0 {
		rule.Probability = 1.0
	}
	rule.Enabled = true
	c.rules[id] = rule
}

func (c *ChaosInjector) DisableRule(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if r, ok := c.rules[id]; ok {
		r.Enabled = false
	}
}

func (c *ChaosInjector) Intercept(ctx context.Context, topic string, payload string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, rule := range c.rules {
		if !rule.Enabled {
			continue
		}

		if rand.Float64() <= rule.Probability {
			c.mu.RUnlock()
			c.mu.Lock()
			c.InjectedCount++
			c.mu.Unlock()
			c.mu.RLock()

			switch rule.Mode {
			case ChaosModeLatency:
				time.Sleep(rule.Latency)
			case ChaosModePartition:
				return "", fmt.Errorf("chaos: simulated network partition for topic %s", topic)
			case ChaosModeCorruption:
				return payload + "_CORRUPTED_CHAOS_BITFLIP", nil
			case ChaosModeCrash:
				return "", fmt.Errorf("chaos: simulated node crash panic on topic %s", topic)
			}
		}
	}

	return payload, nil
}
