# ServGate

```bash
docker run -p 8080:8080 ghcr.io/vyuvaraj/servgate:latest
```

`ServGate` is a high-performance, AI-native programmable API Gateway and reverse proxy for the **Servverse** ecosystem. It combines classical gateway capabilities (routing, auth, rate limiting) with cutting-edge AI middleware (prompt guard, semantic cache, MCP tool registry) and enterprise-grade reliability (circuit breaker, canary, WASM inline processing).

---

## Table of Contents
- [Key Features](#key-features)
- [Architecture](#architecture)
- [API Endpoints](#api-endpoints)
- [Configuration](#configuration)
- [Getting Started](#getting-started)
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
- **Traffic replay engine**: Capture and replay live traffic for shadow testing and debugging

### 🔒 Authentication & Authorization
- **OAuth2 Bearer token validation**: Built-in JWT/Bearer enforcement on routed backend targets
- **Zero-Trust integration**: Works natively with `ServAuth` for passkey, MFA, and RBAC enforcement

### ⚡ Rate Limiting
- **Token bucket rate limiter**: Per-route, per-client configurable burst and sustained rates
- **Sliding window counters**: Precise per-second/per-minute rate windows
- **Distributed rate limiting**: Cross-node coordination via `ServCache` token buckets

### 🧩 WASM Inline Middleware
- **Sandboxed WASI execution**: Compile guest WASM modules to run inline on request/response cycles
- **Hot-swap**: Upload and activate new WASM modules at runtime without restart
- **Use cases**: Header validation, payload mutation, query param enrichment, PII redaction, request signing

### 🚦 Traffic Management
- **Canary deployment routing**: Split traffic by percentage to canary backends; automatic rollback on error-rate breach
- **Circuit breaker**: SLO-based breach detection with automatic open/half-open/closed state transitions
- **Blue/Green support**: Route all traffic atomically between active/new deployment environments

### 🤖 AI & LLM Gateway (AI-native)
- **Prompt Guard**: Injection detection & input sanitization (blocks prompt injection attempts before they reach LLMs)
- **PII Redaction**: Automatically scrub personally identifiable information from prompts/responses
- **Semantic Cache**: Similarity-based response caching — return cached LLM responses for semantically equivalent prompts (configurable cosine threshold)
- **A/B Prompt Testing**: Route prompt variants to different model endpoints, measure quality metrics
- **AI Cost Estimation**: Per-request token cost attribution and estimation before forwarding
- **Cost-Optimization LLM Router**: Intelligently route to cheapest capable model (GPT-4o → Claude → Ollama) based on prompt complexity classification
- **Speculative Prompt Pre-fetching**: Pre-warm LLM inference for high-probability follow-up prompts
- **Real-time AI Bill Savings Telemetry**: Track and report per-route cost savings from cache hits and smart routing

### 🛠️ MCP & Agent Integration
- **Native MCP Tool Registry**: Auto-expose all Servverse services (ServStore, ServQueue, ServFlow, etc.) as MCP-compatible AI agent tools — zero configuration
- **LLM Streaming SSE Passthrough**: Zero-latency server-sent event (SSE) streaming passthrough for token-by-token LLM responses
- **Prompt Injection Detection Guard**: Multi-layer sanitization pipeline for agentic workloads

---

## Architecture

```
Client Request
     │
     ▼
┌─────────────────────────────────────────────┐
│               ServGate                       │
│                                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │ Rate     │  │ Auth     │  │ WASM     │  │
│  │ Limiter  │→ │ Guard    │→ │ Filter   │  │
│  └──────────┘  └──────────┘  └──────────┘  │
│                      │                      │
│  ┌───────────────────▼────────────────────┐ │
│  │         AI Middleware Pipeline         │ │
│  │  PII Redact → Prompt Guard → Sem Cache │ │
│  │  → Cost Estimate → Model Router        │ │
│  └────────────────────────────────────────┘ │
│                      │                      │
│  ┌───────────────────▼────────────────────┐ │
│  │         Circuit Breaker / Router       │ │
│  │   Canary Split │ Blue/Green │ Replay   │ │
│  └────────────────────────────────────────┘ │
└─────────────────────────────────────────────┘
     │
     ▼
Backend Services / LLM Providers / Servverse
```

---

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/routes` | Add a new proxy route |
| `GET` | `/api/v1/routes` | List all active routes |
| `DELETE` | `/api/v1/routes/{id}` | Remove a route |
| `POST` | `/api/v1/wasm/upload` | Upload and activate a WASM middleware module |
| `GET` | `/api/v1/wasm/modules` | List loaded WASM modules |
| `POST` | `/api/v1/ratelimit/config` | Configure rate limit policy for a route |
| `GET` | `/api/v1/circuit-breaker/status` | Current circuit breaker state per route |
| `POST` | `/api/v1/canary/config` | Configure canary traffic split |
| `POST` | `/api/v1/replay/capture` | Start request capture session |
| `POST` | `/api/v1/replay/replay` | Replay captured traffic |
| `GET` | `/api/v1/ai/cache/stats` | Semantic cache hit/miss statistics |
| `POST` | `/api/v1/ai/prompt-guard/config` | Configure prompt injection guard rules |
| `GET` | `/api/v1/ai/cost/report` | Per-route AI cost attribution report |
| `GET` | `/api/v1/mcp/tools` | List all registered MCP tools |
| `POST` | `/api/v1/config/reload` | Trigger hot config reload |
| `/metrics` | `GET` | Prometheus metrics |
| `/healthz` | `GET` | Liveness probe |
| `/readyz` | `GET` | Readiness probe |

---

## Configuration

```json
{
  "routes": [
    {
      "prefix": "/api/v1/orders",
      "target": "http://orders-service:8081",
      "auth": { "type": "bearer", "jwks_url": "https://auth.servverse.net/.well-known/jwks" },
      "rate_limit": { "requests_per_second": 100, "burst": 200 },
      "wasm_module": "payload-validator.wasm",
      "canary": { "weight": 10, "target": "http://orders-v2:8082", "rollback_error_rate": 0.05 }
    }
  ],
  "ai": {
    "semantic_cache": { "enabled": true, "similarity_threshold": 0.92 },
    "prompt_guard": { "enabled": true, "block_on_injection": true },
    "pii_redaction": { "enabled": true },
    "model_router": {
      "providers": [
        { "name": "gpt-4o", "endpoint": "https://api.openai.com/v1", "cost_per_1k_tokens": 0.03 },
        { "name": "claude-3", "endpoint": "https://api.anthropic.com/v1", "cost_per_1k_tokens": 0.015 },
        { "name": "ollama", "endpoint": "http://ollama:11434", "cost_per_1k_tokens": 0 }
      ]
    }
  },
  "circuit_breaker": { "failure_threshold": 5, "timeout_ms": 30000 },
  "otel": { "endpoint": "http://servtrace:4318" }
}
```

---

## Getting Started

```bash
# Run ServGate with default config
docker run -p 8080:8080 \
  -e SERVGATE_CONFIG=/etc/servgate/config.json \
  -v ./config.json:/etc/servgate/config.json \
  ghcr.io/vyuvaraj/servgate:latest

# Add a route
curl -X POST http://localhost:8080/api/v1/routes \
  -H "Content-Type: application/json" \
  -d '{"prefix": "/api/orders", "target": "http://myservice:3000"}'

# Upload a WASM middleware
curl -X POST http://localhost:8080/api/v1/wasm/upload \
  -F "module=@my-filter.wasm" \
  -F "route=/api/orders"
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVGATE_PORT` | `8080` | Listener port |
| `SERVGATE_CONFIG` | `config.json` | Config file path |
| `SERVGATE_OTEL_ENDPOINT` | — | OpenTelemetry collector URL |
| `SERVGATE_SERVCACHE_URL` | — | ServCache URL for distributed rate limiting |
| `SERVGATE_SERVAUTH_URL` | — | ServAuth URL for JWT validation |

---

## AI & LLM Gateway

ServGate provides a complete AI gateway layer for LLM-backed services:

```
POST /api/v1/chat → ServGate AI Pipeline:
  1. PII Redaction (scrub emails, SSNs, phone numbers)
  2. Prompt Injection Guard (block adversarial prompts)
  3. Semantic Cache lookup (return cached if similarity >= threshold)
  4. Complexity Classifier → Model Router (pick cheapest capable model)
  5. Forward to LLM (with SSE streaming passthrough)
  6. Cost tracking + telemetry
```

**MCP Tool Registry** — any Servverse service automatically becomes an AI agent tool:

```json
GET /api/v1/mcp/tools
{
  "tools": [
    { "name": "servstore.get_object", "description": "Retrieve an object from ServStore", ... },
    { "name": "servqueue.publish", "description": "Publish a message to a ServQueue topic", ... },
    { "name": "servflow.trigger_workflow", "description": "Trigger a ServFlow workflow", ... }
  ]
}
```

---

## Security

- OAuth2 Bearer token validation per route (JWKS-based)
- WASM sandbox isolation (no host syscall access by default)
- Prompt injection multi-layer detection (pattern matching + ML classifier)
- PII scrubbing before forwarding to external LLMs
- Audit log for every AI tool call (cost + session attribution)

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
