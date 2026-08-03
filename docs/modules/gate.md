# Pranor Gate — API Gateway & Ingress Router

**Version:** 1.0.0  
**Module Path:** `github.com/vyuvaraj/pranor/gate`  
**Default Port:** 8080  
**License:** AGPL-3.0 (OSS) / Enterprise License (EE with eBPF, GraphQL Federation, Geo-IP Steering)

---

## Overview

Pranor Gate is a high-performance API gateway and reverse proxy that routes, secures, and transforms traffic between clients and upstream services. It features WASM-based plugin extensibility, AI-aware traffic management (prompt guard, semantic caching, PII redaction, token billing), weighted canary/blue-green deployments with automatic promotion, circuit breaking, SSE passthrough, WebSocket proxying, and a declarative per-route configuration model.

Pranor Gate can run as:
- A **standalone binary** with local JSON configuration
- An **integrated module** within the Pranor ecosystem with S3-based dynamic config, JWT auth, OTel tracing, and Console visibility
- An **edge proxy** with Let's Encrypt auto-TLS or dynamic certificate fetching from Pranor Secret

---

## Table of Contents

- [Key Features](#key-features)
- [Architecture](#architecture)
- [Installation & Deployment](#installation--deployment)
- [Configuration](#configuration)
- [API Reference](#api-reference)
- [Routing & Traffic Management](#routing--traffic-management)
- [WASM Plugin System](#wasm-plugin-system)
- [AI Guard & LLM Routing](#ai-guard--llm-routing)
- [Security](#security)
- [Observability](#observability)
- [Client Libraries & CLI](#client-libraries--cli)
- [Enterprise Edition](#enterprise-edition)
- [Operational Runbook](#operational-runbook)

---

## Key Features

| Feature | Description |
|---------|-------------|
| **WASM Plugin Middleware** | Upload and hot-register WebAssembly request/response transform modules per route at runtime. |
| **Rate Limiting** | Per-IP, per-route RPM limits with optional Redis-backed distributed enforcement. |
| **Circuit Breaker** | Automatic circuit breaking on upstream failure thresholds with half-open recovery. |
| **AI Prompt Guard** | Inspects and sanitizes inputs for prompt injection attacks on AI/LLM routes. |
| **Semantic Cache** | Embedding-based response cache for AI endpoints — returns cached responses for semantically similar prompts. |
| **PII Redaction** | Automatic detection and masking of personally identifiable information in AI payloads. |
| **Canary / Blue-Green** | Weighted traffic splitting with automated canary promotion and error-rate rollback. |
| **SSE Passthrough** | Transparent proxying of Server-Sent Events streams without buffering. |
| **WebSocket Proxy** | Full-duplex WebSocket proxying with connection tracking. |
| **mTLS to Upstreams** | Per-route mutual TLS client certificates for service-to-service authentication. |
| **Let's Encrypt Auto-TLS** | Zero-config HTTPS with automatic ACME certificate provisioning. |
| **Response Caching** | Configurable per-route TTL response cache for GET requests. |
| **Backpressure Control** | Concurrent request limiting with queue overflow protection. |
| **OpenAPI Validation** | Request payload validation against OpenAPI 3.0 spec per route. |
| **IP Allowlist/Blocklist** | Per-route network ACLs via CIDR ranges. |
| **Structured Access Logs** | JSONL access logging with request/response metadata. |
| **GitOps Config Sync** | Webhook-triggered git pull + config reload for GitOps workflows. |
| **Dynamic Policy Engine** | ServPolicy integration for fine-grained authorization rules compiled to WASM. |
| **AI Token Billing** | Per-route and per-tenant LLM token usage tracking with budget enforcement. |
| **Traffic Replay** | Record traffic to JSONL and replay against WASM middlewares or candidate backends. |

---

## Architecture

```mermaid
graph TD
    classDef ingress fill:#1e293b,stroke:#38bdf8,stroke-width:2px,color:#fff;
    classDef security fill:#0f172a,stroke:#0d9488,stroke-width:2px,color:#fff;
    classDef proxy fill:#1e1b4b,stroke:#6366f1,stroke-width:2px,color:#fff;
    classDef backend fill:#1e293b,stroke:#64748b,stroke-width:1px,color:#fff;

    subgraph Edge ["🌐 Global Ingress Layer"]
        DNS["Geo-IP Anycast DNS"] :::ingress
        XDP["eBPF XDP Packet Filter<br/><i>(100Gbps DDoS Drop)</i>"] :::ingress
    end

    subgraph Security ["🛡️ Zero-Trust Security & WASM Engine"]
        TLS["PCIe Hardware TLS Offload"] :::security
        WASM["WASM Security Sandbox<br/><i>(Side-Channel Safe)</i>"] :::security
        PromptGuard["AI Prompt Injection Guard<br/><i>(Semantic Vector Filter)</i>"] :::security
    end

    subgraph Core ["⚡ Proxy Router & Rate Limiter"]
        CRDT["Global CRDT Rate Limiter<br/><i>(Sub-ms Gossip Sync)</i>"] :::proxy
        Router["Dynamic Reverse Proxy<br/><i>(Canary / Blue-Green)</i>"] :::proxy
    end

    subgraph Upstream ["☁️ Upstream Microservices"]
        AIModel["LLM / Model Service"] :::backend
        Microservice["gRPC / REST Microservice"] :::backend
    end

    DNS --> XDP
    XDP --> TLS
    TLS --> WASM
    WASM --> PromptGuard
    PromptGuard --> CRDT
    CRDT --> Router
    Router -->|mTLS Stream| AIModel
    Router -->|mTLS Stream| Microservice
```

### Request Processing Sequence & WASM Execution Flow

```mermaid
sequenceDiagram
    autonumber
    participant Client as Client Application
    participant Gate as Pranor Gate Ingress
    participant Auth as Pranor Auth / JWT Validator
    participant WASM as WASM Plugin Sandbox
    participant AI as AI Prompt Guard
    participant Service as Upstream Microservice

    Client->>Gate: HTTP Request / POST /v1/ai/prompt
    Gate->>Auth: Validate JWT / SPIFFE SVID Token
    Auth-->>Gate: Token Validated (Claims + Tenant Context)
    Gate->>WASM: Execute Request Transformer (WASM)
    WASM-->>Gate: Transformed Headers & Body
    Gate->>AI: Evaluate Prompt Injection & PII Redaction
    AI-->>Gate: Sanitized Prompt & Risk Score (Passed)
    Gate->>Service: Forward Request (mTLS + Retrying Transport)
    Service-->>Gate: Response Stream (200 OK)
    Gate->>WASM: Execute Response Transformer (WASM)
    WASM-->>Gate: Final Formatted Payload
    Gate-->>Client: Streamed HTTP Response + X-Token-Cost
```

### Ecosystem Cross-Module Integration

Pranor Gate acts as the front door for the entire Pranor platform, seamlessly interfacing with core infrastructure services:

- **Pranor Auth**: Automatically verifies incoming JWT signatures, SAML claims, and SPIFFE/SPIRE x509 workload identities (`PRANOR_JWT_SECRET`).
- **Pranor Secret**: Dynamically fetches and auto-rotates TLS server certificates and client mTLS credentials without restarting the proxy.
- **Pranor Trace**: Generates W3C-compliant `traceparent` OpenTelemetry headers, emitting distributed trace spans for every proxied request.
- **Pranor Console**: Streams real-time throughput, p99 latency histograms, and active WASM plugin health metrics directly to the control plane dashboard.
- **Pranor Vault**: Pulls dynamic S3-backed JSON route configurations and uploads recorded JSONL traffic replay logs.

---

## Installation & Deployment

### Binary

```bash
cd pranor/gate
go build -o pranor-gate .
./pranor-gate
```

### Docker

```bash
docker run -p 8080:8080 ghcr.io/vyuvaraj/pranor-gate:latest
```

### Docker Compose

```yaml
services:
  gate:
    image: ghcr.io/vyuvaraj/pranor-gate:latest
    ports:
      - "8080:8080"
    volumes:
      - ./config.json:/app/config.json
    environment:
      - PRANOR_JWT_SECRET=your-secret
```

### As Part of Pranor Ecosystem

When running under the Pranor platform, Gate integrates automatically with Auth (JWT/mTLS), Secret (dynamic certificates), Trace (OTel spans), and Console (dashboard visibility). Configuration can be pulled from an S3-compatible store for centralized management.

---

## Configuration

### JSON Config (`config.json`)

```json
{
  "addr": ":8080",
  "auth_token": "gateway-secret-token",
  "tls_cert": "",
  "tls_key": "",
  "routes": [
    {
      "prefix": "/api/v1/services",
      "target": "http://127.0.0.1:8081",
      "middleware": "uppercase",
      "rate_limit_rpm": 120,
      "cache_ttl_seconds": 60,
      "access_log": true
    },
    {
      "prefix": "/ai/v1",
      "target": "http://127.0.0.1:11434",
      "enable_semantic_cache": true,
      "enable_prompt_guard": true,
      "semantic_token_limit_per_min": 10000
    }
  ]
}
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PRANOR_JWT_SECRET` | — | JWT signing key for token-based auth |
| `PRANOR_AUTO_TLS` | `false` | Enable Let's Encrypt auto-TLS |
| `PRANOR_AUTO_TLS_DOMAIN` | — | Domain for ACME certificate |
| `PRANOR_CONFIG_S3_BUCKET` | — | S3 bucket for remote config |
| `PRANOR_DISCOVERY` | — | Service discovery endpoint |
| `PRANOR_SECRET_URL` | — | Pranor Secret service URL for dynamic certs |
| `PRANOR_SECRET_API_KEY` | — | API key for Pranor Secret |
| `PRANOR_SECRET_TENANT_ID` | `default` | Tenant ID for secret lookup |
| `PRANOR_GATE_LIMITS_REDIS_URL` | — | Redis URL for distributed rate limiting |
| `PRANOR_REGISTRY` | `https://registry.pranor.org` | WASM middleware registry URL |
| `PRANOR_CLUSTER` | `default` | Cluster identifier for tenant policies |
| `PRANOR_REGION` | `us-east` | Region identifier for tenant policies |
| `PRANOR_OTLP_ENDPOINT` | — | OpenTelemetry collector URL |

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `config.json` | Path to configuration file |

### CLI Subcommands

| Command | Description |
|---------|-------------|
| `pranor-gate` | Start the gateway server |
| `pranor-gate dashboard` | Launch terminal TUI traffic dashboard |
| `pranor-gate replay --log FILE --middleware FILE.wasm` | Replay recorded traffic through WASM |
| `pranor-gate replay --shadow --log FILE --target URL` | Shadow diff replay against candidate backend |
| `pranor-gate install <name>` | Install WASM middleware from registry |
| `pranor-gate policy compile <file.policy> -o <file.wasm>` | Compile policy DSL to WASM |

---

## API Reference

**Base URL:** `http://localhost:8080`  
**API Version:** `/api/v1/` (recommended) or `/api/` (legacy)

### GET /healthz

Liveness probe.

```json
{"status":"UP","service":"pranor","version":"1.0.0"}
```

### GET /readyz

Readiness probe. Same format as healthz.

---

### GET /api/v1/routes

List all configured routes.

**Response (200):**

```json
[
  {
    "prefix": "/api/v1/services",
    "target": "http://127.0.0.1:8081",
    "middleware": "uppercase",
    "rate_limit_rpm": 120
  }
]
```

### POST /api/v1/routes

Register or update a route dynamically.

**Request:**

```json
{
  "prefix": "/api/v2/users",
  "target": "http://users-service:8080",
  "rate_limit_rpm": 200,
  "cache_ttl_seconds": 30,
  "ip_allowlist": ["10.0.0.0/8"]
}
```

**Response (200):**

```
Route registered successfully
```

### DELETE /api/v1/routes?prefix=/api/v2/users

Remove a route.

**Response (200):**

```
Route deleted successfully
```

---

### POST /api/v1/admin/middleware/{name}

Register a WASM middleware plugin at runtime.

**Request:** Raw `.wasm` binary as request body.

**Response (200):**

```
WASM Middleware auth-check compiled and registered
```

---

### GET /api/v1/admin/connections

List active backend connections.

**Response (200):**

```json
{
  "http://127.0.0.1:8081": 5,
  "http://127.0.0.1:8082": 2
}
```

---

### DELETE /api/v1/admin/cache?prefix=/api/v1/data

Invalidate response cache entries.

**Response (200):**

```json
{
  "status": "success",
  "entries_invalidated": 12,
  "prefix": "/api/v1/data"
}
```

---

### POST /api/v1/admin/policy/reload

Hot-reload the dynamic IAM policy schema.

**Request (optional body):** Policy schema JSON.

**Response (200):**

```json
{"status": "success", "message": "Policy schema updated"}
```

---

### POST /api/v1/admin/policy/revoke

Revoke all sessions for a user.

**Request:**

```json
{"username": "compromised-user"}
```

**Response (200):**

```json
{"status": "success", "message": "Session revoked for user compromised-user"}
```

---

### GET /api/v1/admin/ai-billing

Retrieve AI token usage and cost metrics.

**Response (200):**

```json
{
  "total_tokens": 1523400,
  "total_cost_usd": 4.57,
  "per_tenant": {
    "tenant-a": {"tokens": 800000, "cost_usd": 2.40}
  }
}
```

### POST /api/v1/admin/ai-billing

Set per-tenant AI budget limits.

**Request:**

```json
{
  "tenant_id": "tenant-a",
  "max_cost_per_day_usd": 10.00,
  "max_tokens_per_minute": 50000
}
```

---

### GET /api/v1/admin/ai-cost-attribution

Per-route AI token and cost attribution dashboard.

**Response (200):**

```json
{
  "routes": [
    {
      "prefix": "/ai/v1",
      "total_tokens": 500000,
      "total_cost_usd": 1.50,
      "estimated_savings": 0.30
    }
  ],
  "summary": {
    "total_cost_usd": 4.57,
    "total_tokens": 1523400,
    "estimated_savings": 0.91,
    "savings_percent": 16.6
  }
}
```

---

### GET /api/v1/admin/metrics/ws

WebSocket endpoint streaming real-time gateway metrics (RPS, error rate, active connections).

---

### POST /api/v1/admin/console/sync

Synchronize full route configuration from Pranor Console.

**Request:**

```json
{"routes": [...]}
```

### GET /api/v1/admin/console/sync

Get current gateway state snapshot (routes, connections, metrics).

---

### POST /api/v1/gitops/webhook

Trigger a git pull + config reload for GitOps-managed configuration.

**Response (200):**

```json
{
  "status": "success",
  "message": "GitOps config sync completed successfully",
  "git_output": "Already up to date."
}
```

---

### POST /api/v1/routes/register

Register a route via the compiler connector (for Pranor Lang integration).

---

### GET /api/docs

Embedded interactive API documentation page.

### GET /api/docs/openapi.json

Auto-generated OpenAPI specification from current routes.

---

## Routing & Traffic Management

### Prefix-Based Matching

Routes are matched by longest-prefix on the request URL path. The first matching route wins.

### Weighted Canary / Blue-Green Deployments

Distribute traffic between stable and canary targets by weight:

```json
{
  "prefix": "/api/v1/orders",
  "targets_weighted": [
    {"url": "http://orders-v1:8080", "weight": 90},
    {"url": "http://orders-v2:8080", "weight": 10}
  ],
  "canary_auto_promote": true,
  "canary_promote_step": 10,
  "canary_promote_sec": 60,
  "canary_max_error_rate": 0.01
}
```

The canary engine automatically:
1. Increments canary weight by `canary_promote_step` every `canary_promote_sec` seconds
2. Monitors error rate on the canary target
3. Rolls back to 100% stable if error rate exceeds `canary_max_error_rate`
4. Disables auto-promotion once canary reaches 100%

### Load Balancing

Multiple targets support round-robin and least-connections strategies:

```json
{
  "prefix": "/api/v1/users",
  "targets": ["http://users-1:8080", "http://users-2:8080", "http://users-3:8080"],
  "load_balancer": "least_conn"
}
```

### Circuit Breaker

Automatically opens when upstream error rate exceeds threshold, preventing cascade failures. Half-open state probes recovery.

### Backpressure Control

Per-route concurrency limiting with queue overflow:

```json
{
  "max_concurrent_requests": 100,
  "max_queue_size": 500,
  "queue_timeout_ms": 5000
}
```

Returns `503 Service Unavailable` when queue is full, `504 Gateway Timeout` on queue timeout.

### Response Caching

```json
{
  "cache_ttl_seconds": 60,
  "cache_methods": ["GET"]
}
```

---

## WASM Plugin System

### Architecture

WASM middlewares are compiled via wazero (pure-Go WebAssembly runtime, no CGO). Plugins receive the request, can transform headers/body, and return modified content.

### Registering a Plugin

```bash
# From registry
pranor-gate install jwt-auth

# Upload directly
curl -X POST http://localhost:8080/api/v1/admin/middleware/my-filter \
  -H "Authorization: Bearer gateway-secret-token" \
  --data-binary @my-filter.wasm
```

### Per-Route Assignment

```json
{
  "prefix": "/api/v1/data",
  "middleware": "my-filter",
  "response_middleware": "response-transform"
}
```

### WASM A/B Testing

Split traffic between different WASM middleware versions:

```json
{
  "wasm_split": {
    "targets": [
      {"middleware_name": "filter-v1", "weight": 80},
      {"middleware_name": "filter-v2", "weight": 20}
    ]
  }
}
```

### Policy DSL Compilation

Write human-readable policies and compile to WASM:

```
# auth.policy
allow GET /api/public/*
deny POST /api/admin/* if header.role == "viewer"
allow * * if header.x-internal == "true"
```

```bash
pranor-gate policy compile auth.policy -o auth.wasm
```

---

## AI Guard & LLM Routing

### Prompt Guard

Detects and blocks prompt injection attempts on AI-routed traffic:

```json
{"prefix": "/ai/v1", "prompt_guard": true}
```

### PII Redaction

Masks sensitive data (emails, SSN, credit cards) before forwarding to LLM backends:

```json
{"prefix": "/ai/v1", "pii_redact": true}
```

### Semantic Cache

Caches LLM responses and returns cached versions for semantically similar prompts (cosine similarity > 0.85):

```json
{"prefix": "/ai/v1", "semantic_cache": true}
```

### LLM Routing with Fallback

Route to a primary model with automatic fallback on low confidence:

```json
{
  "llm_routing": {
    "primary": {"url": "http://ollama:11434", "model": "llama3"},
    "fallback": {"url": "https://api.openai.com", "model": "gpt-4"},
    "confidence_header": "X-Confidence",
    "min_confidence": 0.7
  }
}
```

### Semantic Rate Limiting

Token-based rate limiting for LLM routes (tokens-per-minute rather than requests-per-minute):

```json
{"semantic_rate_limit": true, "semantic_token_limit_per_min": 10000}
```

### Prompt A/B Testing

Route different prompt templates to measure response quality.

---

## Security

### Bearer Token Auth

Set `auth_token` in config. All non-health endpoints require:

```
Authorization: Bearer gateway-secret-token
```

### JWT Authentication

When `PRANOR_JWT_SECRET` is set, validates JWT Bearer tokens. Supports policy versioning — stale tokens get `X-Token-Refresh: true` header.

### Dynamic Secret Fetching

Gate can fetch TLS certificates and JWT secrets dynamically from Pranor Secret at startup.

### mTLS to Upstreams

Per-route client certificate for backend authentication:

```json
{
  "client_cert_path": "/certs/client.crt",
  "client_key_path": "/certs/client.key",
  "root_ca_path": "/certs/backend-ca.crt"
}
```

### Multi-Tenant API Keys

Per-key rate limits, route restrictions, and tenant isolation:

```json
{"require_api_key": true, "allowed_tenants": ["tenant-a", "tenant-b"]}
```

### IP Allowlist / Blocklist

```json
{
  "ip_allowlist": ["10.0.0.0/8", "192.168.1.0/24"],
  "ip_blocklist": ["1.2.3.4"]
}
```

### Request Body Size Limits

Per-route body size enforcement (default 10MB):

```json
{"max_body_size": 5242880}
```

### Session Revocation

Instant session revocation via admin API without waiting for token expiry.

### Dynamic IAM Policy (ServPolicy)

Upload OPA-style policy schemas that Gate evaluates inline per request.

---

## Observability

### Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `total_requests` | Counter | Total proxied requests |
| `total_errors` | Counter | Total upstream errors |
| `request_rate` | Gauge | Requests/second (1s window) |
| `error_rate` | Gauge | Errors/second (1s window) |
| `active_connections` | Gauge | Per-target active connections |

### WebSocket Live Metrics

Connect to `/api/v1/admin/metrics/ws` for 1-second streaming metrics updates.

### Access Logging

Structured JSONL access logs per route:

```json
{
  "timestamp": "2026-01-15T10:00:00Z",
  "method": "GET",
  "path": "/api/v1/users/123",
  "status": 200,
  "latency_ms": 42,
  "client_ip": "10.0.1.5",
  "upstream": "http://users:8080"
}
```

### OpenTelemetry Tracing

Every proxied request gets an OTel span with method, route, status code, and upstream latency.

### Terminal Dashboard

```bash
pranor-gate dashboard
```

Live TUI showing real-time RPS, P99 latency, circuit breaker state, cache hit rate.

---

## Client Libraries & CLI

### cURL

```bash
# Register a route
curl -X POST http://localhost:8080/api/v1/routes \
  -H "Authorization: Bearer gateway-secret-token" \
  -H "Content-Type: application/json" \
  -d '{"prefix":"/api/v2/users","target":"http://users:8080","rate_limit_rpm":100}'

# Upload WASM middleware
curl -X POST http://localhost:8080/api/v1/admin/middleware/auth-check \
  -H "Authorization: Bearer gateway-secret-token" \
  --data-binary @auth-check.wasm

# Invalidate cache
curl -X DELETE "http://localhost:8080/api/v1/admin/cache?prefix=/api/v1/data" \
  -H "Authorization: Bearer gateway-secret-token"
```

### Pranor CLI

```bash
pranor gate routes list
pranor gate routes add --prefix /api/v2 --target http://backend:8080 --rate-limit 100
pranor gate middleware install jwt-auth
pranor gate dashboard
pranor gate replay --log traffic.jsonl --middleware filter.wasm
```

---

## Enterprise Edition

| Feature | OSS | EE |
|---------|:---:|:--:|
| WASM plugin middleware | ✓ | ✓ |
| Rate limiting (local) | ✓ | ✓ |
| Rate limiting (Redis distributed) | ✓ | ✓ |
| Circuit breaker | ✓ | ✓ |
| Canary / Blue-Green deployments | ✓ | ✓ |
| AI Prompt Guard & PII Redaction | ✓ | ✓ |
| Semantic cache | ✓ | ✓ |
| SSE passthrough & WebSocket proxy | ✓ | ✓ |
| mTLS to upstreams | ✓ | ✓ |
| Let's Encrypt Auto-TLS | ✓ | ✓ |
| GitOps config sync | ✓ | ✓ |
| Traffic replay engine | ✓ | ✓ |
| AI token billing & budgets | ✓ | ✓ |
| Kernel eBPF XDP DDoS bypass (100Gbps) | — | ✓ |
| Geo-IP latency anycast steering | — | ✓ |
| GraphQL schema stitching & federation | — | ✓ |
| SSL offloading (hardware acceleration) | — | ✓ |
| AI self-defending WAF | — | ✓ |
| Multi-cluster enterprise control plane | — | ✓ |

---

## Operational Runbook

### Route not matching / 502 Bad Gateway

1. Check `/api/v1/routes` for the configured routes
2. Verify the request path has the correct prefix
3. Ensure the upstream target is reachable from the gateway
4. Check circuit breaker state via metrics

### High latency on specific route

1. Check `/api/v1/admin/connections` for connection count
2. Review backpressure settings — `max_concurrent_requests` may be too low
3. Check if circuit breaker is in half-open state (probing slowly)
4. Look at upstream health via WebSocket metrics stream

### Rate limiting kicking in unexpectedly

1. Verify `rate_limit_rpm` is set correctly on the route
2. Check if Redis-based distributed limiting is configured — all instances share state
3. Per-API-key limits may be more restrictive than route limits
4. Review semantic token rate limits for AI routes

### WASM middleware failing

1. Check gateway logs for WASM compilation errors
2. Use `pranor-gate replay --log traffic.jsonl --middleware broken.wasm` to test offline
3. Verify WASM module exports the correct ABI functions
4. Check if the middleware registry URL is reachable

### Canary deployment not promoting

1. Check error rate on canary target — exceeding `canary_max_error_rate` causes rollback
2. Verify `canary_auto_promote` is `true`
3. Ensure at least 3 requests have hit the canary (minimum sample for error rate calculation)
4. Check `canary_promote_sec` interval — promotion may not have triggered yet

### TLS certificate issues

1. If using auto-TLS, ensure port 80 is accessible for HTTP challenge
2. For Pranor Secret integration, verify `PRANOR_SECRET_URL` connectivity
3. Check certificate paths in config for file-based TLS
4. Review gateway startup logs for certificate loading errors

---

## Versioning & Compatibility

- API is versioned at `/api/v1/`
- Legacy `/api/` paths continue to work (internally mapped to v1)
- Configuration format is backward-compatible across minor versions
- WASM ABI is stable — plugins compiled for v1.0 work on all v1.x releases
