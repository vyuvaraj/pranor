package main_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vyuvaraj/pranor/cache/pkg/bloom"
	"github.com/vyuvaraj/pranor/cache/pkg/cache"
	"github.com/vyuvaraj/pranor/cache/pkg/tieredttl"
)

// ============================================================================
// SC.G3: Probabilistic Bloom Filter Tests
// ============================================================================

// --- Tier 1: Feature Coverage (SC.G3) ---

func TestE2E_SC_G3_Tier1_AddAndMayContain_SingleKey(t *testing.T) {
	b := bloom.NewBloom(1000, 0.01)
	key := "user:session:1001"

	b.Add(key)

	if !b.MayContain(key) {
		t.Errorf("expected MayContain(%q) to return true after Add", key)
	}
}

func TestE2E_SC_G3_Tier1_ZeroFalseNegatives_1000Items(t *testing.T) {
	n := 1000
	b := bloom.NewBloom(n, 0.01)

	keys := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = fmt.Sprintf("item_key_%d", i)
		b.Add(keys[i])
	}

	// Verify ZERO false negatives for all added items
	for i := 0; i < n; i++ {
		if !b.MayContain(keys[i]) {
			t.Fatalf("false negative detected for added key %q at index %d", keys[i], i)
		}
	}
}

func TestE2E_SC_G3_Tier1_FalsePositiveRate_WithinBound(t *testing.T) {
	capacity := 1000
	targetFpRate := 0.01
	b := bloom.NewBloom(capacity, targetFpRate)

	// Add 1000 items
	for i := 0; i < capacity; i++ {
		b.Add(fmt.Sprintf("added_key_%d", i))
	}

	// Check 1000 absent items
	falsePositives := 0
	numAbsent := 1000
	for i := 0; i < numAbsent; i++ {
		absentKey := fmt.Sprintf("absent_key_%d", i)
		if b.MayContain(absentKey) {
			falsePositives++
		}
	}

	actualRate := float64(falsePositives) / float64(numAbsent)
	// Allow a reasonable statistical tolerance bound (e.g. max 3x target rate for 1000 samples)
	maxAllowedRate := targetFpRate * 3.0
	if actualRate > maxAllowedRate {
		t.Errorf("false positive rate %.4f exceeded tolerance bound %.4f (%d/%d false positives)",
			actualRate, maxAllowedRate, falsePositives, numAbsent)
	}
}

func TestE2E_SC_G3_Tier1_AbsentKeys_MostlyFalse(t *testing.T) {
	b := bloom.NewBloom(1000, 0.01)
	b.Add("present_key_1")
	b.Add("present_key_2")

	absentKeys := []string{"nonexistent_a", "nonexistent_b", "nonexistent_c", "nonexistent_d"}
	falseCount := 0
	for _, key := range absentKeys {
		if !b.MayContain(key) {
			falseCount++
		}
	}

	if falseCount == 0 {
		t.Errorf("expected absent keys to return false, got MayContain=true for all absent keys")
	}
}

func TestE2E_SC_G3_Tier1_GetterMethods_Metadata(t *testing.T) {
	capacity := 5000
	fpRate := 0.005
	b := bloom.NewBloom(capacity, fpRate)

	if b.Capacity() != capacity {
		t.Errorf("expected Capacity() %d, got %d", capacity, b.Capacity())
	}
	if b.FalsePositiveRate() != fpRate {
		t.Errorf("expected FalsePositiveRate() %v, got %v", fpRate, b.FalsePositiveRate())
	}
	if b.M() == 0 {
		t.Errorf("expected non-zero total bits M()")
	}
	if b.K() == 0 {
		t.Errorf("expected non-zero hash functions count K()")
	}
}

// --- Tier 2: Boundary & Corner Cases (SC.G3) ---

func TestE2E_SC_G3_Tier2_InvalidConfig_Defaults(t *testing.T) {
	b := bloom.NewBloom(0, -1.0)
	if b.Capacity() != 1000 {
		t.Errorf("expected default capacity 1000 for invalid input, got %d", b.Capacity())
	}
	if b.FalsePositiveRate() != 0.01 {
		t.Errorf("expected default FP rate 0.01 for invalid input, got %v", b.FalsePositiveRate())
	}

	b2 := bloom.NewBloom(-500, 2.5)
	if b2.Capacity() != 1000 {
		t.Errorf("expected default capacity 1000 for negative input, got %d", b2.Capacity())
	}
}

func TestE2E_SC_G3_Tier2_EmptyKey_Handling(t *testing.T) {
	b := bloom.NewBloom(100, 0.01)
	emptyKey := ""

	b.Add(emptyKey)

	if !b.MayContain(emptyKey) {
		t.Errorf("expected empty key to return MayContain=true after Add")
	}
}

