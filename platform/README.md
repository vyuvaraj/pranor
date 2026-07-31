# Pranor Platform

`Pranor Platform` is the unified platform layer for the **Pranor** ecosystem — providing a single-binary embedded monolith runtime (`pranord`), platform-wide chaos injection, a cluster administration CLI (`pranorctl`), unified health/readiness APIs, and production deployment manifest generation (Docker Compose & Helm).

---

## Table of Contents
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Components](#components)
  - [pranord — Single-Binary Unified Runtime](#pranord--single-binary-unified-runtime)
  - [Unified Chaos Injection Engine](#unified-chaos-injection-engine)
  - [pranorctl — Cluster Administration CLI](#pranorctl--cluster-administration-cli)
  - [Unified Health & Readiness API](#unified-health--readiness-api)
  - [Distribution Manifest Generator](#distribution-manifest-generator)
- [Getting Started](#getting-started)

---

## Key Features

- **Single-binary `pranord` runtime**: Embed all Pranor components into one unified process for local development and small deployments
- **Unified chaos engine**: Inject faults (network, CPU, memory, disk, clock skew) platform-wide from a single API
- **`pranorctl` CLI**: Cluster-wide administration — list services/nodes, restart services, apply config
- **Unified health API**: `/health`, `/ready`, and `/api/v1/platform/health/rollup` endpoints aggregating all component health
- **Deployment manifest generator**: Programmatically generate `docker-compose.yml` and Helm `values.yaml` for production deployments

---

## Architecture

```
Pranor Platform
├── pkg/platform/     → pranord unified runtime (component registry, start/stop/shutdown)
├── pkg/chaos/        → Unified chaos injection engine (network/CPU/memory/disk/clock)
├── pkg/cli/          → pranorctl cluster administration CLI
├── pkg/health/       → Unified health, readiness & rollup metrics API
└── pkg/distribution/ → Docker Compose & Helm chart manifest generator
```

---

## Components

### pranord — Single-Binary Unified Runtime

`pranord` allows running the entire Pranor stack as a single embedded binary — ideal for local development, CI/CD pipelines, and small self-hosted deployments.

**API: `GET /api/v1/pranord/components`**

```json
{
  "components": [
    { "name": "pranor-gate",   "running": true },
    { "name": "pranor-pulse",  "running": true },
    { "name": "pranor-vault",  "running": true },
    { "name": "pranor-mesh",   "running": false },
    { "name": "pranor-trace",  "running": true }
  ]
}
```

**Usage in code:**

```go
rt := platform.NewServdRuntime()
rt.RegisterComponent("pranor-gate")
rt.RegisterComponent("pranor-pulse")
rt.RegisterComponent("pranor-vault")

rt.StartComponent("pranor-gate")
rt.StartComponent("pranor-pulse")

// Graceful shutdown
defer rt.Shutdown(ctx)
```

---

### Unified Chaos Injection Engine

Inject faults platform-wide from a single API, targeting individual nodes or services.

**Fault types:** `network` · `cpu` · `memory` · `disk` · `clock_skew`

**API: `POST /api/v1/platform/chaos/faults`**

```bash
# Inject 30% network packet loss on node-1 for 60 seconds
curl -X POST http://servplatform:8096/api/v1/platform/chaos/faults \
  -d '{
    "kind": "network",
    "target_node": "node-1",
    "intensity": 0.3,
    "duration": 60000000000
  }'
# → { "id": "fault-1", "kind": "network", "target_node": "node-1", "intensity": 0.3, "active": true }

# Inject CPU spike at 80% intensity on node-2
curl -X POST http://servplatform:8096/api/v1/platform/chaos/faults \
  -d '{"kind": "cpu", "target_node": "node-2", "intensity": 0.8, "duration": 30000000000}'

# Abort a fault
curl -X DELETE http://servplatform:8096/api/v1/platform/chaos/faults/fault-1

# List active faults
curl http://servplatform:8096/api/v1/platform/chaos/faults
```

**Intensity:** `0.0` (no impact) → `1.0` (full impact, e.g., 100% packet loss, max CPU load)

---

### pranorctl — Cluster Administration CLI

`pranorctl` is the Pranor cluster-wide administration CLI. It communicates with Pranor Platform to manage the entire stack.

```bash
# List all services
pranorctl get services
# → ["pranor-gate", "pranor-pulse", "pranor-vault", "pranor-mesh", "pranor-trace"]

# List cluster nodes
pranorctl get nodes
# → ["node-1", "node-2", "node-3"]

# Restart a service
pranorctl restart service payment-api
# → service 'payment-api' restarted successfully

# Apply cluster-wide configuration
pranorctl apply config --file cluster-config.json

# JSON output mode (for scripting)
pranorctl --json get services
```

**Available commands:**

| Command | Description |
|---------|-------------|
| `pranorctl get services` | List all registered services |
| `pranorctl get nodes` | List all cluster nodes |
| `pranorctl restart service <name>` | Restart a named service |
| `pranorctl apply config` | Apply cluster-wide configuration |

---

### Unified Health & Readiness API

Aggregate health across all platform components with a single endpoint.

**Endpoints:**

| Path | Description |
|------|-------------|
| `GET /health` | Liveness probe — `200 OK` if all components healthy, `503` if any failing |
| `GET /ready` | Readiness probe — `200 OK` only if all components registered and healthy |
| `GET /api/v1/platform/health/rollup` | Full per-component breakdown |

```bash
# Liveness check
curl http://servplatform:8096/health
# → { "ok": true }   HTTP 200 (all healthy)
# → { "ok": false }  HTTP 503 (degraded)

# Full rollup
curl http://servplatform:8096/api/v1/platform/health/rollup
# → {
#     "healthy": false,
#     "total": 5, "passing": 4, "failing": 1,
#     "components": [
#       { "name": "pranor-gate",  "healthy": true,  "latency_ns": 2000000 },
#       { "name": "pranor-vault", "healthy": false, "message": "disk full" },
#       ...
#     ]
#   }
```

**Reporting health from a component:**

```go
api := health.NewUnifiedHealthAPI()
api.ReportHealth(health.ComponentHealth{
    Name:    "pranor-gate",
    Healthy: true,
    Latency: 2 * time.Millisecond,
})
```

---

### Distribution Manifest Generator

Programmatically generate production deployment manifests for Docker Compose and Helm.

**Docker Compose:**

```go
gen := distribution.NewDistributionGenerator()
components := []distribution.ComponentSpec{
    {Name: "pranor-gate",  Image: "ghcr.io/vyuvaraj/pranor-gate:latest",  Port: 8080, Replicas: 2, EnvVars: map[string]string{"LOG_LEVEL": "info"}},
    {Name: "pranor-pulse", Image: "ghcr.io/vyuvaraj/pranor-pulse:latest", Port: 9090, Replicas: 3},
    {Name: "pranor-vault", Image: "ghcr.io/vyuvaraj/pranor-vault:latest", Port: 7070, Replicas: 2},
}

compose := gen.GenerateDockerCompose(components)
// compose.Content:
// version: '3.8'
// services:
//   pranor-gate:
//     image: ghcr.io/vyuvaraj/pranor-gate:latest
//     ports:
//       - "8080:8080"
//     environment:
//       LOG_LEVEL: info
//   ...
```

**Helm Values:**

```go
helm := gen.GenerateHelmValues(components)
// helm.Content:
// # Pranor Production Helm Values
// global:
//   imageTag: latest
// services:
//   pranor-gate:
//     enabled: true
//     replicas: 2
//     image: ghcr.io/vyuvaraj/pranor-gate:latest
//     port: 8080
//   ...
```

---

## Getting Started

Pranor Platform is used as a Go library embedded within the Pranor monorepo.

```go
import (
    "github.com/vyuvaraj/pranor/Pranor Platform/pkg/platform"
    "github.com/vyuvaraj/pranor/Pranor Platform/pkg/health"
    "github.com/vyuvaraj/pranor/Pranor Platform/pkg/chaos"
    "github.com/vyuvaraj/pranor/Pranor Platform/pkg/cli"
    "github.com/vyuvaraj/pranor/Pranor Platform/pkg/distribution"
)
```

For `pranorctl`, run:

```bash
go run ./cmd/pranorctl/main.go get services
```

### Package Structure

| Package | Description |
|---------|-------------|
| `pkg/platform` | `ServdRuntime` — single-binary component registry |
| `pkg/chaos` | `UnifiedChaosEngine` — platform-wide fault injection |
| `pkg/cli` | `ServctlCLI` — cluster administration CLI |
| `pkg/health` | `UnifiedHealthAPI` — /health, /ready, /rollup |
| `pkg/distribution` | `DistributionGenerator` — Docker Compose & Helm output |
