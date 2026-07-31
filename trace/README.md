# ServTrace

```bash
docker run -p 8090:8090 ghcr.io/vyuvaraj/servtrace:latest
```

`ServTrace` is the distributed tracing and continuous profiling service for the **Servverse** ecosystem. It ingests OTLP-format traces, assembles waterfall hierarchies, provides SLO burn rate alerting, and delivers eBPF-powered flamegraph profiling with automatic OTel correlation.

---

## Table of Contents
- [Key Features](#key-features)
- [Architecture](#architecture)
- [API Endpoints](#api-endpoints)
- [SLO Burn Rate Alerting](#slo-burn-rate-alerting)
- [Flamegraph Profiling](#flamegraph-profiling)
- [Getting Started](#getting-started)

---

## Key Features

### 📡 OTLP Ingestion & Span Assembly
- **OTLP/HTTP ingestion**: Standard `/v1/traces` endpoint compatible with all OpenTelemetry SDKs and collectors
- **Trace reassembly**: Groups spans by trace ID, links parent-child relationships, calculates absolute and relative duration offsets
- **Waterfall hierarchy tree**: Full span waterfall with nested children, duration bars, and critical path highlighting
- **Configurable in-memory store**: Thread-safe store with oldest-first trace eviction at configurable capacity

### 🔥 eBPF Flamegraph Profiling
- **Continuous eBPF CPU & memory profiler**: Kernel-level profiling via eBPF — no code instrumentation required
- **OTel trace-to-flamegraph correlator**: Automatically correlates a slow trace span to the flamegraph profile captured during that span's execution window
- **In-browser flamegraph visualization**: Interactive flamegraph rendered in ServConsole — click to zoom, search symbol names

### 📊 SLO & Error Budget
- **SLO burn rate alert engine**: Configurable SLO targets (e.g. 99.9% availability) with dual burn rate windows
  - **Fast burn window** (1h): Catches sudden spikes consuming error budget rapidly
  - **Slow burn window** (6h/24h): Catches gradual degradation
- **Error budget tracking**: Real-time remaining error budget per service per SLO definition
- **ServConsole integration**: Live SLO burn rate dashboard with alert status

### 📈 Prometheus Exemplars
- **Exemplar-linked OpenMetrics generator**: Produces Prometheus-compatible OpenMetrics text with `# TYPE` / `# UNIT` annotations and trace exemplar links embedded in histogram observations

### 🗺️ Distributed Dependency Analysis
- **Critical path analyzer**: Identifies the longest-latency path across a distributed trace — pinpoints bottleneck services
- **Distributed dependency map**: Builds a service-call graph from observed trace data; visualized in ServConsole topology view

---

## Architecture

```
OTLP SDK (Go/Python/JS/...)
     │ POST /v1/traces
     ▼
┌──────────────────────────────────────────┐
│               ServTrace                   │
│                                          │
│  ┌────────────────────────────────────┐  │
│  │   Span Ingestion & Reassembly      │  │
│  │   (Group by TraceID, Link parents) │  │
│  └─────────────┬──────────────────────┘  │
│                │                         │
│  ┌─────────────▼──────────────────────┐  │
│  │  In-Memory Trace Store (evicting)  │  │
│  └─────────────┬──────────────────────┘  │
│                │                         │
│  ┌─────────────▼──────────────────────┐  │
│  │  Query Engine                      │  │
│  │  Waterfall │ Critical Path │ Deps  │  │
│  └────────────────────────────────────┘  │
│                                          │
│  ┌─────────────────────┐  ┌───────────┐  │
│  │  eBPF Flamegraph    │  │  SLO Burn │  │
│  │  Profiler + Correlat│  │  Rate Eng.│  │
│  └─────────────────────┘  └───────────┘  │
└──────────────────────────────────────────┘
```

---

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/traces` | OTLP/HTTP trace ingestion (standard OTel endpoint) |
| `GET` | `/api/v1/traces` | List recent traces (filterable by service, status, duration) |
| `GET` | `/api/v1/traces/{traceID}` | Get full trace with span waterfall hierarchy |
| `GET` | `/api/v1/traces/{traceID}/critical-path` | Critical path analysis for a trace |
| `GET` | `/api/v1/services` | List all services seen in ingested traces |
| `GET` | `/api/v1/dependencies` | Distributed service dependency map |
| `GET` | `/api/v1/flamegraph/{service}` | Latest eBPF flamegraph for a service (SVG/JSON) |
| `GET` | `/api/v1/flamegraph/{service}/correlated/{traceID}/{spanID}` | Flamegraph slice correlated to a span |
| `GET` | `/api/v1/slo/{service}/burn-rate` | SLO burn rate for a service |
| `POST` | `/api/v1/slo` | Define an SLO for a service |
| `GET` | `/api/v1/slo` | List all SLO definitions |
| `GET` | `/metrics` | Prometheus OpenMetrics text with exemplar links |
| `GET` | `/healthz` | Liveness probe |

---

## SLO Burn Rate Alerting

Define SLOs with dual burn windows:

```bash
curl -X POST http://servtrace:8090/api/v1/slo \
  -d '{
    "service": "orders-api",
    "slo_name": "availability",
    "target_ratio": 0.999,
    "windows": [
      { "name": "fast", "duration": "1h", "burn_rate_threshold": 14.4 },
      { "name": "slow", "duration": "6h", "burn_rate_threshold": 6.0 }
    ]
  }'
```

Query burn rate:
```bash
curl http://servtrace:8090/api/v1/slo/orders-api/burn-rate
# → { "slo": "availability", "budget_remaining": 0.82, "burn_rate_1h": 2.1, "burn_rate_6h": 0.8, "alerting": false }
```

---

## Flamegraph Profiling

eBPF profiling runs continuously in the background. Access profiles via:

```bash
# Get current CPU flamegraph for orders-api
curl http://servtrace:8090/api/v1/flamegraph/orders-api > flamegraph.svg

# Get flamegraph slice correlated to a specific slow span
curl http://servtrace:8090/api/v1/flamegraph/orders-api/correlated/abc123/span456
```

---

## Getting Started

```bash
docker run -p 8090:8090 \
  -e SERVTRACE_MAX_TRACES=50000 \
  -e SERVTRACE_EBPF_ENABLED=true \
  -e SERVTRACE_OTEL_EXPORT=http://collector:4318 \
  ghcr.io/vyuvaraj/servtrace:latest
```

Configure your services to send OTLP traces:

```bash
# Go
OTEL_EXPORTER_OTLP_ENDPOINT=http://servtrace:8090 ./my-service

# Python
opentelemetry-instrument --exporter-otlp-endpoint http://servtrace:8090 python app.py
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVTRACE_PORT` | `8090` | HTTP listener port |
| `SERVTRACE_MAX_TRACES` | `10000` | Max traces in memory before eviction |
| `SERVTRACE_EBPF_ENABLED` | `false` | Enable eBPF continuous profiling |
| `SERVTRACE_OTEL_EXPORT` | — | Re-export spans to another OTLP collector |