func TestE2E_SC_G3_Tier2_SpecialCharacters_BinaryKeys(t *testing.T) {
	b := bloom.NewBloom(100, 0.01)

	specialKeys := []string{
		"key\x00with\x00nulls",
		"🔥unicode_emoji_key🔥",
		strings.Repeat("very_long_key_", 500),
		"line1\nline2\r\ntab\tquote\"slash\\",
	}

	for _, k := range specialKeys {
		b.Add(k)
	}

	for _, k := range specialKeys {
		if !b.MayContain(k) {
			t.Errorf("false negative for special key: %q", k)
		}
	}
}

func TestE2E_SC_G3_Tier2_OverCapacity_Degradation(t *testing.T) {
	// Size filter for 50 items, but insert 250 items (500% capacity)
	b := bloom.NewBloom(50, 0.01)
	n := 250

	for i := 0; i < n; i++ {
		b.Add(fmt.Sprintf("over_cap_%d", i))
	}

	// Zero false negatives MUST still hold even when over capacity
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("over_cap_%d", i)
		if !b.MayContain(key) {
			t.Fatalf("false negative when over capacity for key %q", key)
		}
	}
}

func TestE2E_SC_G3_Tier2_ConcurrentAddAndCheck_ThreadSafety(t *testing.T) {
	b := bloom.NewBloom(2000, 0.01)
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				key := fmt.Sprintf("worker_%d_item_%d", workerID, j)
				b.Add(key)
				_ = b.MayContain(key)
			}
		}(i)
	}
	wg.Wait()
}

// ============================================================================
// SC.G4: Tiered TTL Policy Engine Tests
// ============================================================================

// --- Tier 1: Feature Coverage (SC.G4) ---

func TestE2E_SC_G4_Tier1_Classify_HotWarmCold(t *testing.T) {
	policy := tieredttl.NewTierPolicy()

	if p := policy.Classify(500 * time.Millisecond); p != tieredttl.TierHot {
		t.Errorf("expected 500ms to classify as Hot, got %v", p)
	}
	if p := policy.Classify(1 * time.Second); p != tieredttl.TierHot {
		t.Errorf("expected 1s to classify as Hot, got %v", p)
	}
	if p := policy.Classify(2 * time.Second); p != tieredttl.TierWarm {
		t.Errorf("expected 2s to classify as Warm, got %v", p)
	}
	if p := policy.Classify(5 * time.Minute); p != tieredttl.TierWarm {
		t.Errorf("expected 5m to classify as Warm, got %v", p)
	}
	if p := policy.Classify(10 * time.Minute); p != tieredttl.TierCold {
		t.Errorf("expected 10m to classify as Cold, got %v", p)
	}
}

func TestE2E_SC_G4_Tier1_TierName_Formatting(t *testing.T) {
	policy := tieredttl.NewTierPolicy()

	if name := policy.TierName(tieredttl.TierHot); name != "Hot" {
		t.Errorf("expected 'Hot', got %q", name)
	}
	if name := policy.TierName(tieredttl.TierWarm); name != "Warm" {
		t.Errorf("expected 'Warm', got %q", name)
	}
	if name := policy.TierName(tieredttl.TierCold); name != "Cold" {
		t.Errorf("expected 'Cold', got %q", name)
	}
	if name := policy.TierName(tieredttl.Tier(99)); name != "Unknown" {
		t.Errorf("expected 'Unknown', got %q", name)
	}
}

func TestE2E_SC_G4_Tier1_SetAndGet_HitStats(t *testing.T) {
	memCache := cache.NewInMemoryCache(100 * time.Millisecond)
	tiered := tieredttl.NewTieredCache(memCache, nil)

	// Set 1 Hot, 1 Warm, 1 Cold
	_ = tiered.Set("k_hot", "val_hot", 500*time.Millisecond)
	_ = tiered.Set("k_warm", "val_warm", 2*time.Minute)
	_ = tiered.Set("k_cold", "val_cold", 1*time.Hour)

	// Get all three
	_, foundHot, _ := tiered.Get("k_hot")
	_, foundWarm, _ := tiered.Get("k_warm")
	_, foundCold, _ := tiered.Get("k_cold")

	if !foundHot || !foundWarm || !foundCold {
		t.Fatalf("expected all keys to be found")
	}

	stats := tiered.Stats()
	if stats.HotHits != 1 {
		t.Errorf("expected HotHits 1, got %d", stats.HotHits)
	}
	if stats.WarmHits != 1 {
		t.Errorf("expected WarmHits 1, got %d", stats.WarmHits)
	}
	if stats.ColdHits != 1 {
		t.Errorf("expected ColdHits 1, got %d", stats.ColdHits)
	}
}

