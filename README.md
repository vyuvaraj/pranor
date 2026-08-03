# Pranor

> **The Unified & Modular Open-Source Backend Infrastructure Engine**

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8.svg?style=flat&logo=go)](https://golang.org)
[![Modules](https://img.shields.io/badge/Modules-19-purple.svg)](#component-catalog)

**Pranor** is a complete, modular backend infrastructure engine designed to build high-performance microservices with zero glue code. It provides a programming language, API gateway, message broker, S3 object store, cache engine, identity provider, and distributed tracing dashboard.

Deploy as a **unified platform** using `pranor run`, or run any component **independently as a standalone Go/WASM binary** inside existing Node.js, Python, Java, Go, or Rust microservice stacks.

---

## System-Wide Ecosystem Architecture Topology

```mermaid
graph TD
    classDef client fill:#1e293b,stroke:#38bdf8,stroke-width:2px,color:#fff;
    classDef ingress fill:#0f172a,stroke:#0d9488,stroke-width:2px,color:#fff;
    classDef services fill:#1e1b4b,stroke:#6366f1,stroke-width:2px,color:#fff;
    classDef storage fill:#0f172a,stroke:#38bdf8,stroke-width:2px,color:#fff;
    classDef ops fill:#1e293b,stroke:#64748b,stroke-width:1px,color:#fff;

    subgraph ExternalIngress ["🌐 External Entrypoints & Tunnels"]
        WebClient["Web Browser / Mobile App"] :::client
        ExternalWebhook["Stripe / GitHub Webhook"] :::client
        TunnelRelay["Pranor Tunnel Daemon<br/><i>(Local Dev Exposure)</i>"] :::client
    end

    subgraph SecurityIngress ["🛡️ Zero-Trust Ingress & Security Tier"]
        Gate["Pranor Gate API Gateway<br/><i>(eBPF DDoS / WASM / AI Guard)</i>"] :::ingress
        Auth["Pranor Auth Engine<br/><i>(OAuth2 / OIDC / WebAuthn / SPIFFE)</i>"] :::ingress
        Secret["Pranor Secret Manager<br/><i>(AES-256-GCM / HSM)</i>"] :::ingress
    end

    subgraph MicroservicesTier ["⚡ Microservices Mesh & Orchestration"]
        Mesh["Pranor Mesh Overlay<br/><i>(P2C Load Balancer & WireGuard)</i>"] :::services
        Flow["Pranor Flow Orchestrator<br/><i>(Saga Workflows & DAG)</i>"] :::services
        Chrono["Pranor Chrono Scheduler<br/><i>(Sub-ms TimeWheel Cron)</i>"] :::services
        Deploy["Pranor Deploy Engine<br/><i>(Canary & Preview Envs)</i>"] :::services
    end

    subgraph EventAndStorageTier ["💾 State, Caching & Data Layer"]
        Pulse["Pranor Pulse Event Broker<br/><i>(STOMP / Kafka / MQTT)</i>"] :::storage
        Vault["Pranor Vault Object Store<br/><i>(S3 REST & HNSW Vector Engine)</i>"] :::storage
        Cache["Pranor Cache Engine<br/><i>(SIMD Vector & Redis Grid)</i>"] :::storage
        Pool["Pranor Pool DB Proxy<br/><i>(RDBMS Read/Write Split)</i>"] :::storage
        Lock["Pranor Lock Engine<br/><i>(Fencing Token Leases)</i>"] :::storage
    end

    subgraph ObservabilityOps ["🔭 Central Observability & Control Plane"]
        Console["Pranor Console UI & TUI<br/><i>(Live Telemetry & ⌘K Search)</i>"] :::ops
        Trace["Pranor Trace Collector<br/><i>(OTLP Spans & eBPF Flamegraphs)</i>"] :::ops
        Notify["Pranor Notify Gateway<br/><i>(Email / SMS / WebPush)</i>"] :::ops
        Hub["Pranor Hub Registry<br/><i>(Package & WASM Registry)</i>"] :::ops
    end

    WebClient --> Gate
    ExternalWebhook --> Gate
    TunnelRelay --> Gate
    Gate --> Auth
    Auth --> Secret
    Gate --> Mesh
    Mesh --> Flow
    Mesh --> Chrono
    Flow --> Pulse
    Pulse --> Vault
    Flow --> Cache
    Mesh --> Pool
    Flow --> Lock
    Gate -.-> Trace
    Mesh -.-> Trace
    Console -.-> Trace
    Console -.-> Gate
    Console -.-> Pulse
    Console -.-> Vault
    Deploy --> Hub
```

---

## Unified Platform vs. Standalone Tools

| Mode | Deployment | Best For |
|---|---|---|
| **Unified Platform** | Run all modules in concert via `pranor dev main.pnr` or `pranor deploy` | Greenfield backend services, all-in-one API stacks, rapid prototyping without glue code. |
| **Standalone Component** | Build & deploy any individual module (e.g. `Pranor Gate`, `Pranor Vault`, `Pranor Pulse`) as a zero-dependency Go binary | Existing stacks needing an AI API Gateway, S3 vector search, or inline WASM stream processing. |

---

## Component Catalog

All modules live at the root level and are integrated into a single Go Workspace (`go.work`):

| Component | Path | Description | Key Features |
|---|---|---|---|
| **Pranor** (CLI) | [`lang/`](./lang) | Compiler & Language Runtime | Domain-specific backend language compiling to native Go binaries. |
| **Pranor Gate** | [`gate/`](./gate) | API Gateway & AI Guard | WASM reverse proxy, AI Prompt Guard, PII redaction, Circuit Breaker, Rate limiting. |
| **Pranor Pulse** | [`pulse/`](./pulse) | Async Event Broker | Compute-in-Queue WASM stream transforms, DLQ auto-offloading, memory backpressure. |
| **Pranor Vault** | [`vault/`](./vault) | S3 Storage + Vector Search | S3-compatible object storage with embedded TF-IDF/vector search and time-travel versioning. |
| **Pranor Cache** | [`cache/`](./cache) | Distributed Cache Engine | Dual-mode memory/Redis caching with TTL, sliding eviction, and OTel metrics. |
| **Pranor Auth** | [`auth/`](./auth) | Identity & Access Provider | OAuth2/OIDC provider, multi-tenant RBAC, MFA, JWT validation middleware. |
| **Pranor Console** | [`console/`](./console) | Observability Dashboard | Central web dashboard, metrics visualizer, SQL workbench, incident analyzer. |
| **Pranor Mesh** | [`mesh/`](./mesh) | Library Service Mesh | Client-side service discovery, load balancing, retries, circuit breaking. |
| **Pranor Chrono** | [`chrono/`](./chrono) | Distributed Job Scheduler | Multi-node leader election, cron schedule parser, persistent state. |
| **Pranor Deploy** | [`deploy/`](./deploy) | Deployment Orchestrator | Single-command Docker, Kubernetes, and TLS certificate deployment pipeline. |
| **Pranor Trace** | [`trace/`](./trace) | Distributed Tracing Engine | OTLP trace collector, waterfall UI, trace anomaly detection. |
| **Pranor Tunnel** | [`tunnel/`](./tunnel) | WebSocket Dev Tunneling | WebSocket relay server, custom subdomain routing, HTTP traffic inspector. |
| **Pranor Pool** | [`pool/`](./pool) | Database Connection Proxy | Connection pooling, SQL query analytics, read/write splitting. |
| **Pranor Notify** | [`notify/`](./notify) | Notification Gateway | Transactional email, Slack, SMS delivery with HTML templating. |
| **Pranor Flow** | [`flow/`](./flow) | Workflow Engine | Stateful DAG execution, saga orchestrator, human approval gates. |
| **Pranor Hub** | [`hub/`](./hub) | Package Registry | Semver resolution, artifact signing, module registry service. |
| **Pranor Core** | [`core/`](./core) | Shared Libraries | Resilient retry, health checks, OTel middleware. |
| **Pranor Lock** | [`lock/`](./lock) | Distributed Locking | Acquire, renew, inspect distributed locks. |
| **Pranor Secret** | [`secret/`](./secret) | Secret Management | Dynamic secret injection, Shamir key unsealing. |

---

## Quickstart

### Prerequisites
* **Go 1.22+** installed on system.

### Build the Compiler
```bash
git clone https://github.com/vyuvaraj/pranor.git
cd pranor

# Build compiler binary
cd lang
go build -o pranor main.go
```

### Write Your First Service
```bash
pranor init myapp
cd myapp
pranor run main.pnr --watch
```

### Install via Package Manager

**macOS/Linux (Homebrew):**
```bash
brew tap vyuvaraj/pranor
brew install pranor
```

**Windows (Scoop):**
```powershell
scoop bucket add pranor https://github.com/vyuvaraj/scoop-pranor
scoop install pranor
```

---

## Enterprise Edition (EE)

For enterprise compliance, multi-tenant RBAC, audited Raft clustering, and dedicated OIDC/SAML integration, see the **Pranor Enterprise Edition**:
* **Enterprise Repo:** [`github.com/vyuvaraj/pranor-ee`](https://github.com/vyuvaraj/pranor-ee) *(Private / License Required)*

---

## License

This monorepo is open-source software licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**. See the [LICENSE](./LICENSE) file for full details.
