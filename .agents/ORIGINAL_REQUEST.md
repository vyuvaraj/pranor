# Original User Request

## Initial Request — 2026-07-26T08:56:38Z

Implement 10 pending OSS roadmap items across 5 Servverse modules
(ServAuth, ServCache, ServCron, ServPool, ServQueue) in the Go monorepo
at `/home/developer/workspace/serv`. Each item is a self-contained
package-level feature with zero external dependency additions.

Working directory: /home/developer/workspace/serv
Integrity mode: development

---

## Selected Items (10 OSS items across 5 modules)

| ID | Module | Feature |
|----|--------|---------|
| SA.G1 | ServAuth | Opaque session token store with server-side revocation |
| SA.G6 | ServAuth | Credential stuffing detection & velocity rate limiter |
| SC.G3 | ServCache | Probabilistic Bloom filter for absent-key elimination |
| SC.G4 | ServCache | Tiered TTL policy engine (Hot / Warm / Cold) |
| CR.G1 | ServCron | DAG job chain pipeline (Job-A → Job-B on success/failure) |
| CR.G2 | ServCron | Per-job retry policy engine (exponential backoff + jitter) |
| CR.G4 | ServCron | Declarative YAML cron-as-code definitions with hot-reload |
| SP.G1 | ServPool | Read/Write split router (primary for writes, replica for reads) |
| SP.G2 | ServPool | Pre-checkout connection health validation (ping + validation query) |
| SQ.G5 | ServQueue | Per-message W3C trace context propagation (OTel traceparent) |

---

## Requirements

### R1. ServAuth — Opaque Session Token Store (SA.G1)
Create `packages/ServAuth/pkg/sessions/token_store.go`.
Implement an opaque refresh token store: `TokenStore` struct with
`Issue(userID string) (token string, err error)`,
`Validate(token string) (userID string, err error)`, and
`Revoke(token string) error`. Tokens must be cryptographically random
(32-byte hex), stored in-memory with TTL (default 7 days), and expire
automatically. Thread-safe. Include tests covering issue, validate,
revoke, and TTL expiry.

### R2. ServAuth — Credential Stuffing Velocity Limiter (SA.G6)
Create `packages/ServAuth/pkg/security/velocity_limiter.go`.
Implement a sliding-window rate limiter tracking failed login attempts
per IP and per username using in-memory counters. Expose:
`RecordFailure(key string)`, `IsBlocked(key string) bool`, and
`Reset(key string)`. Configurable: window duration, max attempts before
block, block duration. Thread-safe. Include tests covering blocking
after threshold, reset, and window expiry.

### R3. ServCache — Probabilistic Bloom Filter (SC.G3)
Create `packages/ServCache/pkg/bloom/bloom.go`.
Implement a Bloom filter with zero external dependencies: bit array +
k hash functions (FNV-based). Expose `NewBloom(capacity int, falsePositiveRate float64) *Bloom`,
`Add(key string)`, `MayContain(key string) bool`. Include tests
verifying: all added keys return MayContain=true, false positive rate
stays below configured threshold for 1000 items, and zero false negatives.

### R4. ServCache — Tiered TTL Policy Engine (SC.G4)
Create `packages/ServCache/pkg/tieredttl/policy.go`.
Implement a three-tier TTL policy engine: Hot (≤1s TTL), Warm (≤5m TTL),
Cold (>5m TTL). Expose `TierPolicy` struct with `Classify(ttl time.Duration) Tier`
and `TierName(t Tier) string`. Wrap the existing `InMemoryCache` with a
`TieredCache` that routes Set calls through the policy and tracks per-tier
hit/miss counters via `Stats() TierStats`. Include tests verifying correct
tier classification and counter increments.

### R5. ServCron — DAG Job Chain Pipeline (CR.G1)
Extend `packages/ServCron/pkg/cron/cron.go`.
Add `OnSuccess string` and `OnFailure string` fields to the `Job` struct
(referencing another job ID to trigger on success/failure). After each
job HTTP callback completes, if `OnSuccess`/`OnFailure` is set and the
referenced job exists and is active, schedule it for immediate execution.
Guard against infinite loops with a max chain depth of 10. Include tests
covering: linear chain (A→B→C), failure branch (A-fail→B), and cycle
guard.

### R6. ServCron — Per-Job Retry Policy (CR.G2)
Extend `packages/ServCron/pkg/cron/cron.go`.
Add `MaxRetries int`, `RetryDelayMs int`, and `RetryBackoffMult float4`
fields to the `Job` struct. On HTTP callback failure, retry up to
`MaxRetries` times with delay = `RetryDelayMs * BackoffMult^attempt` + random jitter
(±10%). Track `RetryCount int` and `LastRetryAt time.Time` on the job.
Include tests covering: successful retry on 2nd attempt, exhausted retries
increment FailureCount, jitter produces non-deterministic delays.

