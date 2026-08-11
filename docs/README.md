# Pranor Documentation

Welcome to the Pranor documentation. Pranor is a unified, modular backend infrastructure engine with its own programming language, designed to build high-performance microservices with zero glue code.

> 💡 **Documentation Version Selector**
> - 📘 **[Pranor v1.0 (Stable Docs)](/pranor/v1.0/)** — Core 16 Infrastructure Modules (Gate, Pulse, Vault, Auth, Mesh, etc.)
> - 🚀 **[Pranor v2.0 (AI Execution Fabric Docs)](/pranor/v2.0/)** — Governed AI Agent Execution Layer (`std/graph`, `std/decision`, `std/agent`, `std/memory`, `std/eval`, `std/flow`)


## Quick Navigation

| Section | Description |
|---------|-------------|
| [Getting Started](./getting-started.md) | Install Pranor and build your first service in 5 minutes |
| [Language Reference](./language/) | Syntax, standard library, CLI commands |
| [Module Docs](./modules/) | Full documentation for each Pranor module |
| [Deployment](./deployment/) | Docker, Kubernetes, standalone deployment guides |
| [Architecture](./architecture/) | System design, security model, observability |
| [Enterprise](./enterprise/) | EE features, licensing, and comparison |
| [Changelog](./changelog.md) | Unified release history |

## Modules

| Module | What it does | Docs |
|--------|-------------|------|
| **Pranor** (CLI) | Compiler & language runtime | [Language →](./language/syntax.md) |
| **Gate** | API Gateway & AI Guard | [gate.md →](./modules/gate.md) |
| **Pulse** | Async Event Broker & Message Queue | [pulse.md →](./modules/pulse.md) |
| **Vault** | S3 Storage & Vector Search | [vault.md →](./modules/vault.md) |
| **Chrono** | Distributed Job Scheduler | [chrono.md →](./modules/chrono.md) |
| **Auth** | Identity & Access Control | [auth.md →](./modules/auth.md) |
| **Cache** | Distributed Cache Engine | [cache.md →](./modules/cache.md) |
| **Mesh** | Service Discovery & Load Balancing | [mesh.md →](./modules/mesh.md) |
| **Trace** | Distributed Tracing Engine | [trace.md →](./modules/trace.md) |
| **Console** | Observability Dashboard | [console.md →](./modules/console.md) |
| **Pool** | Database Connection Proxy | [pool.md →](./modules/pool.md) |
| **Notify** | Email/Slack/SMS Gateway | [notify.md →](./modules/notify.md) |
| **Flow** | Workflow Engine & Saga Orchestrator | [flow.md →](./modules/flow.md) |
| **Deploy** | Docker/K8s Deployment Pipeline | [deploy.md →](./modules/deploy.md) |
| **Tunnel** | WebSocket Dev Tunneling | [tunnel.md →](./modules/tunnel.md) |
| **Hub** | Package Registry | [hub.md →](./modules/hub.md) |
| **Lock** | Distributed Locking | [lock.md →](./modules/lock.md) |
| **Secret** | Secret Management | [secret.md →](./modules/secret.md) |

## Install

```bash
# macOS/Linux
brew tap vyuvaraj/pranor && brew install pranor

# Windows
scoop bucket add pranor https://github.com/vyuvaraj/scoop-pranor
scoop install pranor

# From source
git clone https://github.com/vyuvaraj/pranor && cd pranor/lang && go build -o pranor .
```

## First Service

```bash
pranor init myapp && cd myapp && pranor run main.pnr --watch
```

## v2.0 AI Execution Fabric *(v2.0-dev — merges post v1.0 release)*

Pranor v2.0 extends the ecosystem with a governed AI agent execution layer built on top of the existing infrastructure. All v2.0 modules are CGO-free (`CGO_ENABLED=0`) and follow the OSS/EE build-tag convention.

| Module | Path | Description |
|--------|------|-------------|
| **Pranor Graph** | `std/graph` | Virtual entity context assembly — Hot/Warm/Cold 3-tier with fail-closed contract |
| **Pranor Decision** | `std/decision` | 6-level priority veto ladder: Auth > Budget > Risk > Rules > Learn > Default |
| **Pranor Learn** | `std/learn` | Pluggable ML inference provider (wazero WASM + gRPC sidecar) |
| **Pranor Eval** | `std/eval` | Trajectory replay and quality scoring — 4 evaluators (accuracy, latency, cost, safety) |
| **Trace Schema** | `std/trace` | Canonical OTLP span hierarchy + mandatory attribute contract for all modules |
| **Flow AgentStep** | `std/flow` | AgentStep interface + Saga runner + HITL approval queue |

> **Branch:** All v2.0 features live on `v2.0-dev`. See [v2.0 modules docs](./modules/graph.md) for full API reference.
