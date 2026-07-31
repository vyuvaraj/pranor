package import (
	"context"
	"fmt"
	"sync"
)

type ObjectTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ExpirationRule struct {
	ID             string      `json:"id"`
	Prefix         string      `json:"prefix"`
	Tags           []ObjectTag `json:"tags"`
	ExpirationDays int         `json:"expiration_days"`
}

type AutoPurgeEngine struct {
	rules      map[string][]ExpirationRule
	tags       map[string][]ObjectTag
	mu         sync.RWMutex
	purgedCount uint64
}

func NewAutoPurgeEngine() *AutoPurgeEngine {
	return &AutoPurgeEngine{
		rules: make(map[string][]ExpirationRule),
		tags:  make(map[string][]ObjectTag),
	}
}

func (a *AutoPurgeEngine) AddRule(bucket string, rule ExpirationRule) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if bucket == "" || rule.ID == "" {
		return fmt.Errorf("lifecycle: missing bucket or rule ID")
	}

	a.rules[bucket] = append(a.rules[bucket], rule)
	return nil
}

func (a *AutoPurgeEngine) SetObjectTags(bucket, key string, tags []ObjectTag) {
	a.mu.Lock()
	defer a.mu.Unlock()

	objKey := fmt.Sprintf("%s/%s", bucket, key)
	a.tags[objKey] = tags
}

func (a *AutoPurgeEngine) EvaluatePurge(ctx context.Context, bucket, key string, ageDays int) (bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	rulesList, exists := a.rules[bucket]
	if !exists {
		return false, nil
	}

	for _, rule := range rulesList {
		if ageDays >= rule.ExpirationDays {
			a.purgedCount++
			return true, nil
		}
	}
	return false, nil
}
