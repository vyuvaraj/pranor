# Pranor Pool — Database Connection Proxy

**Version:** 1.0.0  
**Module Path:** `github.com/vyuvaraj/pranor/pool`  
**Default Port:** 8097  
**License:** AGPL-3.0 (OSS) / Enterprise License (EE with pgvector accelerator & multi-dialect)

---

## Overview

Pranor Pool is an intelligent, observable database connection pool manager for the Pranor ecosystem. It provides read/write splitting, connection health validation, leak detection, query telemetry, prepared statement caching, pool saturation alerting, and multi-dialect support for PostgreSQL, MySQL, and SQLite.

Pranor Pool can run as:
- A **standalone binary** providing connection pooling for any PostgreSQL/MySQL application
- An **integrated module** within the Pranor ecosystem with OTel tracing, Console dashboards, and Lock-coordinated DDL migrations

---

## Key Features

| Feature | Description |
|---------|-------------|
| **Read/Write Split** | Auto-routes SELECTs to replicas, writes to primary |
| **Replica Weighting** | Configurable traffic distribution across replicas |
| **Transaction Pinning** | All queries within a transaction pinned to primary |
| **Replica Lag Awareness** | Skip replicas exceeding configurable lag threshold |
| **Pre-checkout Validation** | Ping + validation query before handing connections to callers |
| **Leak Detection** | Age-based and activity-based detection with goroutine stack traces |
| **Query Analytics** | Per-query p50/p75/p90/p99 latency histograms |
| **Slow Query Logger** | Queries exceeding threshold logged with full context |
| **Prepared Statement Cache** | Per-connection cache with automatic invalidation on schema change |
| **Saturation Alerting** | Pool utilization and wait queue depth alerts to Console |

---

## Architecture

```mermaid
graph TD

    subgraph AppCallers ["🌐 Microservice Connection Request"]
        App["Application Microservice Caller"]
        PoolClient["Pranor Pool Go/Python/Java Client"]
    end

    subgraph PoolCore ["⚡ Core Connection Routing and Health Engine"]
        RWRouter["Read/Write Query Router"]
        HealthCheck["Pre-Checkout Validation Engine"]
        LeakDetector["Connection Leak and Goroutine Stack Tracker"]
        StmtCache["Per-Connection Prepared Statement Cache"]
        VectorOffload["PostgreSQL pgvector Accelerator"]
    end

    subgraph DBClusters ["💾 Heterogeneous Relational DB Tier"]
        PrimaryDB["Primary RDBMS"]
        ReplicaPool["Weighted Replica Pool"]
    end

    App --> PoolClient
    PoolClient --> RWRouter
    RWRouter --> HealthCheck
    HealthCheck --> LeakDetector
    LeakDetector --> StmtCache
    StmtCache --> VectorOffload
    VectorOffload --> PrimaryDB
    VectorOffload --> ReplicaPool
```

### Connection Checkout, Read/Write Split & Leak Detection Sequence Flow

```mermaid
sequenceDiagram
    autonumber
    participant App as Application Microservice
    participant Pool as Pranor Pool Manager
    participant Leak as Goroutine Leak Tracker
    participant Stmt as Prepared Statement Cache
    participant DB as Target RDBMS (Primary / Replica)

    App->>Pool: Checkout Connection (Query: "SELECT * FROM users WHERE id = $1")
    Pool->>Pool: Inspect SQL Query Type (Read Query -> Route to Replica Pool)
    Pool->>Leak: Register Goroutine Stack & Start 30s Max-Hold Timer
    Pool->>Stmt: Lookup Cached Prepared Statement ("stmt_users_by_id")
    Stmt-->>Pool: Prepared Statement Handle Ready
    Pool->>DB: Execute Query on Replica DB Instance
    DB-->>Pool: Query Result Set Returned (p99 latency: 1.2ms)
    Pool->>Leak: Cancel Max-Hold Leak Timer & Return Connection to Pool
    Pool-->>App: Connection Released & Stats Updated
```

### Ecosystem Cross-Module Integration

Pranor Pool provides intelligent database proxying across the Pranor ecosystem:

