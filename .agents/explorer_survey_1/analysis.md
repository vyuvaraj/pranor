# Technical Survey & Design Analysis: ServAuth & ServCache

**Author:** Explorer 1 (Survey Phase)  
**Date:** 2026-07-26  
**Target Repository:** `/home/developer/workspace/serv`  
**Scope:** `packages/ServAuth` and `packages/ServCache` (Requirements R1, R2, R3, R4)

---

## 1. Overview of Codebase & Environment

### 1.1 `packages/ServAuth` Status
* **Go Version:** `go 1.25.0` (as specified in `go.mod`)
* **Module Path:** `github.com/vyuvaraj/serv/packages/ServAuth`
* **Dependencies:** `github.com/golang-jwt/jwt/v5`, `golang.org/x/crypto`, `github.com/vyuvaraj/serv/packages/ServShared`
* **Existing Structure:**
  ```text
  packages/ServAuth/
  ├── go.mod
  ├── main.go
  ├── main_test.go
  └── pkg/
      ├── handlers/     (handlers.go, stuffing.go, stuffing_oss.go)
      ├── kms/          (kms.go)
      ├── mfa/          (mfa.go)
      ├── oauth/        (oauth.go)
      ├── sessions/     (sessions.go)
      └── store/        (store.go)
  ```
* **Build & Test Status:** `go test ./...` currently exits 0 (passing `main_test.go`).
* **Missing Directory/Files for Assigned Features:**
  * `pkg/sessions/token_store.go` (new file in existing package `sessions`)
  * `pkg/sessions/token_store_test.go` (new test file)
  * `pkg/security/velocity_limiter.go` (new directory `pkg/security` + new file)
  * `pkg/security/velocity_limiter_test.go` (new test file)

---

### 1.2 `packages/ServCache` Status
* **Go Version:** `go 1.23.0` (as specified in `go.mod`)
* **Module Path:** `github.com/vyuvaraj/serv/packages/ServCache`
* **Dependencies:** `github.com/redis/go-redis/v9`, `github.com/cespare/xxhash/v2` (indirect), `github.com/vyuvaraj/serv/packages/ServShared`
* **Existing Structure:**
  ```text
  packages/ServCache/
  ├── go.mod
  ├── main.go
  ├── main_test.go
  ├── ssl_offload.go
  ├── ssl_offload_oss.go
  └── pkg/
      ├── cache/        (cache.go, cache_test.go, redis.go)
      ├── otel/         (otel.go)
      └── server/       (server.go, server_test.go)
  ```
* **Build & Test Status:** `go test ./...` currently exits 0 (passing `cache_test.go` and `server_test.go`).
* **Missing Directory/Files for Assigned Features:**
  * `pkg/bloom/bloom.go` (new directory `pkg/bloom` + new file)
  * `pkg/bloom/bloom_test.go` (new test file)
  * `pkg/tieredttl/policy.go` (new directory `pkg/tieredttl` + new file)
  * `pkg/tieredttl/policy_test.go` (new test file)

---

## 2. Feature Detailed Specifications & Design Recommendations

### 2.1 SA.G1: Opaque Session Token Store
* **Requirement:** R1 (ORIGINAL_REQUEST.md lines 34-43)
* **File Path:** `packages/ServAuth/pkg/sessions/token_store.go`
* **Package Name:** `sessions`

#### Design Specifications
1. **Data Structures:**
   ```go
   type tokenEntry struct {
       userID    string
       expiresAt time.Time
       revoked   bool
   }

   type TokenStore struct {
       mu         sync.RWMutex
       tokens     map[string]*tokenEntry
       defaultTTL time.Duration
   }
   ```
2. **Constructor:**
   ```go
   func NewTokenStore() *TokenStore
   func NewTokenStoreWithTTL(defaultTTL time.Duration) *TokenStore
   ```
   * Default TTL: 7 days (`7 * 24 * time.Hour`).
3. **Core Methods:**
   * `Issue(userID string) (token string, err error)`:
     * Generates a 32-byte cryptographically random byte slice using `crypto/rand.Read`.
     * Encodes byte slice as a 64-character hex string (`hex.EncodeToString`).
     * Saves `tokenEntry` into `tokens` map under write lock.
     * Returns `token` string.
   * `Validate(token string) (userID string, err error)`:
     * Acquires read lock on `mu`.
     * If `token` is missing: return error `ErrTokenNotFound`.
     * If `entry.revoked` is true: return error `ErrTokenRevoked`.
     * If `time.Now().After(entry.expiresAt)`: return error `ErrTokenExpired`.
     * Returns `entry.userID` and `nil`.
   * `Revoke(token string) error`:
     * Acquires write lock on `mu`.
     * If `token` is missing: return error `ErrTokenNotFound` (or nil depending on idempotency, but error if token invalid).
     * Set `entry.revoked = true`.
