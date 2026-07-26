# ServCloud

```bash
docker run -p 8088:8088 ghcr.io/vyuvaraj/servcloud:latest
```

`ServCloud` is the managed deployment platform and process orchestrator for the **Servverse** ecosystem. It provides PaaS-style service deployment, blue/green and canary strategies, per-branch preview environments, container isolation, and deep integration with `ServGate` for automatic routing registration.

---

## Table of Contents
- [Key Features](#key-features)
- [Architecture](#architecture)
- [API Endpoints](#api-endpoints)
- [Deployment Strategies](#deployment-strategies)
- [Preview Environments](#preview-environments)
- [Getting Started](#getting-started)

---

## Key Features

### 🚀 Core Deployment Platform
- **PaaS deployment API**: Compile and run `.srv` background services on demand via REST API
- **Process isolation**: Dedicated port allocation per deployment; process metrics tracking
- **Dynamic gateway routing registration**: Newly deployed services are automatically registered with `ServGate` — zero manual route configuration
- **Ring-buffer log streaming**: Capture stdout/stderr into a ring buffer; stream logs via REST API
- **OTel tracing**: Deep integration with `ServTrace` via shared tracing — per-deployment spans

### 🔵🟢 Blue/Green Deployment
- **Zero-downtime traffic switch**: Atomic cutover — ServGate switches 100% of traffic to new (green) deployment in a single atomic update
- **Instant rollback**: If issues arise, switch back to blue with one API call
- **Health gate**: Green deployment must pass health checks before cutover is triggered
- **Audit trail**: Every cutover and rollback event logged with timestamp and operator identity

### 🐤 Canary Deployment
- **Configurable traffic split**: Route a percentage (e.g., 5%, 10%, 25%) of traffic to the canary deployment
- **Automatic rollback**: Monitor error rate on canary; if it exceeds configurable threshold, automatically revert 100% traffic to stable
- **Progressive promotion**: Incrementally increase canary traffic weight on success (5% → 25% → 50% → 100%)
- **ServGate integration**: Traffic split is enforced by ServGate's weighted routing — no client-side changes required

### 🌿 Preview Environments
- **Per-branch preview provisioner**: Automatically create complete isolated Servverse environments per git branch — ideal for PR review workflows
- **Ephemeral lifecycle**: Preview environments are automatically cleaned up when the branch is deleted or after a configurable TTL
- **Independent routing**: Each preview gets its own ServGate subdomain (e.g., `feature-x.preview.servverse.net`)
- **Full stack provisioning**: Preview environments include isolated ServQueue, ServStore, and ServCache instances

### 🐳 Container Isolation
- **Docker/OCI container mode**: Deploy services as fully isolated containers (via Docker or OCI runtime) rather than raw processes
- **Resource limits**: Configure per-container CPU and memory limits
- **Network isolation**: Container deployments run in isolated bridge networks

---

## Architecture

```
Developer API Request
        │ POST /api/v1/deployments
        ▼
┌───────────────────────────────────────────────┐
│                 ServCloud                      │
│                                               │
│  ┌────────────────────────────────────────┐   │
│  │  Deployment Orchestrator               │   │
│  │  Build → Deploy → Health Check         │   │
│  └───────────┬────────────────────────────┘   │
│              │                                │
│  ┌───────────▼────────────────────────────┐   │
│  │  Strategy Manager                      │   │
│  │  Direct │ Blue/Green │ Canary           │   │
│  └───────────┬────────────────────────────┘   │
│              │                                │
│  ┌───────────▼────────────────────────────┐   │
│  │  ServGate Registration                 │   │
│  │  (auto-register routes on deploy)      │   │
│  └────────────────────────────────────────┘   │
│                                               │
│  ┌────────────────────┐  ┌─────────────────┐  │
│  │  Log Streamer       │  │ Preview Env Mgr │  │
│  │  (ring buffer)      │  │ (branch → env)  │  │
│  └────────────────────┘  └─────────────────┘  │
└───────────────────────────────────────────────┘
```

---

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/deployments` | Deploy a service (direct, blue/green, or canary) |
| `GET` | `/api/v1/deployments` | List all deployments |
| `GET` | `/api/v1/deployments/{id}` | Get deployment status and metrics |
| `POST` | `/api/v1/deployments/{id}/promote` | Promote canary to stable |
| `POST` | `/api/v1/deployments/{id}/rollback` | Roll back to previous version |
| `POST` | `/api/v1/deployments/{id}/cutover` | Blue/Green: cut all traffic to new version |
| `GET` | `/api/v1/deployments/{id}/logs` | Stream deployment logs (ring buffer) |
| `DELETE` | `/api/v1/deployments/{id}` | Stop and remove a deployment |
| `POST` | `/api/v1/previews` | Create a preview environment for a branch |
| `GET` | `/api/v1/previews` | List active preview environments |
| `DELETE` | `/api/v1/previews/{id}` | Destroy a preview environment |
| `/metrics` | `GET` | Prometheus metrics (deployments active, error rates, rollback events) |
| `/healthz` | `GET` | Liveness probe |

---

## Deployment Strategies

### Direct Deploy
```bash
curl -X POST http://servcloud:8088/api/v1/deployments \
  -d '{"service": "orders-api", "image": "ghcr.io/myorg/orders:v2.1.0", "strategy": "direct", "port": 3000}'
```

### Blue/Green Deploy
```bash
# Deploy green (new version)
curl -X POST http://servcloud:8088/api/v1/deployments \
  -d '{"service": "orders-api", "image": "ghcr.io/myorg/orders:v2.2.0", "strategy": "blue-green"}'
# → { "id": "dep-456", "status": "green-standby", "green_url": "http://green-orders:3001" }

# Cut over all traffic to green
curl -X POST http://servcloud:8088/api/v1/deployments/dep-456/cutover
# → ServGate atomically switches all /api/orders traffic to green

# Rollback if needed
curl -X POST http://servcloud:8088/api/v1/deployments/dep-456/rollback
```

### Canary Deploy
```bash
# Deploy canary at 5% traffic
curl -X POST http://servcloud:8088/api/v1/deployments \
  -d '{
    "service": "orders-api",
    "image": "ghcr.io/myorg/orders:v2.3.0",
    "strategy": "canary",
    "canary_weight": 5,
    "auto_rollback_error_rate": 0.05
  }'

# Progressive promotion: 5% → 25% → 50% → 100%
curl -X POST http://servcloud:8088/api/v1/deployments/dep-789/promote \
  -d '{"weight": 25}'
```

---

## Preview Environments

```bash
# Create preview environment for a feature branch
curl -X POST http://servcloud:8088/api/v1/previews \
  -d '{"branch": "feature/new-checkout", "ttl": "7d"}'
# → { "id": "prev-001", "url": "https://feature-new-checkout.preview.servverse.net", "expires_at": "..." }

# Destroy preview
curl -X DELETE http://servcloud:8088/api/v1/previews/prev-001
```

---

## Getting Started

```bash
docker run -p 8088:8088 \
  -e SERVCLOUD_SERVGATE_URL=http://servgate:8080 \
  -e SERVCLOUD_OTEL_ENDPOINT=http://servtrace:4318 \
  -e SERVCLOUD_CONTAINER_RUNTIME=docker \
  -v /var/run/docker.sock:/var/run/docker.sock \
  ghcr.io/vyuvaraj/servcloud:latest
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVCLOUD_PORT` | `8088` | HTTP listener port |
| `SERVCLOUD_SERVGATE_URL` | — | ServGate URL for route registration |
| `SERVCLOUD_OTEL_ENDPOINT` | — | OpenTelemetry collector URL |
| `SERVCLOUD_CONTAINER_RUNTIME` | `process` | `process` (raw) or `docker` (OCI container) |
| `SERVCLOUD_PREVIEW_DOMAIN` | — | Base domain for preview environments |
| `SERVCLOUD_PREVIEW_TTL` | `7d` | Default preview environment TTL |