- **Pranor Lock**: Coordinates zero-downtime online DDL schema migrations, holding exclusive fencing token leases during migrations.
- **Pranor Trace**: Annotates SQL queries with OpenTelemetry spans, recording query normalization histograms and slow query stack traces.
- **Pranor Vault**: Connects seamlessly to PostgreSQL `pgvector` instances, managing connection pools for S3 vector metadata storage.
- **Pranor Console**: Displays real-time database connection saturation heatmaps, active wait-queue depth, and 1-click connection leak reclaims.

---

## Installation & Deployment

### Binary

```bash
cd pranor/pool
go build -o pranor-pool .
./pranor-pool --port 8097
```

### Docker

```bash
docker run -p 8097:8097 ghcr.io/vyuvaraj/pranor-pool:latest
```

### With OTel and Console

```bash
docker run -p 8097:8097 \
  -e PRANOR_POOL_OTEL_ENDPOINT=http://pranor-trace:8090 \
  -e PRANOR_POOL_PRANOR_CONSOLE_URL=http://pranor-console:8083 \
  ghcr.io/vyuvaraj/pranor-pool:latest
```

### As Part of Pranor Ecosystem

When running under the Pranor platform, Pool integrates automatically with Lock (DDL coordination), Trace (query spans), and Console (saturation dashboards).

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PRANOR_POOL_PORT` | `8097` | HTTP listener port |
| `PRANOR_POOL_OTEL_ENDPOINT` | — | OpenTelemetry collector URL |
| `PRANOR_POOL_PRANOR_CONSOLE_URL` | — | Pranor Console URL for saturation alerts |
| `PRANOR_POOL_DEFAULT_MAX_CONN` | `25` | Default max connections per pool |
| `PRANOR_POOL_LEAK_CHECK_INTERVAL` | `30s` | How often to run leak detection sweep |

### YAML Config (`pool.yaml`)

```yaml
port: "8097"
otel_endpoint: "http://pranor-trace:8090"
console_url: "http://pranor-console:8083"
default_max_connections: 25
leak_check_interval: "30s"
slow_query_threshold_ms: 100
```

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `8097` | HTTP listen port |

---

## API Reference

**Base URL:** `http://localhost:8097`

### POST /api/v1/pools

Create a connection pool.

**Request:**

```json
{
  "name": "orders-db",
  "primary": "postgres://user:pass@primary:5432/orders",
  "replicas": [
    { "dsn": "postgres://user:pass@replica1:5432/orders", "weight": 70 },
    { "dsn": "postgres://user:pass@replica2:5432/orders", "weight": 30 }
  ],
  "max_connections": 50,
  "min_idle": 5,
  "validation_query": "SELECT 1",
  "max_checkout_duration": "30s",
  "slow_query_threshold_ms": 100
}
```

**Response (201):**

```json
{
  "status": "created",
  "name": "orders-db",
  "max_connections": 50
}
```

---

### GET /api/v1/pools/{name}/stats

Pool stats — utilization, wait queue, active connections.

**Response (200):**

```json
{
  "name": "orders-db",
  "total": 50,
  "active": 38,
  "idle": 12,
  "wait_queue": 2,
  "utilization_pct": 76
}
```

---

### GET /api/v1/pools/{name}/leaks

List detected connection leaks.

**Response (200):**

```json
{
  "leaks": [
    {
      "conn_id": "conn-42",
      "held_since": "2026-08-01T10:00:00Z",
      "duration_s": 45,
      "goroutine": "main.go:84",
      "stack_trace": "goroutine 42 [running]:\nmain.handleOrder(...)"
    }
  ]
}
```

---

### POST /api/v1/pools/{name}/reclaim

Force-reclaim all leaked connections.

**Response (200):**

```json
{
  "status": "reclaimed",
  "reclaimed_count": 3
}
```

---

### GET /api/v1/pools/{name}/query-stats

Per-query latency histograms.

**Response (200):**

```json
{
  "queries": [
    {
      "signature": "SELECT * FROM orders WHERE id = ?",
      "p50_ms": 3,
      "p75_ms": 8,
      "p90_ms": 22,
      "p99_ms": 45,
      "count": 10234
    }
  ]
}
```

---

### GET /api/v1/pools/{name}/slow-queries

Recent slow queries.

**Response (200):**

