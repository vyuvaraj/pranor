## 2026-07-26T09:04:57Z
You are Worker 1 Gen 2 (Replacement Implementation for M1: ServAuth).
Working directory: /home/developer/workspace/serv/.agents/worker_m1_gen2

Required Reading:
- `/home/developer/workspace/serv/.agents/ORIGINAL_REQUEST.md` (R1: SA.G1, R2: SA.G6)
- `/home/developer/workspace/serv/PROJECT.md`
- `/home/developer/workspace/serv/.agents/explorer_survey_1/handoff.md`

File Ownership:
- `packages/ServAuth/pkg/sessions/token_store.go`
- `packages/ServAuth/pkg/sessions/token_store_test.go`
- `packages/ServAuth/pkg/security/velocity_limiter.go`
- `packages/ServAuth/pkg/security/velocity_limiter_test.go`

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A teamwork_preview_auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Tasks:
1. Implement `TokenStore` in `packages/ServAuth/pkg/sessions/token_store.go`:
   - `Issue(userID string) (token string, err error)`
   - `Validate(token string) (userID string, err error)`
   - `Revoke(token string) error`
   - Cryptographically random (32-byte hex using `crypto/rand`), stored in-memory with TTL (default 7 days), auto-expiry, thread-safe (`sync.RWMutex`).
2. Implement `VelocityLimiter` in `packages/ServAuth/pkg/security/velocity_limiter.go`:
   - `RecordFailure(key string)`
   - `IsBlocked(key string) bool`
   - `Reset(key string)`
   - Sliding-window rate limiter tracking failed attempts per key (IP / username) using in-memory counters. Configurable window duration, max attempts before block, block duration. Thread-safe (`sync.RWMutex`).
3. Write thorough unit tests covering issue, validate, revoke, TTL expiry, threshold blocking, reset, and window expiry.
4. Run `go build ./...` and `go test ./...` in `/home/developer/workspace/serv/packages/ServAuth`. Ensure exit code 0.
5. Verify `git diff go.mod` in `packages/ServAuth` shows NO external dependency changes.
6. Write `changes.md` and `handoff.md` in `/home/developer/workspace/serv/.agents/worker_m1_gen2`.
7. Send message to orchestrator upon completion.