4. **Key Considerations:**
   * Uses standard `crypto/rand` and `encoding/hex` (zero external dependencies).
   * Thread-safe with `sync.RWMutex`.
   * Clear error return types for revoked, expired, and non-existent tokens.

#### Test Strategy (`token_store_test.go`)
* `TestTokenStore_IssueAndValidate`: Issue token for `user123`, validate returns `user123`.
* `TestTokenStore_Revoke`: Issue token, call `Revoke(token)`, `Validate(token)` returns revoked error.
* `TestTokenStore_TTLExpiry`: Create store with short TTL (10ms), issue token, wait 20ms, `Validate(token)` returns expired error.
* `TestTokenStore_Uniqueness`: Issue 100 tokens, ensure no duplicates and length is 64 hex chars.
* `TestTokenStore_Concurrency`: Parallel workers issuing, validating, and revoking tokens.

---

### 2.2 SA.G6: Credential Stuffing Velocity Limiter
* **Requirement:** R2 (ORIGINAL_REQUEST.md lines 44-52)
* **File Path:** `packages/ServAuth/pkg/security/velocity_limiter.go`
* **Package Name:** `security`

#### Design Specifications
1. **Data Structures:**
   ```go
   type keyState struct {
       failures     []time.Time
       blockedUntil time.Time
   }

   type VelocityLimiter struct {
       mu             sync.Mutex
       windowDuration time.Duration
       maxAttempts    int
       blockDuration  time.Duration
       states         map[string]*keyState
   }
   ```
2. **Constructor:**
   ```go
   func NewVelocityLimiter(windowDuration time.Duration, maxAttempts int, blockDuration time.Duration) *VelocityLimiter
   ```
3. **Core Methods:**
   * `RecordFailure(key string)`:
     * Lock limiter `mu`.
     * Fetch or initialize `keyState` for `key`.
     * Append `time.Now()` to `failures`.
     * Prune timestamps older than `now.Add(-windowDuration)`.
     * If length of `failures` >= `maxAttempts`, set `blockedUntil = now.Add(blockDuration)`.
   * `IsBlocked(key string) bool`:
     * Lock limiter `mu`.
     * Fetch `keyState`. If none exists, return `false`.
     * Check if `now.Before(st.blockedUntil)`. If true, return `true`.
     * Prune old failure timestamps.
     * If length of `failures` >= `maxAttempts`, update `blockedUntil` and return `true`.
     * Otherwise return `false`.
   * `Reset(key string)`:
     * Lock limiter `mu`.
     * Delete key from `states` map.
4. **Key Considerations:**
   * Tracks failures per key (IP address or username).
   * Sliding window prevents boundary burst attacks.
   * Fully thread-safe with `sync.Mutex`. Zero external dependencies.

#### Test Strategy (`velocity_limiter_test.go`)
* `TestVelocityLimiter_ThresholdBlocking`: Set `maxAttempts = 3`. Call `RecordFailure` 3 times -> `IsBlocked` is false. 4th failure -> `IsBlocked` returns true.
* `TestVelocityLimiter_Reset`: Block key, call `Reset(key)`, `IsBlocked` returns false immediately.
* `TestVelocityLimiter_WindowExpiry`: Set `windowDuration = 50ms`. Record 3 failures, wait 60ms, record 1 failure -> `IsBlocked` remains false.
* `TestVelocityLimiter_BlockExpiry`: Set `blockDuration = 50ms`. Block key, sleep 60ms -> `IsBlocked` returns false.
* `TestVelocityLimiter_Concurrency`: Test parallel failed attempts for distinct keys.

---

### 2.3 SC.G3: Probabilistic Bloom Filter
* **Requirement:** R3 (ORIGINAL_REQUEST.md lines 53-60)
* **File Path:** `packages/ServCache/pkg/bloom/bloom.go`
* **Package Name:** `bloom`

#### Design Specifications
1. **Data Structures:**
   ```go
   type Bloom struct {
       mu        sync.RWMutex
       bitset    []uint64
       bitSize   uint64
       numHashes uint32
   }
   ```
2. **Mathematical Parameter Calculation:**
   * Bit size $m = \lceil - \frac{capacity \cdot \ln(p)}{(\ln 2)^2} \rceil$
   * Number of hash functions $k = \lceil \frac{m}{capacity} \cdot \ln 2 \rceil$
3. **Zero-Dependency Hash Strategy:**
   * Use standard library `hash/fnv`.
   * Compute double-hashing (Kirsch-Mitzenmacher optimization):
     * $h_1 = \text{FNV-1a 64-bit hash of } key$
     * $h_2 = \text{FNV-1 64-bit hash of } key$ (or salt key with a fixed byte)
     * For $i = 0 \dots k-1$: $index = (h_1 + uint64(i) \times h_2) \pmod{m}$
