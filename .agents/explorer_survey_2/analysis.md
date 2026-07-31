# Pranor Chrono Architectural Survey & Technical Analysis

**Module Path**: `packages/Pranor Chrono`  
**Package Namespace**: `github.com/vyuvaraj/pranor/packages/Pranor Chrono`  
**Target Requirements**: R5 (CR.G1), R6 (CR.G2), R7 (CR.G4)

---

## Executive Summary

This report presents a thorough survey and architectural analysis of `packages/Pranor Chrono` for implementing requirements R5 (CR.G1: DAG Job Chain Pipeline), R6 (CR.G2: Per-Job Retry Policy Engine), and R7 (CR.G4: Declarative YAML Cron-as-Code).

Key findings:
1. `packages/Pranor Chrono` is currently a functional Go 1.23 service with existing scheduler and leader election components.
2. `go.mod` does **not** contain `gopkg.in/yaml.v3`. Per requirement R7 and monorepo constraints ("zero external dependency additions"), a minimal custom YAML subset parser must be implemented in `pkg/config/jobs_loader.go`.
3. Existing `go test ./...` in `packages/Pranor Chrono` passes.
4. Job execution logic in `pkg/cron/cron.go` can be cleanly extended for DAG chaining and retries while preserving existing features (e.g. S3 persistence, syslog emission, queue publishing).

---

## Codebase Analysis

### Existing Package Structure
```
packages/Pranor Chrono/
├── main.go                       # Entry point (HTTP server + leader election + scheduler)
├── main_test.go                  # Integration tests
├── go.mod                        # Go module file
├── go.sum                        # Go sum file
└── pkg/
    ├── cron/                     # Core cron scheduler package
    │   ├── cron.go               # Job struct, Scheduler, HTTP execution, cron expression parser
    │   ├── cron_hardening_test.go# Tests for missed execution, leader election, edge cases
    │   ├── distributed.go        # Leader election implementations (Standalone, Redis, Pranor Lock)
    │   ├── smart_schedule.go     # Conflict analyzer
    │   └── smart_schedule_test.go# Tests for smart schedule analysis
    ├── otel/                     # OpenTelemetry setup
    │   └── otel.go
    └── server/                   # HTTP API handlers
        └── server.go
```

### Module Dependencies (`go.mod`)
- `github.com/redis/go-redis/v9 v9.7.0`
- `github.com/vyuvaraj/pranor/packages/Pranor Core v0.0.0`
- `gopkg.in/yaml.v3`: **NOT present**.

---

## Detailed Requirement Analysis & Design Proposals

### 1. CR.G1: DAG Job Chain Pipeline (Requirement R5)

#### Goal
Enable job chaining where job completion automatically triggers dependent jobs based on execution outcome (`OnSuccess` or `OnFailure`), with recursion protection (max depth 10).

#### Struct Extensions (`pkg/cron/cron.go`)
Extend `Job` struct with:
```go
type Job struct {
    // Existing fields: ID, Interval, Cron, TargetURL, Payload, NextTopic, NextRun, LastRun, Status, LastOutcome, FailureCount
    OnSuccess string `json:"on_success,omitempty"` // ID of active job to trigger on HTTP success (2xx)
    OnFailure string `json:"on_failure,omitempty"` // ID of active job to trigger on HTTP failure
}
```

#### Execution Logic
1. Refactor `executeJob(job *Job)` to internal `executeJobWithDepth(job *Job, depth int)`. `executeJob(job)` defaults to `depth = 0`.
2. When job HTTP callback finishes and outcome is evaluated (`LastOutcome` is set to `"success"` or `"failed"`):
   - If `LastOutcome == "success"` and `job.OnSuccess != ""`:
     Look up `nextJob` by ID `job.OnSuccess` in `s.jobs`. If `nextJob` exists, `nextJob.Status == "active"`, and `depth < 10`:
     Trigger immediate execution via `go s.executeJobWithDepth(nextJob, depth+1)`.
   - If `LastOutcome == "failed"` and `job.OnFailure != ""`:
     Look up `nextJob` by ID `job.OnFailure` in `s.jobs`. If `nextJob` exists, `nextJob.Status == "active"`, and `depth < 10`:
     Trigger immediate execution via `go s.executeJobWithDepth(nextJob, depth+1)`.
3. If `depth >= 10`, abort chain progression to prevent infinite loop cycles (cycle guard).

---

### 2. CR.G2: Per-Job Retry Policy Engine (Requirement R6)

#### Goal
Implement exponential backoff with jitter retries for failed job HTTP callbacks before marking a job as failed.