func TestE2E_SC_G4_Tier1_Get_MissStats(t *testing.T) {
	memCache := cache.NewInMemoryCache(100 * time.Millisecond)
	tiered := tieredttl.NewTieredCache(memCache, nil)

	// Register key as Hot tier
	_ = tiered.Set("k_hot_miss", "val", 100*time.Millisecond)

	// Delete from underlying to force a miss on recorded Hot key
	_ = memCache.Delete("k_hot_miss")

	_, found, _ := tiered.Get("k_hot_miss")
	if found {
		t.Fatalf("expected miss")
	}

	stats := tiered.Stats()
	if stats.HotMisses != 1 {
		t.Errorf("expected HotMisses 1, got %d", stats.HotMisses)
	}
}

func TestE2E_SC_G4_Tier1_DeleteAndClear_Behavior(t *testing.T) {
	memCache := cache.NewInMemoryCache(100 * time.Millisecond)
	tiered := tieredttl.NewTieredCache(memCache, nil)

	_ = tiered.Set("key1", "val1", 1*time.Second)
	_ = tiered.Set("key2", "val2", 1*time.Minute)

	if err := tiered.Delete("key1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, found1, _ := tiered.Get("key1")
	if found1 {
		t.Errorf("expected key1 to be deleted")
	}

	if err := tiered.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	_, found2, _ := tiered.Get("key2")
	if found2 {
		t.Errorf("expected key2 to be cleared")
	}
}

// --- Tier 2: Boundary & Corner Cases (SC.G4) ---

func TestE2E_SC_G4_Tier2_Boundary_1s_5m(t *testing.T) {
	policy := tieredttl.NewTierPolicy()

	// 1 second exactly -> Hot
	if policy.Classify(1*time.Second) != tieredttl.TierHot {
		t.Errorf("1s should be Hot")
	}
	// 1s + 1ms -> Warm
	if policy.Classify(1*time.Second+1*time.Millisecond) != tieredttl.TierWarm {
		t.Errorf("1s+1ms should be Warm")
	}
	// 5m exactly -> Warm
	if policy.Classify(5*time.Minute) != tieredttl.TierWarm {
		t.Errorf("5m should be Warm")
	}
	// 5m + 1ms -> Cold
	if policy.Classify(5*time.Minute+1*time.Millisecond) != tieredttl.TierCold {
		t.Errorf("5m+1ms should be Cold")
	}
}

func TestE2E_SC_G4_Tier2_NilUnderlyingOrPolicy_Defaults(t *testing.T) {
	tiered := tieredttl.NewTieredCache(nil, nil)
	if tiered.Underlying() == nil {
		t.Errorf("expected non-nil default underlying cache")
	}

	err := tiered.Set("default_key", "default_val", 5*time.Second)
	if err != nil {
		t.Errorf("Set failed on default TieredCache: %v", err)
	}
}

func TestE2E_SC_G4_Tier2_DeletePattern_Wildcards(t *testing.T) {
	memCache := cache.NewInMemoryCache(100 * time.Millisecond)
	tiered := tieredttl.NewTieredCache(memCache, nil)

	_ = tiered.Set("user:101", "alice", 1*time.Minute)
	_ = tiered.Set("user:102", "bob", 1*time.Minute)
	_ = tiered.Set("session:99", "active", 1*time.Minute)

	err := tiered.DeletePattern("user:*")
	if err != nil {
		t.Fatalf("DeletePattern failed: %v", err)
	}

	_, found1, _ := tiered.Get("user:101")
	_, found2, _ := tiered.Get("user:102")
	_, foundSess, _ := tiered.Get("session:99")

	if found1 || found2 {
		t.Errorf("expected user:* keys to be deleted")
	}
	if !foundSess {
		t.Errorf("expected session:99 to remain")
	}
}

func TestE2E_SC_G4_Tier2_Get_ExpiredItem_MissTracking(t *testing.T) {
	memCache := cache.NewInMemoryCache(10 * time.Millisecond)
	tiered := tieredttl.NewTieredCache(memCache, nil)

	_ = tiered.Set("expiring_hot", "val", 10*time.Millisecond)

	time.Sleep(30 * time.Millisecond)

	_, found, _ := tiered.Get("expiring_hot")
	if found {
		t.Fatalf("expected item to be expired and missed")
	}

	stats := tiered.Stats()
	if stats.HotMisses != 1 {
		t.Errorf("expected 1 HotMiss for expired Hot item, got %d", stats.HotMisses)
	}
}

func TestE2E_SC_G4_Tier2_NegativeTTL_Handling(t *testing.T) {
	policy := tieredttl.NewTierPolicy()
	tier := policy.Classify(-5 * time.Second)
	if tier != tieredttl.TierHot {
		t.Errorf("expected negative TTL to classify as Hot (<= 1s), got %v", tier)
	}
}

// ============================================================================
// Tier 3: Cross-Feature Combinations (SC.G3 + SC.G4)
// ============================================================================

func TestE2E_SC_Tier3_Cross_BloomAndTieredTTL(t *testing.T) {
	// Combination: Wrap TieredCache with Bloom filter front guard
	// Before accessing TieredCache.Get(key), check bloom.MayContain(key).
	// If false -> bypass cache completely (zero latency, zero miss stats).
	// If true -> check TieredCache.

	memCache := cache.NewInMemoryCache(100 * time.Millisecond)
	tiered := tieredttl.NewTieredCache(memCache, nil)
	filter := bloom.NewBloom(1000, 0.01)

	// Populate cache and bloom filter
	keys := []string{"session:alpha", "session:beta", "session:gamma"}
	for _, k := range keys {
		_ = tiered.Set(k, "active_data", 2*time.Minute)
		filter.Add(k)
	}

	// Helper function simulating Bloom + TieredCache query
	queryCache := func(k string) (interface{}, bool, bool) {
		// Return (value, found, bypassed)
		if !filter.MayContain(k) {
			return nil, false, true // Bypassed
		}
		val, found, _ := tiered.Get(k)
		return val, found, false
	}

	// Test existing key
	val, found, bypassed := queryCache("session:alpha")
	if bypassed || !found || val != "active_data" {
		t.Errorf("expected hit for session:alpha, got bypassed=%v found=%v val=%v", bypassed, found, val)
	}

	// Test absent key
	absentKey := "definitely_absent_key_99999"
	_, foundAbsent, bypassedAbsent := queryCache(absentKey)
	if !bypassedAbsent {
		// Bloom filter may have false positive, but for this distinct key it should bypass
		if foundAbsent {
			t.Errorf("absent key should not be found in cache")
		}
	}

	// Verify stats: only queried keys incremented miss counter
	stats := tiered.Stats()
	if stats.WarmHits != 1 {
		t.Errorf("expected 1 WarmHit for session:alpha, got %d", stats.WarmHits)
	}
}

// ============================================================================
// Tier 4: Real-World Application Scenarios (Pranor Cache)
// ============================================================================

func TestE2E_SC_Tier4_Scenario_HighThroughputMultiTierCaching(t *testing.T) {
	// Scenario: E-Commerce Catalog & User Session Caching Architecture
	// 1. High-frequency live pricing (Hot Tier: <=1s TTL)
	// 2. Product metadata (Warm Tier: <=5m TTL)
	// 3. System static configurations (Cold Tier: >5m TTL)
	// 4. Bloom filter eliminates database hits for invalid product IDs

	memCache := cache.NewInMemoryCache(50 * time.Millisecond)
	tiered := tieredttl.NewTieredCache(memCache, nil)
	productBloom := bloom.NewBloom(500, 0.01)

	// Step 1: Pre-load catalog products into cache & bloom filter
	for id := 1; id <= 100; id++ {
		productID := fmt.Sprintf("prod_%d", id)
		productBloom.Add(productID)

		// Set Hot price (1s), Warm info (2m), Cold specs (1h)
		_ = tiered.Set(productID+":price", 29.99+float64(id), 1*time.Second)
		_ = tiered.Set(productID+":info", "Product Description", 3*time.Minute)
		_ = tiered.Set(productID+":specs", "Product Specifications", 1*time.Hour)
	}

	// Step 2: Simulate 500 API requests for existing and non-existing products
	hits := 0
	bypasses := 0

	for req := 1; req <= 200; req++ {
		targetID := fmt.Sprintf("prod_%d", (req%100)+1)

		if !productBloom.MayContain(targetID) {
			bypasses++
			continue
		}

		if _, found, _ := tiered.Get(targetID + ":price"); found {
			hits++
		}
		if _, found, _ := tiered.Get(targetID + ":info"); found {
			hits++
		}
		if _, found, _ := tiered.Get(targetID + ":specs"); found {
			hits++
		}
	}

	// Step 3: Query 50 invalid product IDs
	for req := 1000; req < 1050; req++ {
		invalidID := fmt.Sprintf("invalid_prod_%d", req)
		if !productBloom.MayContain(invalidID) {
			bypasses++
		}
	}

	stats := tiered.Stats()
	if hits < 600 { // 200 requests * 3 items = 600 hits
		t.Errorf("expected at least 600 cache hits, got %d (HotHits=%d, WarmHits=%d, ColdHits=%d)",
			hits, stats.HotHits, stats.WarmHits, stats.ColdHits)
	}
	if bypasses < 40 {
		t.Errorf("expected Bloom filter to bypass at least 40 invalid product queries, got %d", bypasses)
	}
}
