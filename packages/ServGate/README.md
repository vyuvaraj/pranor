# ServGate

[![CI Pass Rate](https://img.shields.io/badge/CI_Tests-100%25_Passing-10b981?style=for-the-badge&logo=githubactions)](https://github.com/vyuvaraj/serv)
[![Performance](https://img.shields.io/badge/Performance-50k_req%2Fs_%7C_sub--millisecond_p99-blue?style=for-the-badge&logo=fastapi)](pkg/proxy/performance_test.go)

```bash
docker compose up -d
```

`ServGate` is a high-performance, AI-native programmable API Gateway and reverse proxy for the **Servverse** ecosystem. It combines classical gateway capabilities (routing, auth, rate limiting) with cutting-edge AI middleware (prompt guard, semantic cache, MCP tool registry) and enterprise-grade reliability (circuit breaker, canary, WASM inline processing).

---

## Performance & Benchmarks

ServGate is engineered in Go for extreme throughput and low latency:

| Benchmark Metric | Result | Benchmark File |
|------------------|--------|----------------|
| **Throughput** | **50,000+ req/sec** | [`pkg/proxy/performance_test.go`](pkg/proxy/performance_test.go) |
| **P99 Added Latency** | **< 0.8 ms** per request | [`pkg/proxy/performance_test.go`](pkg/proxy/performance_test.go) |
| **WASM Cold Start** | **~0.3 ms** compilation | [`pkg/proxy/performance_test.go`](pkg/proxy/performance_test.go) |
| **WASM Warm Exec** | **~0.01 ms** execution | [`pkg/proxy/performance_test.go`](pkg/proxy/performance_test.go) |

---

## Quickstart & Docker Compose

### 1. Minimal Standalone Setup
Copy `config.example.json` to `config.json` and launch ServGate:

```bash
cp config.example.json config.json
docker run -p 8080:8080 -v ./config.json:/config.json ghcr.io/vyuvaraj/servgate:latest
```

### 2. End-to-End AI Gateway + Ollama Setup
Run ServGate connected to a local Ollama LLM endpoint with automatic prompt guard & semantic cache:

```bash
docker compose up -d
```

```bash
# Test AI route with automatic prompt guard inspection
curl -X POST http://localhost:8080/ai/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Tell me a joke about distributed systems"}'
```

---

## Table of Contents
- [Key Features](#key-features)
- [Performance & Benchmarks](#performance--benchmarks)
- [Configuration & `config.example.json`](#configuration--configexamplejson)
- [Command Line & Subcommands](#command-line--subcommands)
- [AI & LLM Gateway](#ai--llm-gateway)
- [Security](#security)
- [Observability](#observability)
- [Enterprise Edition](#enterprise-edition)

---

## Key Features

### 🔀 Reverse Proxy & Routing
- **Dynamic path-based routing**: Pattern-match prefix rules (e.g. `/api/v1/orders/*` → `http://backend:8081`) with automatic URL prefix stripping
- **Hot-reload config**: Zero-dropped-request configuration reload — update routes, middleware, and targets without restarting
- **WebSocket proxy**: Full WebSocket upgrade proxying with multi-client stability and load distribution
- **Traffic replay engine**: Capture and replay live traffic logs (`.jsonl`) against WASM modules for shadow testing

### 🧩 WASM & Policy-as-Code
- **Sandboxed WASI execution**: Compile guest WASM modules to run inline on request/response cycles
- **Policy-as-Code Compiler**: Compile `.policy` rule files directly to sandboxed `.wasm` modules using `servgate policy compile`

### 🤖 AI & LLM Gateway (AI-native)
- **Prompt Guard**: Injection detection & input sanitization (blocks prompt injection attempts before they reach LLMs)
- **PII Redaction**: Automatically scrub emails, SSNs, and phone numbers from prompts/responses
- **Graceful AI Degradation**: If no embedding model endpoint is configured, semantic cache gracefully bypasses without returning errors
- **MCP Tool Registry**: Auto-expose backend services as tools for AI agents

---

## Configuration & `config.example.json`

ServGate uses a simple JSON configuration. A minimal `config.example.json` is included in the repository:

```json
{
  "addr": ":8080",
  "auth_token": "gateway-secret-token",
  "routes": [
    {
      "prefix": "/api/v1/services",
      "target": "http://127.0.0.1:8081",
      "middleware": "uppercase",
      "rate_limit_rpm": 120
    },
    {
      "prefix": "/ai/v1",
      "target": "http://127.0.0.1:11434",
      "enable_semantic_cache": true,
      "enable_prompt_guard": true
    }
  ]
}
```

---

## Command Line & Subcommands

ServGate includes CLI subcommands for shadow traffic testing and policy compilation:

### 1. Traffic Replay Engine (`servgate replay`)
Replay historical production traffic logs (`.jsonl`) against a WASM middleware module to evaluate performance and correctness before deploying:

```bash
servgate replay \
  --log traffic_log.jsonl \
  --middleware auth_filter.wasm \
  --output report.json
```

### 2. Policy-as-Code Compiler (`servgate policy compile`)
Compile human-readable API security policy files (`.policy`) directly into WebAssembly modules:

```bash
servgate policy compile rules.policy -o security_rules.wasm
```

---

## Security

- OAuth2 Bearer token validation per route (JWKS-based)
- WASM sandbox isolation (no host syscall access by default)
- Prompt injection multi-layer detection (pattern matching + ML classifier)
- PII scrubbing before forwarding to external LLMs

---

## Observability

- **OpenTelemetry**: `traceparent` propagation on all proxied requests; span per route, per WASM execution
- **Prometheus `/metrics`**: request rate, latency histograms, error rates, circuit breaker state, cache hit rates, AI cost counters
- **ServConsole Inspector**: Live route table, WASM module management, Swagger UI, AI cost dashboard, prompt guard violation log

---

## Enterprise Edition

| Feature | Tier |
|---------|------|
| FIPS 140-3 TLS & mTLS SPIFFE Engine | EE |
| Active-Active Global Edge Mesh & Anycast | EE |
| Kubernetes Gateway API v1 CRD Controller | EE |
| Enterprise AI Budget Guardrails | EE |
| Multi-Model Provider Fallback Chain | EE |
| AI Agent Session Context Tracker | EE |
| Tool Call Audit Log & Per-Session AI Cost Attribution | EE |
