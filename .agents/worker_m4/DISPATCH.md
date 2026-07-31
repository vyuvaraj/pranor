## 2026-07-26T09:00:46Z
<USER_REQUEST>
You are Worker 4 (Implementation for M4: Pranor Pool).
Working directory: /home/developer/workspace/pranor/.agents/worker_m4

Required Reading:
- `/home/developer/workspace/pranor/.agents/ORIGINAL_REQUEST.md` (R8: SP.G1, R9: SP.G2)
- `/home/developer/workspace/pranor/PROJECT.md`
- `/home/developer/workspace/pranor/.agents/explorer_survey_3/handoff.md`

File Ownership:
- `packages/Pranor Pool/pkg/routing/rw_splitter.go`
- `packages/Pranor Pool/pkg/routing/rw_splitter_test.go`
- `packages/Pranor Pool/pkg/pool/health_checker.go`
- `packages/Pranor Pool/pkg/pool/health_checker_test.go`

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A teamwork_preview_auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Tasks:
1. Create `packages/Pranor Pool/pkg/routing/rw_splitter.go`:
   - `ClassifyQuery(sql string) QueryType` (`QueryTypeRead` for SELECT, WITH; `QueryTypeWrite` for INSERT, UPDATE, DELETE, CREATE, DROP, ALTER). Case-insensitive, leading whitespace/comments stripped.
   - `Route(sql string, primary Manager, replicas []Manager) Manager` returning appropriate pool with round-robin replica distribution and fallback to primary.
2. Create `packages/Pranor Pool/pkg/pool/health_checker.go`:
   - `HealthChecker` wrapping `Manager`: on `Acquire()`, run `ValidateFn func(*DbConn) bool` (default: always true). Track `HealthyAcquires`, `StaleDiscarded` in `HealthStats`. Discard unhealthy conn (`Release`) and retry up to 3 times before returning error.
3. Write thorough unit tests covering all SQL verbs, casing, whitespace, replica routing, health validation pass/discard/retry/error.
4. Run `go build ./...` and `go test ./...` in `/home/developer/workspace/pranor/packages/Pranor Pool`. Ensure exit code 0.
5. Verify `git diff go.mod` in `packages/Pranor Pool` shows NO external dependency changes.
6. Write `changes.md` and `handoff.md` in `/home/developer/workspace/pranor/.agents/worker_m4`.
7. Send message to orchestrator upon completion.
</USER_REQUEST>
