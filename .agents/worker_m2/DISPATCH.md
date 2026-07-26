## 2026-07-26T09:00:46Z

You are Worker 2 (Implementation for M2: ServCache).
Working directory: /home/developer/workspace/serv/.agents/worker_m2

Required Reading:
- `/home/developer/workspace/serv/.agents/ORIGINAL_REQUEST.md` (R3: SC.G3, R4: SC.G4)
- `/home/developer/workspace/serv/PROJECT.md`
- `/home/developer/workspace/serv/.agents/explorer_survey_1/handoff.md`

File Ownership:
- `packages/ServCache/pkg/bloom/bloom.go`
- `packages/ServCache/pkg/bloom/bloom_test.go`
- `packages/ServCache/pkg/tieredttl/policy.go`
- `packages/ServCache/pkg/tieredttl/policy_test.go`

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A teamwork_preview_auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Tasks:
1. Implement `Bloom` in `packages/ServCache/pkg/bloom/bloom.go`:
   - `NewBloom(capacity int, falsePositiveRate float64) *Bloom`
   - `Add(key string)`
   - `MayContain(key string) bool`
   - Zero external dependencies: bit array + k hash functions (FNV-based: FNV-1a with seed/double hashing). Verify zero false negatives, false positive rate below threshold for 1000 items. Thread-safe.
2. Implement `TierPolicy` & `TieredCache` in `packages/ServCache/pkg/tieredttl/policy.go`:
   - Hot (<=1s TTL), Warm (<=5m TTL), Cold (>5m TTL).
   - `Classify(ttl time.Duration) Tier`, `TierName(t Tier) string`.
   - Wrap existing `InMemoryCache` with `TieredCache` routing Set calls through policy, tracking per-tier hit/miss counters via `Stats() TierStats`. Thread-safe.
3. Write thorough unit tests for Bloom filter and Tiered TTL policy.
4. Run `go build ./...` and `go test ./...` in `/home/developer/workspace/serv/packages/ServCache`. Ensure exit code 0.
5. Verify `git diff go.mod` in `packages/ServCache` shows NO external dependency changes.
6. Write `changes.md` and `handoff.md` in `/home/developer/workspace/serv/.agents/worker_m2`.
7. Send message to orchestrator upon completion.
