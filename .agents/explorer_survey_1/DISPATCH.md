## 2026-07-26T08:57:56Z
You are Explorer 1 (Survey phase).
Working directory: /home/developer/workspace/serv/.agents/explorer_survey_1

Your task:
1. Read `/home/developer/workspace/serv/.agents/ORIGINAL_REQUEST.md` (specifically requirements R1, R2, R3, R4).
2. Investigate the codebase at `/home/developer/workspace/serv/packages/Pranor Auth` and `/home/developer/workspace/serv/packages/Pranor Cache`.
3. Check existing package structure, go.mod files, imports, exported interfaces/structs, build/test setups (`go build ./...`, `go test ./...`), and existing tests.
4. Detail requirements for:
   - SA.G1: Opaque session token store (`packages/Pranor Auth/pkg/sessions/token_store.go`)
   - SA.G6: Credential stuffing velocity limiter (`packages/Pranor Auth/pkg/security/velocity_limiter.go`)
   - SC.G3: Probabilistic Bloom filter (`packages/Pranor Cache/pkg/bloom/bloom.go`)
   - SC.G4: Tiered TTL policy engine (`packages/Pranor Cache/pkg/tieredttl/policy.go`)
5. Document existing code layout, missing packages/files, potential helper functions or existing types to integrate with, test suite setup, and design recommendations.
6. Write `analysis.md` and `handoff.md` in your working directory `/home/developer/workspace/serv/.agents/explorer_survey_1`.
7. Send a message to orchestrator when finished.
