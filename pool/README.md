# ServPool

```bash
docker run -p 8094:8094 ghcr.io/vyuvaraj/servpool:latest
```

`ServPool` is an intelligent, observable database connection pool manager for the **Servverse** ecosystem. It provides read/write splitting, connection health validation, leak detection, query telemetry, prepared statement caching, and pool saturation alerting.

---

## Table of Contents
- [Key Features](#key-features)
- [Architecture](#architecture)
- [API Endpoints](#api-endpoints)
- [Read/Write Split Routing](#readwrite-split-routing)
- [Connection Health & Leak Detection](#connection-health--leak-detection)
- [Query Analytics](#query-analytics)
- [Prepared Statement Cache](#prepared-statement-cache)
- [Getting Started](#getting-started)

---

## Key Features

### 🔀 Read/Write Split Routing
- **Primary for writes, replica for reads**: Automatically routes `SELECT` queries to read replicas and `INSERT`/`UPDATE`/`DELETE` to the primary
- **Configurable replica weighting**: Assign traffic weights per replica (e.g., 70% to replica-1, 30% to replica-2) for load distribution
- **Transaction pinning**: Within an active transaction, all queries are pinned to the primary regardless of query type
- **Replica lag awareness**: Skip replicas with lag > configurable threshold (uses `SHOW SLAVE STATUS` or Postgres `pg_stat_replication`)

### ✅ Connection Health Validation
- **Pre-checkout validation**: Before handing a connection to a caller, ServPool pings it and runs a configurable validation query (e.g., `SELECT 1`) — eliminates "stale connection" errors
- **Unhealthy connection eviction**: Connections that fail validation are immediately evicted and replaced with fresh ones
- **Background health sweeps**: Periodic background sweeps validate idle connections in the pool

### 🔍 Connection Leak Detection
- **Age-based detection**: Connections held longer than configurable `max_checkout_duration` are flagged as leaked
- **Activity-based detection**: Connections with no query activity for `idle_timeout` are reclaimed
- **Goroutine tracking**: Each checkout is tracked with the acquiring goroutine ID and stack trace for leak attribution
- **Forced reclaim**: Leaked connections are forcibly returned to the pool and the offending caller is logged

### 📊 Query Analytics
- **Per-query execution time histogram**: Tracks `p50`, `p75`, `p90`, `p99` query latency per query signature
- **Slow query logger**: Queries exceeding configurable `slow_query_threshold` are logged with full context (query, args, duration, caller)
- **Query normalization**: Normalizes queries by replacing literal values for accurate aggregation
- **Prometheus metrics**: Exposes per-query latency histograms via `/metrics`
- **ServConsole integration**: Pool saturation and query analytics visible in ServConsole dashboard

### 💾 Prepared Statement Cache
- **Multi-dialect support**: Caches prepared statements for PostgreSQL, MySQL, and SQLite
- **Automatic cache invalidation**: Detects schema changes and invalidates affected prepared statements
- **Connection-local cache**: Each connection maintains its own prepared statement cache; ServPool manages the lifecycle
- **Cache hit rate metrics**: Track cache hits vs. prepared statement re-preparations

### 🚨 Saturation Alerting
- **Pool utilization monitoring**: Tracks checked-out vs. total connections as a utilization percentage
- **Wait queue depth**: Monitors how many callers are waiting for a connection — leading indicator of saturation
- **ServConsole alert**: Pushes saturation alerts to ServConsole when utilization exceeds configurable thresholds (e.g., >80%, >95%)
- **Prometheus alerting rules**: Pre-built alert rules for pool saturation and wait queue depth

---

## Architecture

```
Application Caller
      │ checkout connection
      ▼
┌──────────────────────────────────────────────────┐
│                   ServPool                        │
│                                                  │
│  ┌─────────────────────────────────────────────┐ │
│  │  Read/Write Router                          │ │
│  │  SELECT → Replica Pool   │ DML → Primary    │ │
│  └──────────┬──────────────────────────────────┘ │
│             │                                    │
│  ┌──────────▼─────────────────────────────────┐  │
│  │  Pre-checkout Health Validator              │  │
│  │  Ping + Validation Query → evict if fail   │  │
│  └──────────┬─────────────────────────────────┘  │
│             │                                    │
│  ┌──────────▼─────────────────────────────────┐  │
│  │  Leak Detector + Goroutine Tracker          │  │
│  └─────────────────────────────────────────────┘  │
│                                                  │
│  ┌───────────────────┐  ┌──────────────────────┐ │
│  │ Query Analytics   │  │ Prepared Stmt Cache  │ │
│  │ (p99 histograms)  │  │ (per-connection)      │ │
│  └───────────────────┘  └──────────────────────┘ │
└──────────────────────────────────────────────────┘
      │
      ├── Primary DB (writes)
      ├── Replica-1 DB (reads, weight: 70%)
      └── Replica-2 DB (reads, weight: 30%)
```

---

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/pools` | Create a connection pool |
| `GET` | `/api/v1/pools` | List all pools |
| `GET` | `/api/v1/pools/{name}/stats` | Pool stats (utilization, wait queue, active connections) |
| `GET` | `/api/v1/pools/{name}/leaks` | List detected connection leaks |
| `POST` | `/api/v1/pools/{name}/reclaim` | Force-reclaim all leaked connections |
| `GET` | `/api/v1/pools/{name}/slow-queries` | Recent slow queries log |
| `GET` | `/api/v1/pools/{name}/query-stats` | Per-query latency histograms |
| `GET` | `/api/v1/pools/{name}/prepared-cache` | Prepared statement cache contents |
| `/metrics` | `GET` | Prometheus metrics (pool utilization, query latency, cache hit rates) |
| `/healthz` | `GET` | Liveness probe |

---

## Read/Write Split Routing

```bash
# Create a pool with primary + replicas
curl -X POST http://servpool:8094/api/v1/pools \
  -d '{
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
  }'
```

---

## Connection Health & Leak Detection

```bash
# Check pool stats (utilization + wait queue depth)
curl http://servpool:8094/api/v1/pools/orders-db/stats
# → { "total": 50, "active": 38, "idle": 12, "wait_queue": 2, "utilization_pct": 76 }

# View detected leaks
curl http://servpool:8094/api/v1/pools/orders-db/leaks
# → [ { "conn_id": "conn-42", "held_since": "2026-07-26T10:00:00Z", "goroutine": "main.go:84", ... } ]

# Force reclaim leaked connections
curl -X POST http://servpool:8094/api/v1/pools/orders-db/reclaim
```

---

## Query Analytics

```bash
# View p99 latency by query signature
curl http://servpool:8094/api/v1/pools/orders-db/query-stats
# → { "queries": [ { "signature": "SELECT * FROM orders WHERE id = ?", "p50": 3, "p99": 45, "count": 10234 }, ... ] }

# Recent slow queries
curl http://servpool:8094/api/v1/pools/orders-db/slow-queries
```

---

## Prepared Statement Cache

ServPool automatically caches prepared statements per connection:

```go
// Application uses ServPool client — no special code needed
db := servpool.Open("orders-db", "http://servpool:8094")
rows, err := db.Query("SELECT id, total FROM orders WHERE user_id = $1", userID)
// ServPool automatically uses cached prepared statement on subsequent calls
```

---

## Getting Started

```bash
docker run -p 8094:8094 \
  -e SERVPOOL_OTEL_ENDPOINT=http://servtrace:4318 \
  -e SERVPOOL_SERVCONSOLE_URL=http://servconsole:8083 \
  ghcr.io/vyuvaraj/servpool:latest
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVPOOL_PORT` | `8094` | HTTP listener port |
| `SERVPOOL_OTEL_ENDPOINT` | — | OpenTelemetry collector URL |
| `SERVPOOL_SERVCONSOLE_URL` | — | ServConsole URL for saturation alerts |
| `SERVPOOL_DEFAULT_MAX_CONN` | `25` | Default max connections per pool |
| `SERVPOOL_LEAK_CHECK_INTERVAL` | `30s` | How often to run leak detection sweep |
