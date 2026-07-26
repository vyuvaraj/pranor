# Project: Servverse 10 OSS Roadmap Items

## Architecture
- 5 independent Go packages/modules in monorepo:
  - `packages/ServAuth` (Go 1.25.0): sessions (SA.G1), security (SA.G6)
  - `packages/ServCache` (Go 1.23.0): bloom (SC.G3), tieredttl (SC.G4)
  - `packages/ServCron` (Go 1.25.0): cron (CR.G1, CR.G2), config (CR.G4)
  - `packages/ServPool` (Go 1.23.0): routing (SP.G1), pool (SP.G2)
  - `packages/ServQueue` (Go 1.25.0): tracing (SQ.G5), core (SQ.G5 integration)

## Feature Inventory
Every feature from the Survey phase appears here with its assigned milestone.
| # | Feature | Description | Milestone | Source |
|---|---------|-------------|-----------|--------|
| 1 | SA.G1 | ServAuth — Opaque session token store with server-side revocation (`packages/ServAuth/pkg/sessions/token_store.go`) | M1 | survey |
| 2 | SA.G6 | ServAuth — Credential stuffing detection & velocity rate limiter (`packages/ServAuth/pkg/security/velocity_limiter.go`) | M1 | survey |
| 3 | SC.G3 | ServCache — Probabilistic Bloom filter for absent-key elimination (`packages/ServCache/pkg/bloom/bloom.go`) | M2 | survey |
| 4 | SC.G4 | ServCache — Tiered TTL policy engine (`packages/ServCache/pkg/tieredttl/policy.go`) | M2 | survey |
| 5 | CR.G1 | ServCron — DAG job chain pipeline (`packages/ServCron/pkg/cron/cron.go`) | M3 | survey |
| 6 | CR.G2 | ServCron — Per-job retry policy engine (`packages/ServCron/pkg/cron/cron.go`) | M3 | survey |
| 7 | CR.G4 | ServCron — Declarative YAML cron-as-code definitions (`packages/ServCron/pkg/config/jobs_loader.go`) | M3 | survey |
| 8 | SP.G1 | ServPool — Read/Write split router (`packages/ServPool/pkg/routing/rw_splitter.go`) | M4 | survey |
| 9 | SP.G2 | ServPool — Pre-checkout connection health validation (`packages/ServPool/pkg/pool/health_checker.go`) | M4 | survey |
| 10 | SQ.G5 | ServQueue — W3C trace context propagation (`packages/ServQueue/pkg/tracing/traceparent.go` & `packages/ServQueue/pkg/core/engine.go`) | M5 | survey |

## Milestones
| # | Name | Scope | Dependencies | Status |
|---|------|-------|-------------|--------|
| M1 | ServAuth | SA.G1 (token store) & SA.G6 (velocity limiter) | none | IN_PROGRESS |
| M2 | ServCache | SC.G3 (Bloom filter) & SC.G4 (Tiered TTL) | none | IN_PROGRESS |
| M3 | ServCron | CR.G1 (DAG chain), CR.G2 (retries), CR.G4 (YAML loader) | none | IN_PROGRESS |
| M4 | ServPool | SP.G1 (RW splitter) & SP.G2 (Health checker) | none | IN_PROGRESS |
| M5 | ServQueue | SQ.G5 (W3C trace context propagation) | none | IN_PROGRESS |
| M-E2E | E2E Testing | Comprehensive E2E test suite across all 5 modules | M1, M2, M3, M4, M5 | IN_PROGRESS |

## Interface Contracts
### ServAuth Interface Contracts
- `TokenStore`: `Issue(userID string) (token string, err error)`, `Validate(token string) (userID string, err error)`, `Revoke(token string) error`
- `VelocityLimiter`: `RecordFailure(key string)`, `IsBlocked(key string) bool`, `Reset(key string)`

### ServCache Interface Contracts
- `Bloom`: `NewBloom(capacity int, falsePositiveRate float64) *Bloom`, `Add(key string)`, `MayContain(key string) bool`
- `TieredCache`: `TierPolicy.Classify(ttl time.Duration) Tier`, `TierName(t Tier) string`, `Stats() TierStats`

### ServCron Interface Contracts
- `Job` additions: `OnSuccess string`, `OnFailure string`, `MaxRetries int`, `RetryDelayMs int`, `RetryBackoffMult float64`, `RetryCount int`, `LastRetryAt time.Time`
- YAML loader: `LoadJobsFile(path string) ([]cron.Job, error)`, `WatchJobsFile(path string, onChange func([]cron.Job)) error`

### ServPool Interface Contracts
- `RWSplitter`: `ClassifyQuery(sql string) QueryType`, `Route(sql string, primary Manager, replicas []Manager) Manager`
- `HealthChecker`: `Acquire() (*DbConn, error)`, `ValidateFn func(*DbConn) bool`, `Stats() HealthStats`

### ServQueue Interface Contracts
- `Traceparent`: `Inject(headers map[string]string, traceID, spanID string)`, `Extract(headers map[string]string) (traceID, spanID string, sampled bool, ok bool)`, `NewTraceID() string`, `NewSpanID() string`
- `LogEntry` addition: `Traceparent string`
- `Engine`: `Append(topic, payload string, metadata ...map[string]string) (LogEntry, error)`

## Code Layout
- `packages/ServAuth/pkg/sessions/token_store.go` & `token_store_test.go`
- `packages/ServAuth/pkg/security/velocity_limiter.go` & `velocity_limiter_test.go`
- `packages/ServCache/pkg/bloom/bloom.go` & `bloom_test.go`
- `packages/ServCache/pkg/tieredttl/policy.go` & `policy_test.go`
- `packages/ServCron/pkg/cron/cron.go` & `cron_test.go`
- `packages/ServCron/pkg/config/jobs_loader.go` & `jobs_loader_test.go`
- `packages/ServPool/pkg/routing/rw_splitter.go` & `rw_splitter_test.go`
- `packages/ServPool/pkg/pool/health_checker.go` & `health_checker_test.go`
- `packages/ServQueue/pkg/tracing/traceparent.go` & `traceparent_test.go`
- `packages/ServQueue/pkg/core/engine.go` & `engine_test.go`
