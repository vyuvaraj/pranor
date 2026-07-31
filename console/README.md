# ServConsole

```bash
docker run -p 8083:8083 ghcr.io/vyuvaraj/servconsole:latest
```

`ServConsole` is the unified, premium management dashboard and observability console for the **Servverse** ecosystem. It provides a single pane of glass for managing ServGate, ServQueue, ServStore, ServMesh, ServCloud, ServTrace, ServFlow, and all other Servverse components — with a glassmorphic, real-time UI designed for power users.

---

## Table of Contents
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Dashboard Modules](#dashboard-modules)
- [API Endpoints](#api-endpoints)
- [Getting Started](#getting-started)
- [Configuration](#configuration)

---

## Key Features

### 🎛️ Unified Management
- **Single pane of glass**: Manage the entire Servverse stack from one premium UI
- **Glassmorphic dark UI**: Premium visual design with smooth animations and real-time data refresh
- **Multi-tab navigation**: Navigate between components in organized tabs without page reloads
- **Global `⌘K` search**: Fuzzy search across all Servverse resources — services, routes, queues, buckets, workflows, traces — instantly

### 🚪 API Gateway Management (ServGate)
- **Live route audits**: View, create, and delete proxy routes in real-time
- **WASM hot-swap interface**: Upload and activate WASM middleware modules without restarting ServGate
- **AI middleware audit panel**: Monitor Prompt Guard violations, Semantic Cache similarity hits, PII scrubbing events, AI cost per request
- **OpenAPI Swagger UI**: Interactive API documentation browser for all registered gateway routes
- **Circuit breaker status board**: Live open/half-open/closed state per route with SLO metrics

### 📨 Queue Inspector (ServQueue)
- **Topic browser**: Real-time topic list with message rates, partition counts, and replication status
- **Schema registry browser**: Browse, compare, and evolve message schemas
- **DLQ browser & one-click replay**: Inspect dead letter messages; replay individual or bulk messages with one click
- **Consumer group lag dashboard**: Per-consumer-group, per-partition offset lag visualization with historical trend

### 🗃️ Storage Inspector (ServStore)
- **Bucket browser**: Navigate bucket contents, upload/download files, manage object metadata
- **Vector index namespace browser**: Inspect HNSW graph stats, index namespaces, embedding coverage
- **Branch management**: Create, diff, and merge CoW bucket branches from the UI
- **Tiering policy editor**: Configure hot/warm/cold tiering rules visually

### 🔭 Observability & Telemetry
- **eBPF flamegraph telemetry**: Live CPU and memory flamegraph profiling from the kernel layer — visualized in-browser
- **OTel trace correlation**: Click from a slow request directly into its distributed trace waterfall
- **SLO burn rate alerts**: Real-time error budget burn rate dashboards per service, with fast/slow window indicators
- **Service topology live graph**: Interactive dependency map of all Servverse services with live traffic flow edges

### 🔥 Chaos Engineering Panel
- **Chaos control panel**: Design and trigger chaos experiments (latency injection, error rate simulation, network partition) across ServMesh nodes
- **Experiment lifecycle management**: Start, monitor, and abort experiments; view blast radius before triggering
- **Historical experiment log**: Full audit trail of past chaos events with impact metrics

### 🛎️ Alerts & Incidents
- **Alert rule management**: Define threshold and anomaly-based alert rules across all Servverse metrics
- **Incident timeline**: Structured incident management with severity triage, notes, and resolution tracking

### 🌿 Provisioning & Environments
- **Environment provisioner**: Create complete isolated Servverse environments (dev/staging/prod) with one click
- **Branch preview provisioner**: Automatically spin up ephemeral ServCloud environments per git branch for PR previews

### ⚙️ Customization
- **Theme selector**: Dark, light, and glassmorphism themes; custom accent color
- **Pinned dashboard widgets**: Pin any metric chart or panel to a personal dashboard
- **Custom keyboard shortcuts**: User-configurable keybindings for common operations

---

## Architecture

```
Browser
  │
  ├─── Glassmorphic UI (SPA)
  │       ├─── Global ⌘K Search
  │       ├─── Real-time WebSocket feeds
  │       └─── Multi-tab navigation
  │
  ▼
ServConsole Backend (Go)
  │
  ├─── /api/v1/gateway/*    → ServGate integration
  ├─── /api/v1/queue/*      → ServQueue integration
  ├─── /api/v1/storage/*    → ServStore integration
  ├─── /api/v1/mesh/*       → ServMesh integration
  ├─── /api/v1/trace/*      → ServTrace integration
  ├─── /api/v1/chaos/*      → Chaos control plane
  ├─── /api/v1/incidents/*  → Incident management
  ├─── /api/v1/search       → Global resource search
  └─── WebSocket /ws/feeds  → Live topology & metrics
```

---

## Dashboard Modules

| Module | Description |
|--------|-------------|
| **Gateway Inspector** | ServGate routes, WASM modules, circuit breakers, AI middleware stats |
| **Queue Inspector** | Topic browser, consumer lag, DLQ management, schema registry |
| **Storage Inspector** | Bucket browser, vector index namespaces, branch management |
| **Topology Graph** | Live service dependency graph with traffic flow visualization |
| **Flamegraph Profiler** | eBPF-powered CPU/memory flamegraph per service |
| **Chaos Panel** | Design, trigger, and monitor chaos experiments |
| **SLO Dashboard** | Error budget burn rate, SLO compliance per service |
| **Trace Explorer** | Distributed trace waterfall search and correlation |
| **Incident Manager** | Alert rules, incident triage, resolution tracking |
| **Provisioner** | Environment and branch preview management |
| **AI Cost Dashboard** | Per-service AI token spend, model routing savings |

---

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/search?q=` | Global resource search (⌘K) |
| `GET` | `/api/v1/topology/graph` | Live service topology graph data |
| `GET` | `/ws/topology` | WebSocket: real-time topology updates |
| `GET` | `/api/v1/flamegraph/{service}` | eBPF flamegraph for a service |
| `GET` | `/api/v1/slo/{service}/burn-rate` | SLO burn rate metrics |
| `POST` | `/api/v1/chaos/experiments` | Create a chaos experiment |
| `DELETE` | `/api/v1/chaos/experiments/{id}` | Abort a chaos experiment |
| `GET` | `/api/v1/incidents` | List active incidents |
| `POST` | `/api/v1/incidents` | Create an incident |
| `GET` | `/api/v1/queue/dlq/{topic}` | DLQ browser |
| `POST` | `/api/v1/queue/dlq/{topic}/replay` | One-click DLQ replay |
| `GET` | `/api/v1/queue/consumers/{group}/lag` | Consumer lag per group |
| `POST` | `/api/v1/environments` | Provision an environment |
| `POST` | `/api/v1/branch-preview` | Provision a branch preview |
| `GET` | `/api/v1/preferences` | Get user preferences |
| `PUT` | `/api/v1/preferences` | Update user preferences |

---

## Getting Started

```bash
docker run -p 8083:8083 \
  -e SERVCONSOLE_SERVGATE_URL=http://servgate:8080 \
  -e SERVCONSOLE_SERVQUEUE_URL=http://servqueue:9090 \
  -e SERVCONSOLE_SERVSTORE_URL=http://servstore:7070 \
  -e SERVCONSOLE_SERVTRACE_URL=http://servtrace:4318 \
  -e SERVCONSOLE_SERVMESH_URL=http://servmesh:8095 \
  ghcr.io/vyuvaraj/servconsole:latest
```

Open `http://localhost:8083` in your browser.

---

## Configuration

| Variable | Description |
|----------|-------------|
| `SERVCONSOLE_PORT` | HTTP port (default: `8083`) |
| `SERVCONSOLE_SERVGATE_URL` | ServGate backend URL |
| `SERVCONSOLE_SERVQUEUE_URL` | ServQueue backend URL |
| `SERVCONSOLE_SERVSTORE_URL` | ServStore backend URL |
| `SERVCONSOLE_SERVTRACE_URL` | ServTrace OTLP URL |
| `SERVCONSOLE_SERVMESH_URL` | ServMesh backend URL |
| `SERVCONSOLE_AUTH_TOKEN` | Static admin auth token |