4. **Core Methods:**
   * `NewBloom(capacity int, falsePositiveRate float64) *Bloom`
   * `Add(key string)`:
     * Under write lock, compute $k$ hash indices and set bits in `bitset` (`bitset[idx/64] |= 1 << (idx%64)`).
   * `MayContain(key string) bool`:
     * Under read lock, compute $k$ hash indices. If any bit is 0, return `false`. Return `true` if all bits are 1.

#### Test Strategy (`bloom_test.go`)
* `TestBloom_ZeroFalseNegatives`: Add 1000 items. Assert `MayContain(item) == true` for every added item.
* `TestBloom_FalsePositiveRate`: Create Bloom filter for capacity 1000, $fp = 0.05$. Add 1000 keys. Query 10,000 un-added keys. Calculate observed $FP = \text{count} / 10000$. Assert observed $FP \le 0.05$.
* `TestBloom_EdgeCases`: Check handling of capacity 0 or small values, empty string keys, high load factor.

---

### 2.4 SC.G4: Tiered TTL Policy Engine
* **Requirement:** R4 (ORIGINAL_REQUEST.md lines 61-68)
* **File Path:** `packages/ServCache/pkg/tieredttl/policy.go`
* **Package Name:** `tieredttl`

#### Design Specifications
1. **Types & Structs:**
   ```go
   type Tier int

   const (
       TierHot Tier = iota
       TierWarm
       TierCold
   )

   type TierPolicy struct{}

   type TierStats struct {
       HotHits    int64
       HotMisses  int64
       WarmHits   int64
       WarmMisses int64
       ColdHits   int64
       ColdMisses int64
   }

   type TieredCache struct {
       mu         sync.RWMutex
       underlying *cache.InMemoryCache
       policy     TierPolicy
       keyTiers   map[string]Tier
       stats      TierStats
   }
   ```
2. **Classification Logic:**
   * `TierPolicy.Classify(ttl time.Duration) Tier`:
     * If `ttl <= 1 * time.Second` -> `TierHot`
     * Else if `ttl <= 5 * time.Minute` -> `TierWarm`
     * Else -> `TierCold`
   * `TierPolicy.TierName(t Tier) string`:
     * `TierHot` -> `"Hot"`
     * `TierWarm` -> `"Warm"`
     * `TierCold` -> `"Cold"`
3. **TieredCache Methods:**
   * `NewTieredCache(underlying *cache.InMemoryCache, policy TierPolicy) *TieredCache`
   * `Set(key string, value interface{}, ttl time.Duration) error`:
     * Determine tier: `tier := policy.Classify(ttl)`.
     * Store `keyTiers[key] = tier` under lock.
     * Delegate to `underlying.Set(key, value, ttl)`.
   * `Get(key string) (interface{}, bool, error)`:
     * Delegate to `underlying.Get(key)`.
     * Look up tier from `keyTiers[key]`. If not present, default to `TierCold`.
     * If `found`: increment `HotHits`, `WarmHits`, or `ColdHits` atomically or under mutex.
     * If `!found`: increment `HotMisses`, `WarmMisses`, or `ColdMisses` and delete key from `keyTiers`.
     * Return `value, found, err`.
   * `Delete(key string) error`:
     * Remove from `keyTiers` and call `underlying.Delete(key)`.
   * `Stats() TierStats`:
     * Return snapshot of `stats`.

#### Test Strategy (`policy_test.go`)
* `TestTierPolicy_Classification`:
  * 500ms -> Hot, 1s -> Hot
  * 1s 1ms -> Warm, 5m -> Warm
  * 5m 1s -> Cold, 1h -> Cold
* `TestTieredCache_HitMissCounters`:
  * Set item A with TTL 500ms (Hot). Call `Get("A")` twice -> `HotHits = 2`.
  * Call `Get("B")` (absent) -> `ColdMisses = 1` or appropriate tier miss.
* `TestTieredCache_Delete`:
  * Set item, delete item, verify removal and clean stats behavior.

---

## 3. Recommended Implementation Plan & Verification

1. **ServAuth**:
   * Implement `packages/ServAuth/pkg/sessions/token_store.go` and `token_store_test.go`.
   * Create directory `packages/ServAuth/pkg/security/` and implement `velocity_limiter.go` and `velocity_limiter_test.go`.
   * Verification: Run `go test -v ./...` inside `packages/ServAuth`.

2. **ServCache**:
   * Create directory `packages/ServCache/pkg/bloom/` and implement `bloom.go` and `bloom_test.go`.
   * Create directory `packages/ServCache/pkg/tieredttl/` and implement `policy.go` and `policy_test.go`.
   * Verification: Run `go test -v ./...` inside `packages/ServCache`.

3. **Global Acceptance Checks**:
   * `go build ./...` passes in both packages.
   * `git diff go.mod` shows no additions of external dependencies.
   * `go test ./...` passes all unit tests without using `t.Skip()`.