### R7. ServCron — YAML Cron-as-Code (CR.G4)
Create `packages/ServCron/pkg/config/jobs_loader.go`.
Parse a YAML file into []cron.Job. YAML schema:
```yaml
jobs:
  - id: daily-report
    cron: "0 9 * * 1-5"
    target_url: http://api/report
    on_success: notify-slack
  - id: notify-slack
    target_url: http://slack/webhook
```
Expose `LoadJobsFile(path string) ([]cron.Job, error)` and
`WatchJobsFile(path string, onChange func([]cron.Job)) error`
(uses `os.Stat` polling every 5s — no external file-watch library).
No external YAML library — parse manually using `encoding/json` after
converting with `gopkg.in/yaml.v3` **only if already in go.mod**,
otherwise implement a minimal YAML subset parser for the schema above.
Check `go.mod` first. Include tests covering file load and field mapping.

### R8. ServPool — Read/Write Split Router (SP.G1)
Create `packages/ServPool/pkg/routing/rw_splitter.go`.
Implement `RWSplitter` that classifies a query string as read or write:
`ClassifyQuery(sql string) QueryType` returns `QueryTypeRead` or
`QueryTypeWrite` based on the leading SQL keyword (SELECT, WITH → read;
INSERT, UPDATE, DELETE, CREATE, DROP, ALTER → write).
Expose `Route(sql string, primary Manager, replicas []Manager) Manager`
returning the appropriate pool. Include tests for all major SQL verbs,
case-insensitivity, and whitespace-leading queries.

### R9. ServPool — Connection Health Validation (SP.G2)
Create `packages/ServPool/pkg/pool/health_checker.go`.
Implement `HealthChecker` that wraps a `Manager`: before `Acquire()`,
run a configurable validation function `ValidateFn func(*DbConn) bool`
(default: always true — callers inject a real ping). Track
`HealthyAcquires`, `StaleDiscarded` counters in `HealthStats`. On
validation failure, discard the connection (call `Release` then try
again, up to 3 attempts). Include tests covering: healthy conn passes,
unhealthy conn discarded and retried, all-unhealthy returns error.

### R10. ServQueue — W3C Trace Context Propagation (SQ.G5)
Create `packages/ServQueue/pkg/tracing/traceparent.go`.
Implement W3C Trace Context (RFC-compliant):
`Inject(headers map[string]string, traceID, spanID string)` — writes
`traceparent: 00-{traceID}-{spanID}-01` header.
`Extract(headers map[string]string) (traceID, spanID string, sampled bool, ok bool)` — parses
`traceparent` header and returns components.
`NewTraceID() string` and `NewSpanID() string` — generate random
hex-encoded IDs (16 bytes for traceID, 8 bytes for spanID).
Integrate into `packages/ServQueue/pkg/core/engine.go`'s `Append`
method: if a `traceparent` header is present in an optional
`map[string]string` metadata arg, store it on `LogEntry`.
Include tests covering inject/extract round-trip, invalid header rejection,
and new ID generation uniqueness.

---

## Acceptance Criteria

### Build
- [ ] `go build ./...` passes with exit code 0 in each modified module directory
- [ ] No new external dependencies added to any `go.mod` (check with `git diff go.mod`)

### Tests
- [ ] `go test ./...` passes in each modified module directory
- [ ] Each new package has at least 3 test functions covering happy path, edge cases, and error cases
- [ ] No test uses `t.Skip()` to bypass assertions

### Correctness
- [ ] SA.G1: `Revoke(token)` causes subsequent `Validate(token)` to return error
- [ ] SA.G6: After N+1 failures within window, `IsBlocked(key)` returns true
- [ ] SC.G3: Zero false negatives — `MayContain` returns true for every added key
- [ ] CR.G1: Job chain A→B executes B immediately after A succeeds
- [ ] CR.G2: Failed job retries with increasing delay before incrementing FailureCount
- [ ] SP.G1: `ClassifyQuery("SELECT ...")` → Read, `ClassifyQuery("INSERT ...")` → Write
- [ ] SQ.G5: `Extract(Inject(headers, tid, sid))` round-trips correctly

When all items are complete and all acceptance criteria pass, commit everything to the `serv` repo with:
```
git add packages/ServAuth/pkg/sessions/ packages/ServAuth/pkg/security/ packages/ServCache/pkg/bloom/ packages/ServCache/pkg/tieredttl/ packages/ServCron/pkg/cron/ packages/ServCron/pkg/config/ packages/ServPool/pkg/routing/ packages/ServPool/pkg/pool/health_checker.go packages/ServQueue/pkg/tracing/ packages/ServQueue/pkg/core/
git commit -m "feat: implement 10 OSS roadmap items (SA.G1, SA.G6, SC.G3, SC.G4, CR.G1, CR.G2, CR.G4, SP.G1, SP.G2, SQ.G5)"
git push origin main
```
Then report back with: which files were created, test results per module, and any deviations from the spec.
