# Handoff Report: Pranor Auth & Pranor Cache Codebase Survey

**Author:** Explorer 1 (Survey Phase)  
**Date:** 2026-07-26  
**Target:** Pranor Auth & Pranor Cache modules  
**Working Directory:** `/home/developer/workspace/serv/.agents/explorer_survey_1`

---

## 1. Observation

1. **Repository Location & Environment:**
   * Monorepo base path: `/home/developer/workspace/serv`
   * Target package directories: `/home/developer/workspace/serv/packages/Pranor Auth` and `/home/developer/workspace/serv/packages/Pranor Cache`.
   * Go environment: Go version 1.25.0 in Pranor Auth (`go.mod` line 3), Go version 1.23.0 in Pranor Cache (`go.mod` line 3).
2. **`packages/Pranor Auth` Initial State:**
   * Command `go test ./...` in `packages/Pranor Auth` completed with exit code 0 (`ok github.com/vyuvaraj/pranor/packages/Pranor Auth 1.058s`).
   * Existing packages: `pkg/handlers`, `pkg/kms`, `pkg/mfa`, `pkg/oauth`, `pkg/sessions`, `pkg/store`.
   * Existing `pkg/sessions/sessions.go` contains session tracking and basic failed login helpers.
   * Target implementation path SA.G1: `packages/Pranor Auth/pkg/sessions/token_store.go`.
   * Target implementation path SA.G6: `packages/Pranor Auth/pkg/security/velocity_limiter.go` (`pkg/security` directory currently does not exist).
3. **`packages/Pranor Cache` Initial State:**
   * Command `go test ./...` in `packages/Pranor Cache` completed with exit code 0 (`ok github.com/vyuvaraj/pranor/packages/Pranor Cache 0.965s`, `ok .../pkg/cache 0.214s`, `ok .../pkg/server 0.007s`).
   * Existing packages: `pkg/cache`, `pkg/otel`, `pkg/server`.
   * `pkg/cache/cache.go` exports `InMemoryCache` with `Get`, `Set`, `Delete`, `Clear`, `DeletePattern`.
   * Target implementation path SC.G3: `packages/Pranor Cache/pkg/bloom/bloom.go` (`pkg/bloom` directory currently does not exist).
   * Target implementation path SC.G4: `packages/Pranor Cache/pkg/tieredttl/policy.go` (`pkg/tieredttl` directory currently does not exist).
4. **Requirements Reference:**
   * Requirements R1, R2, R3, R4 in `/home/developer/workspace/serv/.agents/ORIGINAL_REQUEST.md` (lines 34–68).

---

## 2. Logic Chain

1. **Baseline Health:**
   * Execution of `go test ./...` in both `packages/Pranor Auth` and `packages/Pranor Cache` verified that existing test suites pass cleanly.
2. **Package & File Requirements:**
   * **SA.G1 (Opaque Session Token Store):** Lives in existing package `sessions` (`packages/Pranor Auth/pkg/sessions/token_store.go`). Needs `TokenStore` with `Issue(userID)`, `Validate(token)`, `Revoke(token)`. `crypto/rand` generates 32-byte hex tokens, stored in-memory with 7-day default TTL.
   * **SA.G6 (Credential Stuffing Velocity Limiter):** Requires creating new directory `packages/Pranor Auth/pkg/security/` and `velocity_limiter.go`. Needs `VelocityLimiter` tracking sliding window failures per key (IP / username) with `RecordFailure`, `IsBlocked`, and `Reset`.
   * **SC.G3 (Probabilistic Bloom Filter):** Requires creating new directory `packages/Pranor Cache/pkg/bloom/` and `bloom.go`. Needs `Bloom` struct with `NewBloom(capacity, fpRate)`, `Add(key)`, `MayContain(key)`. Mathematical bit array sizing $m = \lceil - capacity \cdot \ln(p) / (\ln 2)^2 \rceil$ and $k = \lceil (m/capacity) \cdot \ln 2 \rceil$ using zero-dependency FNV hashing (`hash/fnv`).
   * **SC.G4 (Tiered TTL Policy Engine):** Requires creating new directory `packages/Pranor Cache/pkg/tieredttl/` and `policy.go`. Needs `TierPolicy` classifying TTLs (Hot $\le 1\text{s}$, Warm $\le 5\text{m}$, Cold $> 5\text{m}$) and `TieredCache` wrapping `InMemoryCache` with `Stats() TierStats` tracking per-tier hit/miss counters.
3. **No External Dependencies:**
   * All 4 features rely exclusively on Go standard library (`crypto/rand`, `encoding/hex`, `hash/fnv`, `math`, `sync`, `time`). Zero modification to `go.mod` dependencies is needed.

---

## 3. Caveats

* **Scope Limit:** This survey covers only Pranor Auth (R1, R2) and Pranor Cache (R3, R4). Other roadmap features (Pranor Chrono CR.G1/G2/G4, Pranor Pool SP.G1/G2, Pranor Pulse SQ.G5) are surveyed by peer agents.
* **Code Modification Constraint:** Explorer 1 is strictly read-only for codebase packages; source implementation files must be created/edited during the implementation phase by implementer agents.
* **Concurrency Assumptions:** All 4 features require thread-safe data structures (`sync.RWMutex` or `sync.Mutex`) to prevent data races under concurrent API calls or test stress tests.

---

## 4. Conclusion

`packages/Pranor Auth` and `packages/Pranor Cache` are clean, building, and fully prepared for the implementation phase. The design specifications in `analysis.md` cover exact exported structs, method signatures, data structures, math formulas, error handling, thread safety, and unit test strategies for SA.G1, SA.G6, SC.G3, and SC.G4.

---

## 5. Verification Method

To verify the investigation and subsequent implementations:

1. **Verify Baseline Build and Tests:**
   ```bash
   cd /home/developer/workspace/serv/packages/Pranor Auth && go test ./...
   cd /home/developer/workspace/serv/packages/Pranor Cache && go test ./...
   ```
2. **Verify Created Target Package Files (post-implementation):**
   ```bash
   ls -la /home/developer/workspace/serv/packages/Pranor Auth/pkg/sessions/token_store.go
   ls -la /home/developer/workspace/serv/packages/Pranor Auth/pkg/security/velocity_limiter.go
   ls -la /home/developer/workspace/serv/packages/Pranor Cache/pkg/bloom/bloom.go
   ls -la /home/developer/workspace/serv/packages/Pranor Cache/pkg/tieredttl/policy.go
   ```
3. **Verify Zero Dependency Additions:**
   ```bash
   cd /home/developer/workspace/serv && git diff packages/Pranor Auth/go.mod packages/Pranor Cache/go.mod
   ```
   (Must output empty diff).
