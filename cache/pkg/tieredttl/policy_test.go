package tieredttl_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/vyuvaraj/pranor/cache/pkg/cache"
	"github.com/vyuvaraj/pranor/cache/pkg/tieredttl"
)

func TestTierPolicy_ClassificationAndNaming(t *testing.T) {
	policy := tieredttl.NewTierPolicy()

	tests := []struct {
		ttl          time.Duration
		expectedTier tieredttl.Tier
		expectedName string
	}{
		{100 * time.Millisecond, tieredttl.TierHot, "Hot"},
		{1 * time.Second, tieredttl.TierHot, "Hot"},
		{1*time.Second + 1*time.Millisecond, tieredttl.TierWarm, "Warm"},
		{1 * time.Minute, tieredttl.TierWarm, "Warm"},
		{5 * time.Minute, tieredttl.TierWarm, "Warm"},
		{5*time.Minute + 1*time.Millisecond, tieredttl.TierCold, "Cold"},
		{1 * time.Hour, tieredttl.TierCold, "Cold"},
	}

	for _, tt := range tests {
		tier := policy.Classify(tt.ttl)
		if tier != tt.expectedTier {
			t.Errorf("Classify(%v) = %v; want %v", tt.ttl, tier, tt.expectedTier)
		}

		name := policy.TierName(tier)
		if name != tt.expectedName {
			t.Errorf("TierName(%v) = %v; want %v", tier, name, tt.expectedName)
		}

		// Also check package-level helpers
		if tieredttl.Classify(tt.ttl) != tt.expectedTier {
			t.Errorf("tieredttl.Classify(%v) = %v; want %v", tt.ttl, tieredttl.Classify(tt.ttl), tt.expectedTier)
		}
		if tieredttl.TierName(tier) != tt.expectedName {
			t.Errorf("tieredttl.TierName(%v) = %v; want %v", tier, tieredttl.TierName(tier), tt.expectedName)
		}
	}
}

func TestTieredCache_HitAndMissCounters(t *testing.T) {
	memCache := cache.NewInMemoryCache(100 * time.Millisecond)
	tc := tieredttl.NewTieredCache(memCache, nil)

	// Set Hot item
	err := tc.Set("hot_key", "val_hot", 500*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to Set hot_key: %v", err)
	}
	tier, ok := tc.GetTier("hot_key")
	if !ok || tier != tieredttl.TierHot {
		t.Fatalf("expected hot_key tier to be TierHot, got %v, ok=%v", tier, ok)
	}

	// Set Warm item
	err = tc.Set("warm_key", "val_warm", 2*time.Minute)
	if err != nil {
		t.Fatalf("failed to Set warm_key: %v", err)
	}

	// Set Cold item
	err = tc.Set("cold_key", "val_cold", 10*time.Minute)
	if err != nil {
		t.Fatalf("failed to Set cold_key: %v", err)
	}

	// Execute Gets for Hits
	_, found, err := tc.Get("hot_key")
	if !found || err != nil {
		t.Fatalf("expected hot_key to be found")
	}

	_, found, err = tc.Get("warm_key")
	if !found || err != nil {
		t.Fatalf("expected warm_key to be found")
	}
	_, found, err = tc.Get("warm_key")
	if !found || err != nil {
		t.Fatalf("expected warm_key to be found second time")
	}

	_, found, err = tc.Get("cold_key")
	if !found || err != nil {
		t.Fatalf("expected cold_key to be found")
	}

	// Execute Gets for Misses
	_, found, err = tc.Get("non_existent_key")
	if found {
		t.Fatalf("expected non_existent_key to not be found")
	}

	stats := tc.Stats()
	if stats.HotHits != 1 {
		t.Errorf("expected HotHits == 1, got %d", stats.HotHits)
	}
	if stats.WarmHits != 2 {
		t.Errorf("expected WarmHits == 2, got %d", stats.WarmHits)
	}
	if stats.ColdHits != 1 {
		t.Errorf("expected ColdHits == 1, got %d", stats.ColdHits)
	}
	if stats.ColdMisses != 1 {
		t.Errorf("expected ColdMisses == 1, got %d", stats.ColdMisses)
	}
}

func TestTieredCache_ExpiryAndMisses(t *testing.T) {
	memCache := cache.NewInMemoryCache(10 * time.Millisecond)
	tc := tieredttl.NewTieredCache(memCache, nil)

	// Set a Hot item with very short TTL
	tc.Set("short_hot", "value", 30*time.Millisecond)

	// Wait for expiration
	time.Sleep(50 * time.Millisecond)

	_, found, _ := tc.Get("short_hot")
	if found {
		t.Fatalf("expected short_hot to have expired")
	}

	stats := tc.Stats()
	if stats.HotMisses != 1 {
		t.Errorf("expected HotMisses == 1 after expiration, got %d", stats.HotMisses)
	}
}

func TestTieredCache_DeleteAndClear(t *testing.T) {
	memCache := cache.NewInMemoryCache(100 * time.Millisecond)
	tc := tieredttl.NewTieredCache(memCache, nil)

	tc.Set("k1", "v1", 500*time.Millisecond)
	tc.Set("k2", "v2", 2*time.Minute)

	if err := tc.Delete("k1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, ok := tc.GetTier("k1"); ok {
		t.Fatalf("expected k1 to be removed from GetTier after Delete")
	}

	if err := tc.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}
	if _, ok := tc.GetTier("k2"); ok {
		t.Fatalf("expected k2 to be removed from GetTier after Clear")
	}
}

func TestTieredCache_DeletePattern(t *testing.T) {
	memCache := cache.NewInMemoryCache(100 * time.Millisecond)
	tc := tieredttl.NewTieredCache(memCache, nil)

	tc.Set("user:101", "v1", 1*time.Minute)
	tc.Set("user:102", "v2", 1*time.Minute)
	tc.Set("item:201", "v3", 1*time.Minute)

	err := tc.DeletePattern("user:*")
	if err != nil {
		t.Fatalf("DeletePattern failed: %v", err)
	}

	if _, found, _ := tc.Get("user:101"); found {
		t.Errorf("expected user:101 to be deleted")
	}
	if _, found, _ := tc.Get("user:102"); found {
		t.Errorf("expected user:102 to be deleted")
	}
	if _, found, _ := tc.Get("item:201"); !found {
		t.Errorf("expected item:201 to remain")
	}
}

func TestTieredCache_Concurrency(t *testing.T) {
	memCache := cache.NewInMemoryCache(100 * time.Millisecond)
	tc := tieredttl.NewTieredCache(memCache, nil)

	var wg sync.WaitGroup
	workers := 10
	opsPerWorker := 50

	ttls := []time.Duration{
		500 * time.Millisecond,
		2 * time.Minute,
		10 * time.Minute,
	}

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < opsPerWorker; i++ {
				key := fmt.Sprintf("key_%d_%d", workerID, i)
				ttl := ttls[i%len(ttls)]

				_ = tc.Set(key, fmt.Sprintf("val_%d_%d", workerID, i), ttl)
				_, _, _ = tc.Get(key)
				_ = tc.Stats()
			}
		}(w)
	}

	wg.Wait()

	stats := tc.Stats()
	totalHits := stats.HotHits + stats.WarmHits + stats.ColdHits
	if totalHits == 0 {
		t.Errorf("expected total hits > 0 under concurrent operations")
	}
}
