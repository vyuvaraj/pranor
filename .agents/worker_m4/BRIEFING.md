# BRIEFING — 2026-07-26T09:00:46Z

## Mission
Implement M4: Pranor Pool features - Read/Write splitting (`rw_splitter.go`) and Connection Health Checker (`health_checker.go`), with unit tests and zero external dependency changes.

## 🔒 My Identity
- Archetype: implementer / qa / specialist
- Roles: implementer, qa, specialist
- Working directory: /home/developer/workspace/pranor/.agents/worker_m4
- Original parent: 3c3be357-fb61-4d2b-9ed3-40099ef64f03
- Milestone: M4 (Pranor Pool)

## 🔒 Key Constraints
- File Ownership:
  - `packages/Pranor Pool/pkg/routing/rw_splitter.go`
  - `packages/Pranor Pool/pkg/routing/rw_splitter_test.go`
  - `packages/Pranor Pool/pkg/pool/health_checker.go`
  - `packages/Pranor Pool/pkg/pool/health_checker_test.go`
- No external dependency changes (`git diff go.mod` must be empty).
- Genuine implementation with thorough tests.

## Current Parent
- Conversation ID: 3c3be357-fb61-4d2b-9ed3-40099ef64f03
- Updated: 2026-07-26T09:00:46Z

## Task Summary
- **What to build**:
  - `rw_splitter.go`: `ClassifyQuery(sql string) QueryType` and `Route(...) Manager` with round-robin replica selection and fallback.
  - `health_checker.go`: `HealthChecker` wrapping `Manager`, validating connections on acquire with retry (up to 3 times), tracking `HealthStats`.
  - Comprehensive unit tests in `rw_splitter_test.go` and `health_checker_test.go`.
- **Success criteria**:
  - `go build ./...` and `go test ./...` in `packages/Pranor Pool` pass with exit code 0.
  - No changes in `go.mod`.
  - Hand-off and changes reports created in `.agents/worker_m4`.

## Change Tracker
- **Files modified**: [None yet]
- **Build status**: Untested
- **Pending issues**: None

## Quality Status
- **Build/test result**: Untested
- **Lint status**: Untested
- **Tests added/modified**: [TBD]

## Loaded Skills
- None loaded.

## Key Decisions Made
- [Initial state]

## Artifact Index
- `/home/developer/workspace/pranor/.agents/worker_m4/DISPATCH.md` — Dispatch prompt record
- `/home/developer/workspace/pranor/.agents/worker_m4/BRIEFING.md` — Current state & briefing
