## 2026-07-26T09:05:00Z
You are Worker 4 Gen 2 (Replacement Implementation for M4: ServPool).
Working directory: /home/developer/workspace/serv/.agents/worker_m4_gen2

Required Reading:
- `/home/developer/workspace/serv/.agents/ORIGINAL_REQUEST.md` (R8: SP.G1, R9: SP.G2)
- `/home/developer/workspace/serv/PROJECT.md`
- `/home/developer/workspace/serv/.agents/explorer_survey_3/handoff.md`

File Ownership:
- `packages/ServPool/pkg/routing/rw_splitter.go`
- `packages/ServPool/pkg/routing/rw_splitter_test.go`
- `packages/ServPool/pkg/pool/health_checker.go`
- `packages/ServPool/pkg/pool/health_checker_test.go`

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A teamwork_preview_auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Tasks:
1. Create `packages/ServPool/pkg/routing/rw_splitter.go`:
   - `ClassifyQuery(sql string) QueryType` (`QueryTypeRead` for SELECT, WITH; `QueryTypeWrite` for INSERT, UPDATE, DELETE, CREATE, DROP, ALTER). Case-insensitive, leading whitespace/comments stripped.
   - `Route(sql string, primary Manager, replicas []Manager) Manager` returning appropriate pool with round-robin replica distribution and fallback to primary.
2. Create `packages/ServPool/pkg/pool/health_checker.go`:
   - `HealthChecker` wrapping `Manager`: on `Acquire()`, run `ValidateFn func(*DbConn) bool` (default: always true). Track `HealthyAcquires`, `StaleDiscarded` in `HealthStats`. Discard unhealthy conn (`Release`) and retry up to 3 times before returning error.
3. Write thorough unit tests covering all SQL verbs, casing, whitespace, replica routing, health validation pass/discard/retry/error.
4. Run `go build ./...` and `go test ./...` in `/home/developer/workspace/serv/packages/ServPool`. Ensure exit code 0.
5. Verify `git diff go.mod` in `packages/ServPool` shows NO external dependency changes.
6. Write `changes.md` and `handoff.md` in `/home/developer/workspace/serv/.agents/worker_m4_gen2`.
7. Send message to orchestrator upon completion.
