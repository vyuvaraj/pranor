# Changes Summary — Milestone M2 (Pranor Cache)

## Created Files

1. **`packages/Pranor Cache/pkg/bloom/bloom.go`**
   - Implemented `Bloom` struct with thread-safe `sync.RWMutex`.
   - `NewBloom(capacity int, falsePositiveRate float64) *Bloom`: calculates optimal bit array size $m = \lceil -capacity \cdot \ln(p) / (\ln 2)^2 \rceil$ and hash count $k = \lceil (m/capacity) \cdot \ln 2 \rceil$.
   - `Add(key string)`: computes $k$ indices via Kirsch-Mitzenmacher double hashing using standard library `hash/fnv` (FNV-1a 64-bit with seed byte) and sets bits in `[]uint64` bitset under write lock.
   - `MayContain(key string) bool`: checks $k$ bit indices under read lock, returning `false` if any bit is missing (zero false negatives) or `true` if all set.
   - Helper methods `Capacity()`, `FalsePositiveRate()`, `M()`, and `K()`.

2. **`packages/Pranor Cache/pkg/bloom/bloom_test.go`**
   - `TestBloom_ZeroFalseNegatives`: verifies that 1000 inserted items all return `MayContain = true`.
   - `TestBloom_FalsePositiveRate`: verifies that for 1000 inserted items and 10,000 un-added queries, observed false positive rate stays below target threshold.
   - `TestBloom_EdgeAndInvalidParameters`: tests fallback for invalid parameters, empty string, long key.
   - `TestBloom_Concurrency`: tests concurrent readers and writers across multiple goroutines.

3. **`packages/Pranor Cache/pkg/tieredttl/policy.go`**
   - Implemented `Tier` enum (`TierHot`, `TierWarm`, `TierCold`).
   - Implemented `TierPolicy` with `Classify(ttl time.Duration) Tier` (Hot <= 1s, Warm <= 5m, Cold > 5m) and `TierName(t Tier) string`. Also exposed package-level `Classify` and `TierName` functions.
   - Implemented `TierStats` struct tracking `HotHits`, `HotMisses`, `WarmHits`, `WarmMisses`, `ColdHits`, `ColdMisses`.
   - Implemented `TieredCache` wrapping `cache.Cache` interface:
     - `Set(key, value, ttl)`: classifies TTL, saves key-to-tier mapping, delegates to underlying cache.
     - `Get(key)`: retrieves value from underlying cache, updates hit/miss counter for the key's tier under lock, cleans mapping on expiration/miss.
     - `Delete(key)`, `Clear()`, `DeletePattern(pattern)`: cleans tier mappings and delegates to underlying cache.
     - `Stats()`: returns snapshot of `TierStats`.

4. **`packages/Pranor Cache/pkg/tieredttl/policy_test.go`**
   - `TestTierPolicy_ClassificationAndNaming`: verifies boundary values for Hot (<=1s), Warm (<=5m), Cold (>5m), and tier names.
   - `TestTieredCache_HitAndMissCounters`: verifies set/get hits and misses update respective tier counters.
   - `TestTieredCache_ExpiryAndMisses`: verifies expired keys record misses for their configured tier.
   - `TestTieredCache_DeleteAndClear`: verifies key deletion and clearing.
   - `TestTieredCache_DeletePattern`: verifies wildcard pattern deletion cleans tier mapping and cache.
   - `TestTieredCache_Concurrency`: tests concurrent set/get/stats operations.

## Dependency Verification
- Run `git diff go.mod` in `packages/Pranor Cache`: no changes. Zero external dependencies added.
