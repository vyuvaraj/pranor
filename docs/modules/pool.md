# Pranor Pool

```bash
docker run -p 8094:8094 ghcr.io/vyuvaraj/pranor-pool:latest
```

`Pranor Pool` is an intelligent, observable database connection pool manager for the **Pranor** ecosystem. It provides read/write splitting, connection health validation, leak detection, query telemetry, prepared statement caching, and pool saturation alerting.

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
- **Pre-checkout validation**: Before handing a connection to a caller, Pranor Pool pings it and runs a configurable validation query (e.g., `SELECT 1`) — eliminates "stale connection" errors
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
- **Pranor Console integration**: Pool saturation and query analytics visible in Pranor Console dashboard

### 💾 Prepared Statement Cache
- **Multi-dialect support**: Caches prepared statements for PostgreSQL, MySQL, and SQLite
- **Automatic cache invalidation**: Detects schema changes and invalidates affected prepared statements
- **Connection-local cache**: Each connection maintains its own prepared statement cache; Pranor Pool manages the lifecycle
- **Cache hit rate metrics**: Track cache hits vs. prepared statement re-preparations

### 🚨 Saturation Alerting
- **Pool utilization monitoring**: Tracks checked-out vs. total connections as a utilization percentage
- **Wait queue depth**: Monitors how many callers are waiting for a connection — leading indicator of saturation
- **Pranor Console alert**: Pushes saturation alerts to Pranor Console when utilization exceeds configurable thresholds (e.g., >80%, >95%)
- **Prometheus alerting rules**: Pre-built alert rules for pool saturation and wait queue depth

---

## Architecture

```mermaid
graph TD
    classDef client fill:#1e293b,stroke:#38bdf8,stroke-width:2px,color:#fff;
    classDef engine fill:#0f172a,stroke:#0d9488,stroke-width:2px,color:#fff;
    classDef storage fill:#1e1b4b,stroke:#6366f1,stroke-width:2px,color:#fff;
    classDef monitor fill:#1e293b,stroke:#64748b,stroke-width:1px,color:#fff;

    subgraph AppCallers ["🌐 Microservice Connection Request"]
        App["Application Microservice Caller"] :::client
        PoolClient["Pranor Pool Go/Python/Java Client"] :::client
    end

    subgraph PoolCore ["⚡ Core Connection Routing & Health Engine"]
        RWRouter["Read/Write Query Router<br/><i>(SELECT → Replica | DML → Primary)</i>"] :::engine
        HealthCheck["Pre-Checkout Validation Engine<br/><i>(Ping & Active Connection Evictor)</i>"] :::engine
        LeakDetector["Connection Leak & Goroutine Stack Tracker"] :::engine
        StmtCache["Per-Connection Prepared Statement Cache"] :::engine
        VectorOffload["PostgreSQL pgvector Accelerator<br/><i>(Enterprise EE)</i>"] :::engine
    end

    subgraph DBClusters ["💾 Heterogeneous Relational DB Tier"]
        PrimaryDB["Primary RDBMS<br/><i>(PostgreSQL / MySQL Writes)</i>"] :::storage
        ReplicaPool["Weighted Replica Pool<br/><i>(70% Replica-1 / 30% Replica-2 Reads)</i>"] :::storage
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
curl -X POST http://pranor-pool:8094/api/v1/pools \
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
curl http://pranor-pool:8094/api/v1/pools/orders-db/stats
# → { "total": 50, "active": 38, "idle": 12, "wait_queue": 2, "utilization_pct": 76 }

# View detected leaks
curl http://pranor-pool:8094/api/v1/pools/orders-db/leaks
# → [ { "conn_id": "conn-42", "held_since": "2026-07-26T10:00:00Z", "goroutine": "main.go:84", ... } ]

# Force reclaim leaked connections
curl -X POST http://pranor-pool:8094/api/v1/pools/orders-db/reclaim
```

---

## Query Analytics

```bash
# View p99 latency by query signature
curl http://pranor-pool:8094/api/v1/pools/orders-db/query-stats
# → { "queries": [ { "signature": "SELECT * FROM orders WHERE id = ?", "p50": 3, "p99": 45, "count": 10234 }, ... ] }

# Recent slow queries
curl http://pranor-pool:8094/api/v1/pools/orders-db/slow-queries
```

---

## Prepared Statement Cache

Pranor Pool automatically caches prepared statements per connection:

```go
// Application uses Pranor Pool client — no special code needed
db := pranor-pool.Open("orders-db", "http://pranor-pool:8094")
rows, err := db.Query("SELECT id, total FROM orders WHERE user_id = $1", userID)
// Pranor Pool automatically uses cached prepared statement on subsequent calls
```

---

## Getting Started

```bash
docker run -p 8094:8094 \
  -e PRANOR_POOL_OTEL_ENDPOINT=http://pranor-trace:4318 \
  -e PRANOR_POOL_PRANOR_CONSOLE_URL=http://pranor-console:8083 \
  ghcr.io/vyuvaraj/pranor-pool:latest
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PRANOR_POOL_PORT` | `8094` | HTTP listener port |
| `PRANOR_POOL_OTEL_ENDPOINT` | — | OpenTelemetry collector URL |
| `PRANOR_POOL_PRANOR_CONSOLE_URL` | — | Pranor Console URL for saturation alerts |
| `PRANOR_POOL_DEFAULT_MAX_CONN` | `25` | Default max connections per pool |
| `PRANOR_POOL_LEAK_CHECK_INTERVAL` | `30s` | How often to run leak detection sweep |
