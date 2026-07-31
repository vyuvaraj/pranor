# Changes Report — M3: Pranor Chrono (CR.G1, CR.G2, CR.G4)

## Files Modified and Created

### 1. `packages/Pranor Chrono/pkg/cron/cron.go` (Modified)
- **`Job` Struct Extensions**: Added `OnSuccess string`, `OnFailure string`, `MaxRetries int`, `RetryDelayMs int`, `RetryBackoffMult float64`, `RetryCount int`, `LastRetryAt time.Time` with JSON tags.
- **DAG Job Chain Execution (`CR.G1`)**: Added `executeJobWithDepth(job *Job, depth int)` which passes chain depth through executions. When job completes, if `depth < 10`, triggers `OnSuccess` or `OnFailure` target job immediately if present, active, and registered in scheduler. Max depth of 10 prevents infinite recursion in cycles.
- **Per-Job Retry Policy Engine (`CR.G2`)**: Added `CalculateRetryDelay(retryDelayMs int, backoffMult float64, attempt int)` calculating exponential backoff `RetryDelayMs * BackoffMult^attempt` with ±10% random jitter. On HTTP callback failure, retries up to `MaxRetries` times, updating `RetryCount` and `LastRetryAt`. Updates `FailureCount` only if all retries are exhausted.
- **Unscheduled Jobs Support**: Updated `calculateNextRun` and `checkAndRunJobs` to support trigger-only callback jobs (without `cron` or `interval` schedules) without raising errors or auto-triggering on tick.

### 2. `packages/Pranor Chrono/pkg/cron/cron_test.go` (Created)
- `TestDAGChainLinear`: Verifies A -> B -> C execution chain triggered on success.
- `TestDAGChainFailureBranch`: Verifies A-fail -> B execution branch triggered on failure.
- `TestDAGChainCycleGuard`: Verifies cyclic chain A -> B -> A halts precisely at max depth 10.
- `TestRetrySuccessfulOnSecondAttempt`: Verifies retry engine succeeds on 2nd attempt, resetting failure count and setting `RetryCount = 1`.
- `TestRetryExhaustedRetries`: Verifies exhausted retries increment `FailureCount` and set `LastOutcome = "failed"`.
- `TestRetryJitterNonDeterministicDelays`: Verifies `CalculateRetryDelay` produces varied, non-deterministic values within ±10% bounds.

### 3. `packages/Pranor Chrono/pkg/config/jobs_loader.go` (Created)
- **`LoadJobsFile(path string)`**: Reads file and parses YAML using `ParseJobsYAML`.
- **`WatchJobsFile(path string, onChange func([]cron.Job))`**: Uses 5-second `os.Stat` polling loop to watch for modification time updates and invoke `onChange`. Validates initial file existence before launching background polling.
- **`ParseJobsYAML(data []byte)`**: Minimal zero-dependency YAML subset parser supporting `jobs:` array with string, int, and float64 field mappings, comments, and quoting.

### 4. `packages/Pranor Chrono/pkg/config/jobs_loader_test.go` (Created)
- `TestLoadJobsFile`: Verifies parsing YAML content with all job fields into `[]cron.Job`.
- `TestWatchJobsFile`: Verifies error on non-existent initial file and hot-reload `onChange` callback invocation upon file modification.

---

## Build and Test Verification

```bash
# 1. Compilation
cd /home/developer/workspace/serv/packages/Pranor Chrono
go build ./...
# Exit code: 0

# 2. Test Execution
go test -v ./...
# Exit code: 0 (All packages passed: pkg/cron, pkg/config)

# 3. Dependency Integrity Check
git diff go.mod
# Exit code: 0 (Zero external dependency additions)
```
