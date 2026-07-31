# Project: Pranor 10 OSS Roadmap Items

## Architecture
- 5 independent Go packages/modules in monorepo:
  - `packages/Pranor Auth` (Go 1.25.0): sessions (SA.G1), security (SA.G6)
  - `packages/Pranor Cache` (Go 1.23.0): bloom (SC.G3), tieredttl (SC.G4)
  - `packages/Pranor Chrono` (Go 1.25.0): cron (CR.G1, CR.G2), config (CR.G4)
  - `packages/Pranor Pool` (Go 1.23.0): routing (SP.G1), pool (SP.G2)
  - `packages/Pranor Pulse` (Go 1.25.0): tracing (SQ.G5), core (SQ.G5 integration)

## Feature Inventory
Every feature from the Survey phase appears here with its assigned milestone.
| # | Feature | Description | Milestone | Source |
|---|---------|-------------|-----------|--------|
| 1 | SA.G1 | Pranor Auth — Opaque session token store with server-side revocation (`packages/Pranor Auth/pkg/sessions/token_store.go`) | M1 | survey |
| 2 | SA.G6 | Pranor Auth — Credential stuffing detection & velocity rate limiter (`packages/Pranor Auth/pkg/security/velocity_limiter.go`) | M1 | survey |
| 3 | SC.G3 | Pranor Cache — Probabilistic Bloom filter for absent-key elimination (`packages/Pranor Cache/pkg/bloom/bloom.go`) | M2 | survey |
| 4 | SC.G4 | Pranor Cache — Tiered TTL policy engine (`packages/Pranor Cache/pkg/tieredttl/policy.go`) | M2 | survey |
| 5 | CR.G1 | Pranor Chrono — DAG job chain pipeline (`packages/Pranor Chrono/pkg/cron/cron.go`) | M3 | survey |
| 6 | CR.G2 | Pranor Chrono — Per-job retry policy engine (`packages/Pranor Chrono/pkg/cron/cron.go`) | M3 | survey |
| 7 | CR.G4 | Pranor Chrono — Declarative YAML cron-as-code definitions (`packages/Pranor Chrono/pkg/config/jobs_loader.go`) | M3 | survey |
| 8 | SP.G1 | Pranor Pool — Read/Write split router (`packages/Pranor Pool/pkg/routing/rw_splitter.go`) | M4 | survey |
| 9 | SP.G2 | Pranor Pool — Pre-checkout connection health validation (`packages/Pranor Pool/pkg/pool/health_checker.go`) | M4 | survey |
| 10 | SQ.G5 | Pranor Pulse — W3C trace context propagation (`packages/Pranor Pulse/pkg/tracing/traceparent.go` & `packages/Pranor Pulse/pkg/core/engine.go`) | M5 | survey |

## Milestones
| # | Name | Scope | Dependencies | Status |
|---|------|-------|-------------|--------|
| M1 | Pranor Auth | SA.G1 (token store) & SA.G6 (velocity limiter) | none | IN_PROGRESS |
| M2 | Pranor Cache | SC.G3 (Bloom filter) & SC.G4 (Tiered TTL) | none | IN_PROGRESS |
| M3 | Pranor Chrono | CR.G1 (DAG chain), CR.G2 (retries), CR.G4 (YAML loader) | none | IN_PROGRESS |
| M4 | Pranor Pool | SP.G1 (RW splitter) & SP.G2 (Health checker) | none | IN_PROGRESS |
| M5 | Pranor Pulse | SQ.G5 (W3C trace context propagation) | none | IN_PROGRESS |
| M-E2E | E2E Testing | Comprehensive E2E test suite across all 5 modules | M1, M2, M3, M4, M5 | IN_PROGRESS |

## Interface Contracts
### Pranor Auth Interface Contracts
- `TokenStore`: `Issue(userID string) (token string, err error)`, `Validate(token string) (userID string, err error)`, `Revoke(token string) error`
- `VelocityLimiter`: `RecordFailure(key string)`, `IsBlocked(key string) bool`, `Reset(key string)`

### Pranor Cache Interface Contracts
- `Bloom`: `NewBloom(capacity int, falsePositiveRate float64) *Bloom`, `Add(key string)`, `MayContain(key string) bool`
- `TieredCache`: `TierPolicy.Classify(ttl time.Duration) Tier`, `TierName(t Tier) string`, `Stats() TierStats`

### Pranor Chrono Interface Contracts
- `Job` additions: `OnSuccess string`, `OnFailure string`, `MaxRetries int`, `RetryDelayMs int`, `RetryBackoffMult float64`, `RetryCount int`, `LastRetryAt time.Time`
- YAML loader: `LoadJobsFile(path string) ([]cron.Job, error)`, `WatchJobsFile(path string, onChange func([]cron.Job)) error`

### Pranor Pool Interface Contracts
- `RWSplitter`: `ClassifyQuery(sql string) QueryType`, `Route(sql string, primary Manager, replicas []Manager) Manager`
- `HealthChecker`: `Acquire() (*DbConn, error)`, `ValidateFn func(*DbConn) bool`, `Stats() HealthStats`

### Pranor Pulse Interface Contracts
- `Traceparent`: `Inject(headers map[string]string, traceID, spanID string)`, `Extract(headers map[string]string) (traceID, spanID string, sampled bool, ok bool)`, `NewTraceID() string`, `NewSpanID() string`
- `LogEntry` addition: `Traceparent string`
- `Engine`: `Append(topic, payload string, metadata ...map[string]string) (LogEntry, error)`

## Code Layout
- `packages/Pranor Auth/pkg/sessions/token_store.go` & `token_store_test.go`
- `packages/Pranor Auth/pkg/security/velocity_limiter.go` & `velocity_limiter_test.go`
- `packages/Pranor Cache/pkg/bloom/bloom.go` & `bloom_test.go`
- `packages/Pranor Cache/pkg/tieredttl/policy.go` & `policy_test.go`
- `packages/Pranor Chrono/pkg/cron/cron.go` & `cron_test.go`
- `packages/Pranor Chrono/pkg/config/jobs_loader.go` & `jobs_loader_test.go`
- `packages/Pranor Pool/pkg/routing/rw_splitter.go` & `rw_splitter_test.go`
- `packages/Pranor Pool/pkg/pool/health_checker.go` & `health_checker_test.go`
- `packages/Pranor Pulse/pkg/tracing/traceparent.go` & `traceparent_test.go`
- `packages/Pranor Pulse/pkg/core/engine.go` & `engine_test.go`