#### Struct Extensions (`pkg/cron/cron.go`)
Extend `Job` struct with:
```go
type Job struct {
    // CR.G2 Retry Policy fields:
    MaxRetries       int       `json:"max_retries,omitempty"`
    RetryDelayMs     int       `json:"retry_delay_ms,omitempty"`
    RetryBackoffMult float64   `json:"retry_backoff_mult,omitempty"`
    RetryCount       int       `json:"retry_count"`
    LastRetryAt      time.Time `json:"last_retry_at,omitempty"`
}
```

#### Execution & Retry Logic
1. At start of execution (`executeJobWithDepth`): reset `job.RetryCount = 0`.
2. Determine backoff multiplier: default `backoff = job.RetryBackoffMult`; if `backoff <= 0`, set `backoff = 1.0`.
3. Perform HTTP request to `job.TargetURL`.
4. If request fails (`err != nil` or `statusCode < 200 || statusCode >= 300`):
   - Check if `job.MaxRetries > 0` and `job.RetryCount < job.MaxRetries`.
   - If retries remain:
     - Increment `job.RetryCount`.
     - Record `job.LastRetryAt = time.Now()`.
     - Compute delay: `baseMs = float64(job.RetryDelayMs) * math.Pow(backoff, float64(job.RetryCount-1))`.
     - Apply ±10% random jitter: `jitter := 0.9 + rand.Float64()*0.2`, `delay = time.Duration(baseMs * jitter) * time.Millisecond`.
     - `time.Sleep(delay)`.
     - Retry HTTP POST request.
   - If retries exhausted (or `MaxRetries == 0`):
     - Mark `LastOutcome = "failed"`, increment `FailureCount++`.
5. If request succeeds (initially or after retry):
   - Mark `LastOutcome = "success"`, reset `FailureCount = 0`.

---

### 3. CR.G4: Declarative YAML Cron-as-Code (Requirement R7)

#### Goal
Parse job definitions from a declarative YAML file into `[]cron.Job`, and support hot-reloading via file polling (`os.Stat` every 5 seconds).

#### File Placement
Create `packages/Pranor Chrono/pkg/config/jobs_loader.go`.

#### Package API
```go
package config

import (
    "github.com/vyuvaraj/pranor/packages/Pranor Chrono/pkg/cron"
)

func LoadJobsFile(path string) ([]cron.Job, error)
func WatchJobsFile(path string, onChange func([]cron.Job)) error
```

#### Minimal YAML Parser Design
Since `gopkg.in/yaml.v3` is absent from `go.mod`, implement a minimal line-based YAML subset parser in `jobs_loader.go`:
- Supported schema format:
```yaml
jobs:
  - id: daily-report
    cron: "0 9 * * 1-5"
    target_url: http://api/report
    on_success: notify-slack
    max_retries: 3
    retry_delay_ms: 500
    retry_backoff_mult: 2.0
  - id: notify-slack
    target_url: http://slack/webhook
```
- Parser logic:
  1. Read file bytes and split by lines.
  2. Filter blank lines and comments starting with `#`.
  3. Detect `- ` prefix to instantiate a new `cron.Job`.
  4. Parse key-value pairs (e.g. `id: ...`, `target_url: ...`, `max_retries: ...`).
  5. Strip string quotes (`"..."` or `'...'`).
  6. Parse integers for `max_retries`, `retry_delay_ms` and floats for `retry_backoff_mult`.
  7. Return `[]cron.Job`.

#### Hot-Reload File Watcher
- `WatchJobsFile(path string, onChange func([]cron.Job)) error`:
  - Run `os.Stat(path)` to verify existence and capture initial `ModTime` / `Size`.
  - Return error if file doesn't exist.
  - Launch a background goroutine with a 5-second `time.Ticker`.
  - On tick, run `os.Stat(path)`: if `ModTime` or `Size` changed:
    - Invoke `LoadJobsFile(path)`.
    - If load succeeds, call `onChange(jobs)` with the updated job slice.

---

## Verification Plan

### Command Verification
1. `go build ./...` inside `packages/Pranor Chrono` must compile cleanly with 0 exit code.
2. `go test ./...` inside `packages/Pranor Chrono` must execute all unit and integration tests with 0 exit code.
3. `git diff go.mod` must produce empty output (no dependency additions).

### Test Coverage Requirements
1. **CR.G1 (DAG Chain)**:
   - Test linear chain execution: A -> B -> C.
   - Test failure branch execution: A (fail) -> B (triggered on failure).
   - Test cycle guard: A -> B -> A terminated at depth 10.
2. **CR.G2 (Retries)**:
   - Test success on 2nd attempt.
   - Test exhausted retries (increments `FailureCount`).
   - Test non-deterministic jitter delay generation.
3. **CR.G4 (YAML Loader)**:
   - Test loading jobs from YAML file with all fields.
   - Test `WatchJobsFile` file modification trigger.
   - Test invalid path or malformed file error handling.