```json
{
  "queries": [
    {
      "query": "SELECT * FROM orders JOIN items ON ...",
      "duration_ms": 340,
      "timestamp": "2026-08-01T10:01:30Z",
      "caller": "handlers.go:156"
    }
  ]
}
```

---

### GET /healthz

Liveness probe.

```json
{"status":"UP","service":"pranor-pool","version":"1.0.0"}
```

---

## Security

### Standalone Mode

In standalone mode, Pool provides unauthenticated connection pooling. DSN credentials are stored in memory only.

### Ecosystem Mode (Full Auth Stack)

When running within the Pranor ecosystem:

1. **JWT Auth** — validates Bearer tokens for pool management API
2. **Tenant Isolation** — pools scoped per tenant namespace
3. **OTel Tracing** — every query generates a trace span
4. **Credential Injection** — DSN passwords can be sourced from Pranor Secret

### Connection Security

- **TLS to database** — supports `sslmode=require` in PostgreSQL DSNs
- **Credential rotation** — integrates with Pranor Secret for dynamic password rotation
- **No credential exposure** — DSN passwords never exposed in API responses

---

## Observability

### Prometheus Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `pranor_pool_connections_active` | Gauge | Currently checked-out connections |
| `pranor_pool_connections_idle` | Gauge | Idle connections in pool |
| `pranor_pool_wait_queue_depth` | Gauge | Callers waiting for a connection |
| `pranor_pool_utilization_pct` | Gauge | Pool utilization percentage |
| `pranor_pool_query_duration_ms` | Histogram | Query execution latency |
| `pranor_pool_leaks_detected_total` | Counter | Connection leaks detected |
| `pranor_pool_stmt_cache_hits_total` | Counter | Prepared statement cache hits |
| `pranor_pool_slow_queries_total` | Counter | Slow queries logged |

### OpenTelemetry Tracing

Pool emits spans for:
- `pool.checkout` — connection checkout with routing decision
- `pool.query` — SQL query execution
- `pool.leak.detect` — leak detection event
- `pool.health.validate` — connection validation

### Logging

Structured JSON logs with fields: `level`, `timestamp`, `trace_id`, `pool`, `query_signature`, `duration_ms`, `connection_id`.

---

## Enterprise Edition

| Feature | OSS | EE |
|---------|:---:|:--:|
| Connection pooling (PostgreSQL, MySQL, SQLite) | ✓ | ✓ |
| Read/write split routing | ✓ | ✓ |
| Replica weighting | ✓ | ✓ |
| Pre-checkout validation | ✓ | ✓ |
| Leak detection | ✓ | ✓ |
| Query analytics | ✓ | ✓ |
| Slow query logger | ✓ | ✓ |
| Prepared statement cache | ✓ | ✓ |
| Saturation alerting | ✓ | ✓ |
| PostgreSQL pgvector accelerator | — | ✓ |
| Multi-dialect federation (cross-DB routing) | — | ✓ |
| AI query optimization advisor | — | ✓ |
| Connection pool sharding | — | ✓ |

---

## Operational Runbook

### Pool saturation (high utilization)

1. Check `/api/v1/pools/{name}/stats` for utilization percentage
2. Monitor `pranor_pool_wait_queue_depth` — if growing, pool is undersized
3. Increase `max_connections` in pool configuration
4. Check for connection leaks: `GET /api/v1/pools/{name}/leaks`
5. Force reclaim leaks: `POST /api/v1/pools/{name}/reclaim`

### Connection leaks accumulating

1. Monitor `pranor_pool_leaks_detected_total` metric
2. Review leak stack traces: `GET /api/v1/pools/{name}/leaks`
3. Identify code paths that checkout but don't release connections
4. Reduce `max_checkout_duration` to catch leaks sooner
5. Ensure all `defer conn.Close()` patterns are correct in application code

### Replica lag causing stale reads

1. Check replica lag via database metrics
2. Configure lag threshold in pool — lagging replicas auto-excluded
3. Monitor how many queries fall back to primary due to lag
4. Consider adding more replicas or optimizing replication

### Slow queries increasing

1. Review `/api/v1/pools/{name}/slow-queries` for patterns
2. Check `pranor_pool_query_duration_ms` histogram for p99 growth
3. Use query normalization to identify expensive query signatures
4. Work with DBA to add indexes or optimize queries
5. Consider prepared statement cache to reduce parse overhead
