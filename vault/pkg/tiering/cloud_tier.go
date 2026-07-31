package import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ColdTargetProvider string

const (
	AWSGlacierDeepArchive ColdTargetProvider = "AWS_GLACIER"
	AzureBlobArchive      ColdTargetProvider = "AZURE_BLOB_ARCHIVE"
	GCSColdline           ColdTargetProvider = "GCS_COLDLINE"
)

type MultiCloudTierPolicy struct {
	Bucket           string             `json:"bucket"`
	DaysBeforeArchiving int             `json:"days_before_archiving"`
	TargetProvider   ColdTargetProvider `json:"target_provider"`
	TargetBucket     string             `json:"target_bucket"`
}

type CloudTierEngine struct {
	policies map[string]MultiCloudTierPolicy
	mu       sync.RWMutex
	moved    uint64
}

func NewCloudTierEngine() *CloudTierEngine {
	return &CloudTierEngine{
		policies: make(map[string]MultiCloudTierPolicy),
	}
}

func (c *CloudTierEngine) SetPolicy(policy MultiCloudTierPolicy) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if policy.Bucket == "" || policy.TargetBucket == "" {
		return fmt.Errorf("cloud tier: missing bucket name or target bucket")
	}

	c.policies[policy.Bucket] = policy
	return nil
}

func (c *CloudTierEngine) EvaluateAndMigrate(ctx context.Context, bucket, key string, objectAgeDays int) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	pol, exists := c.policies[bucket]
	if !exists {
		return false, nil
	}

	if objectAgeDays >= pol.DaysBeforeArchiving {
		c.moved++
		return true, nil
	}
	return false, nil
}

func (c *CloudTierEngine) GetTieringStats() (uint64, time.Time) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.moved, time.Now()
}
