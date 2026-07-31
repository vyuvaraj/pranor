# Pranor Deploy Roadmap

This roadmap outlines the planned development phases for the Pranor Deploy managed deployment platform.

---

## Differentiating Factors (Why Pranor Deploy vs. K8s/Heroku/Nomad?)
* **Zero-Config Infrastructure**: No Dockerfiles or K8s manifests required. Pranor Deploy compiles `.pnr` files directly and infers infrastructure needs from the code.
* **Auto-Routing Gateway Sync**: Deployed services register their routes instantly with `Pranor Gate` reverse-proxies, updating route mapping dynamically.
* **Built-in Telemetry**: Redirects standard output/error pipes to a memory ring buffer, enabling dashboard sync without external logging agents.

---

## Phase 1: Local Orchestrator MVP (Completed)
- [x] **Process Manager**: Spawns, monitors, and stops service processes dynamically.
- [x] **Go-compiler fallback**: Fallback mock server generation if native compiler is not available on path.
- [x] **Port Allocation**: Dynamically discovers and allocates free TCP ports to running services.

## Phase 2: API & Gateway Integration (Completed)
- [x] **REST API**: JSON endpoints for deployments, listing status, logs retrieval, and service deletion.
- [x] **Route Registration Sync**: Auto-updates API Gateways (like Pranor Gate) on new deployments.
- [x] **Console log capture**: Standard output and error redirecting to in-memory ring buffer.

## Phase 3: Telemetry & Console Integration (Completed)
- [x] **Pranor Console Dashboard**: Expose deployment history, rollbacks, and active process graphs in the console.
- [x] **Health Monitoring**: Periodically ping running services and flag unhealthy processes.
- [x] **CPU/Memory stats**: Query system OS metrics for resource consumption monitoring.

## Phase 4: Production Isolation & Security (Planned)
- [x] **WASM Isolation**: Direct execution of compiled WASM targets in-process for sandbox isolation. [June 29, 2026]
- [x] **Docker Engine runner**: Spin up individual services in isolated Docker containers instead of native processes. [June 29, 2026]
- [x] **Shared OIDC Authentication**: Enforce bearer token validation via shared `PRANOR_JWT_SECRET`. [June 29, 2026]


## Phase 5: Production PaaS Features (Next Level — Pending)
- [ ] **Resource Quotas & Limits**: Per-deployment CPU/memory caps with OOM protection.
- [ ] **Secret Injection from Pranor Vault**: Resolve `${{secrets.KEY}}` references from encrypted Pranor Vault bucket at deploy time. Rotate secrets without redeployment.
- [ ] **Pranor Auth Integration**: Auto-provision Pranor Auth OIDC configuration for deployed services. Services get identity management out of the box.
- [ ] **Build Packs**: Auto-detect project type (Go, Node, Python, Pranor) and build without user-provided Dockerfile.
- [ ] **Deployment Previews**: Branch-based preview deployments with unique URLs (like Vercel previews).
- [ ] **Horizontal Auto-scaling**: Scale instances up/down based on request rate from Pranor Gate metrics.
- [ ] **Integrated CI Pipeline**: Run `pranor test` before deploy. Reject deploys that fail tests.
- [ ] **Multi-region Deployment**: Deploy to multiple regions with Pranor Mesh-based global load balancing.

## Phase 6: Architectural Depth & DevOps (Pending)
- [ ] **GitOps Deployment Sync** — Trigger deploys automatically on git push via webhook; store deployment manifest in repository for auditability (OPS.5)
- [ ] **`pranor cloud diff`** — Preview infrastructure changes (environment vars, resources, routes) before applying a deploy — like `terraform plan` for Pranor Deploy (DevOps)
- [ ] **Deploy Annotations** — Annotate each deploy with commit SHA, author, and changelog; surface in Pranor Console timeline and in Pranor Trace spans for change correlation (DX)
- [ ] **Local `pranor cloud emulate`** — Emulate the full production deploy pipeline locally: health checks, rolling update, rollback — catching breakage before pushing (DX)

> See [UNIFIED_ROADMAP.md](../pranor-repo/UNIFIED_ROADMAP.md) for the full ecosystem priority matrix.


---

## Phase 7: Test Coverage (Pending — Phase 22)

> **Issue:** Only 7 test functions in 1 file.

| # | Item | Effort | Description | Status |
|---|------|--------|-------------|--------|
| 7.1 | **Expand test suite** | Medium | From 7 → 30+ test functions: deploy lifecycle, port allocation conflicts, health monitoring recovery, gateway route sync, Docker runner isolation, rollback | [ ] |
