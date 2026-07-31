# ServMesh

```bash
docker run -p 8095:8095 ghcr.io/vyuvaraj/servmesh:latest
```

`ServMesh` is the intelligent service mesh for the **Servverse** ecosystem, providing latency-aware load balancing, distributed rate limiting, live topology telemetry, and chaos fault injection — all without requiring sidecar proxies.

---

## Table of Contents
- [Key Features](#key-features)
- [Architecture](#architecture)
- [API Endpoints](#api-endpoints)
- [Load Balancing](#load-balancing)
- [Rate Limiting](#rate-limiting)
- [Chaos Fault Injection](#chaos-fault-injection)
- [Getting Started](#getting-started)

---

## Key Features

### ⚖️ Load Balancing
- **Latency-aware Power-of-Two-Choices (P2C)**: On each routing decision, sample two random backends and pick the one with lower observed latency — dramatically reduces tail latency compared to round-robin
- **Locality preference**: Prefer backends in the same availability zone/region before spilling over to remote nodes; configurable locality weight
- **Health-aware routing**: Unhealthy backends are automatically excluded; exponential recovery probing

### 🚦 Distributed Rate Limiting
- **Global rate limiting via ServCache token buckets**: Rate limit counters stored in ServCache — all mesh nodes share state for true global enforcement (not per-node)
- **Per-service and per-route policies**: Define separate rate limits per service, per endpoint pattern
- **Burst control**: Token bucket allows short bursts above sustained rate

### 🗺️ Live Topology Telemetry
- **Real-time service topology graph**: ServMesh tracks all observed service-to-service call edges and pushes live updates to ServConsole via WebSocket
- **Traffic flow visualization**: Annotates edges with RPS, error rate, and p99 latency in real-time
- **Dependency discovery**: Automatically discovers service dependencies without manual configuration

### 💥 Chaos Fault Injection
- **Latency injection**: Add artificial delay (configurable distribution: fixed, uniform, normal) to selected service calls
- **Error rate simulation**: Inject synthetic HTTP errors (configurable status code and percentage)
- **Network partition simulation**: Block traffic between specified service pairs
- **Abort experiments**: Immediately restore normal traffic flow; auto-expiry on configured duration
- **Blast radius preview**: Preview which service pairs are affected before triggering

---

## Architecture

```
Service A ──→ ServMesh Router ──→ Service B (selected by P2C)
                    │
                    ├── ServCache (distributed rate limit counters)
                    ├── Chaos Engine (inject faults)
                    └── Topology Emitter (→ ServConsole WebSocket)
```

---

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/services` | Register a service endpoint |
| `GET` | `/api/v1/services` | List all registered services |
| `POST` | `/api/v1/route` | Route a request (P2C selection) |
| `GET` | `/api/v1/topology` | Current topology graph snapshot |
| `POST` | `/api/v1/ratelimit/policy` | Set rate limit policy for a service |
| `GET` | `/api/v1/ratelimit/policy` | List rate limit policies |
| `POST` | `/api/v1/chaos/inject` | Inject a chaos fault |
| `POST` | `/api/v1/chaos/abort/{id}` | Abort an active chaos fault |
| `GET` | `/api/v1/chaos/active` | List active chaos faults |
| `/metrics` | `GET` | Prometheus metrics (routing decisions, rate limit hits, fault injection events) |
| `/healthz` | `GET` | Liveness probe |

---

## Load Balancing

```bash
# Register backends for a service
curl -X POST http://servmesh:8095/api/v1/services \
  -d '{"name": "orders-api", "endpoints": ["http://orders-1:3000", "http://orders-2:3000", "http://orders-3:3000"], "locality_zone": "us-east-1a"}'

# Route a request (servmesh selects backend via P2C)
curl -X POST http://servmesh:8095/api/v1/route \
  -d '{"service": "orders-api", "caller_zone": "us-east-1a"}'
# → { "selected_endpoint": "http://orders-2:3000", "latency_p99_ms": 12 }
```

---

## Rate Limiting

```bash
# Set global rate limit for a service
curl -X POST http://servmesh:8095/api/v1/ratelimit/policy \
  -d '{"service": "orders-api", "requests_per_second": 500, "burst": 1000}'
```

ServMesh uses `ServCache` token buckets — the rate limit is enforced globally across all ServMesh nodes:

```
Node 1 ──┐
Node 2 ──┼──→ ServCache token bucket ──→ allow/deny
Node 3 ──┘    (shared global counter)
```

---

## Chaos Fault Injection

```bash
# Inject 200ms latency into 30% of calls to payments-api
curl -X POST http://servmesh:8095/api/v1/chaos/inject \
  -d '{
    "target_service": "payments-api",
    "fault_type": "latency",
    "latency_ms": 200,
    "percentage": 30,
    "duration": "5m"
  }'

# Inject 5% HTTP 503 errors
curl -X POST http://servmesh:8095/api/v1/chaos/inject \
  -d '{"target_service": "inventory-api", "fault_type": "error", "error_code": 503, "percentage": 5, "duration": "2m"}'

# Abort an experiment
curl -X POST http://servmesh:8095/api/v1/chaos/abort/exp-123
```

---

## Getting Started

```bash
docker run -p 8095:8095 \
  -e SERVMESH_SERVCACHE_URL=http://servcache:6379 \
  -e SERVMESH_SERVCONSOLE_WS_URL=ws://servconsole:8083/ws/topology \
  -e SERVMESH_OTEL_ENDPOINT=http://servtrace:4318 \
  ghcr.io/vyuvaraj/servmesh:latest
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVMESH_PORT` | `8095` | HTTP listener port |
| `SERVMESH_SERVCACHE_URL` | — | ServCache URL for distributed rate limit state |
| `SERVMESH_SERVCONSOLE_WS_URL` | — | ServConsole WebSocket URL for topology push |
| `SERVMESH_LOCALITY_ZONE` | — | Availability zone for locality-preference routing |
| `SERVMESH_OTEL_ENDPOINT` | — | OpenTelemetry collector URL |

---

## Enterprise Edition (Planned)

| Feature | Tier |
|---------|------|
| Automatic WireGuard Kernel Tunnel Mesh | EE |
| SPIFFE/SPIRE mTLS Workload Identity Attestation | EE |
