## 2026-07-26T09:00:46Z
You are Worker 3 (Implementation for M3: Pranor Chrono).
Working directory: /home/developer/workspace/pranor/.agents/worker_m3

Required Reading:
- `/home/developer/workspace/pranor/.agents/ORIGINAL_REQUEST.md` (R5: CR.G1, R6: CR.G2, R7: CR.G4)
- `/home/developer/workspace/pranor/PROJECT.md`
- `/home/developer/workspace/pranor/.agents/explorer_survey_2/handoff.md`

File Ownership:
- `packages/Pranor Chrono/pkg/cron/cron.go`
- `packages/Pranor Chrono/pkg/cron/cron_test.go`
- `packages/Pranor Chrono/pkg/config/jobs_loader.go`
- `packages/Pranor Chrono/pkg/config/jobs_loader_test.go`

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A teamwork_preview_auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Tasks:
1. Extend `Job` struct in `packages/Pranor Chrono/pkg/cron/cron.go`:
   - Add `OnSuccess string` and `OnFailure string`. On callback completion, if set and target job exists/active, schedule immediately. Guard against infinite loops (max depth 10).
   - Add `MaxRetries int`, `RetryDelayMs int`, `RetryBackoffMult float64`, `RetryCount int`, `LastRetryAt time.Time`. On HTTP callback failure, retry up to `MaxRetries` with exponential backoff (`RetryDelayMs * BackoffMult^attempt` + random +-10% jitter).
2. Create `packages/Pranor Chrono/pkg/config/jobs_loader.go`:
   - `LoadJobsFile(path string) ([]cron.Job, error)`
   - `WatchJobsFile(path string, onChange func([]cron.Job)) error` (using 5s `os.Stat` polling).
   - Custom minimal YAML subset parser (since `gopkg.in/yaml.v3` is NOT in `go.mod`).
3. Write thorough unit tests covering DAG chain (A->B->C, failure branch, cycle guard), retries (successful retry on 2nd attempt, exhausted retries, jitter), and YAML loading/watching.
4. Run `go build ./...` and `go test ./...` in `/home/developer/workspace/pranor/packages/Pranor Chrono`. Ensure exit code 0.
5. Verify `git diff go.mod` in `packages/Pranor Chrono` shows NO external dependency changes.
6. Write `changes.md` and `handoff.md` in `/home/developer/workspace/pranor/.agents/worker_m3`.
7. Send message to orchestrator upon completion.
