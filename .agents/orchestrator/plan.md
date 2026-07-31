# Project Implementation Plan: Pranor 10 OSS Roadmap Items

## Phase 0: Survey & Specification Mapping
- Spawn 3 parallel Explorers to map the 5 modules, existing codebase patterns, interfaces, types, go.mod dependencies, build/test commands, and detailed requirements for R1-R10.
  - Explorer 1 (`teamwork_preview_explorer_survey_1`): Pranor Auth (SA.G1, SA.G6) & Pranor Cache (SC.G3, SC.G4)
  - Explorer 2 (`teamwork_preview_explorer_survey_2`): Pranor Chrono (CR.G1, CR.G2, CR.G4)
  - Explorer 3 (`teamwork_preview_explorer_survey_3`): Pranor Pool (SP.G1, SP.G2) & Pranor Pulse (SQ.G5)
- Synthesize findings into `PROJECT.md § Feature Inventory` and architecture layout.

## Phase 1: Parallel Tracks Setup
- Track 1: Implementation Track (Milestone Decomposition by module / item)
- Track 2: E2E Testing Track (Requirement-driven test suite creation)

## Phase 2: Milestone Execution (Implementation & Unit Tests)
- Milestone 1 (M1): Pranor Auth (SA.G1, SA.G6)
- Milestone 2 (M2): Pranor Cache (SC.G3, SC.G4)
- Milestone 3 (M3): Pranor Chrono (CR.G1, CR.G2, CR.G4)
- Milestone 4 (M4): Pranor Pool (SP.G1, SP.G2)
- Milestone 5 (M5): Pranor Pulse (SQ.G5)
Each milestone follows the iteration loop: Explorer -> Worker -> Reviewer (2x) -> Challenger (2x) -> Forensic Auditor -> Gate.

## Phase 3: Final E2E Integration & Verification
- Final Milestone: Pass 100% E2E tests across all 5 modules.
- Phase 2 Coverage Hardening: Challenger stress testing.
- Verify `go build ./...` and `go test ./...` in all 5 packages.
- Verify git diff for `go.mod` (no new external dependencies).

## Phase 4: Git Commit & Push & Victory Report
- `git add ...`
- `git commit -m ...`
- `git push origin main`
- Sentinel Victory Report.
