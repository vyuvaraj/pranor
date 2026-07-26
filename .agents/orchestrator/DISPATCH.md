## 2026-07-26T08:57:09Z
You are the Project Orchestrator for the Go monorepo at `/home/developer/workspace/serv`.

Your objective is to orchestrate the implementation of 10 pending OSS roadmap items across 5 Servverse modules (ServAuth, ServCache, ServCron, ServPool, ServQueue) as specified in `/home/developer/workspace/serv/.agents/ORIGINAL_REQUEST.md`.

Your working directory is `/home/developer/workspace/serv/.agents/orchestrator`.
Please create your plan (`plan.md`), track progress (`progress.md`), and maintain state (`context.md`) in your folder.

Requirements summary:
1. SA.G1: ServAuth — Opaque Session Token Store (`packages/ServAuth/pkg/sessions/token_store.go`)
2. SA.G6: ServAuth — Credential Stuffing Velocity Limiter (`packages/ServAuth/pkg/security/velocity_limiter.go`)
3. SC.G3: ServCache — Probabilistic Bloom Filter (`packages/ServCache/pkg/bloom/bloom.go`)
4. SC.G4: ServCache — Tiered TTL Policy Engine (`packages/ServCache/pkg/tieredttl/policy.go`)
5. CR.G1: ServCron — DAG Job Chain Pipeline (extend `packages/ServCron/pkg/cron/cron.go`)
6. CR.G2: ServCron — Per-Job Retry Policy (extend `packages/ServCron/pkg/cron/cron.go`)
7. CR.G4: ServCron — YAML Cron-as-Code (`packages/ServCron/pkg/config/jobs_loader.go`)
8. SP.G1: ServPool — Read/Write Split Router (`packages/ServPool/pkg/routing/rw_splitter.go`)
9. SP.G2: ServPool — Connection Health Validation (`packages/ServPool/pkg/pool/health_checker.go`)
10. SQ.G5: ServQueue — W3C Trace Context Propagation (`packages/ServQueue/pkg/tracing/traceparent.go` and `packages/ServQueue/pkg/core/engine.go`)

Ensure all builds (`go build ./...`) and tests (`go test ./...`) pass in each module, no external dependencies are added to go.mod, and git commit & push are performed upon completion.

When all work is completed and verified, report victory to me (the Sentinel).
