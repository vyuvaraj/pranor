# ServStore

```bash
docker run -p 7070:7070 ghcr.io/vyuvaraj/servstore:latest
```

`ServStore` is a high-performance, S3-compatible distributed object storage system for the **Servverse** ecosystem. It combines classical cloud storage (erasure coding, multi-region replication) with advanced capabilities: AI-native semantic vector search, browser-local OPFS sync, P2P chunk seeding, and Git-like bucket branching.

---

## Table of Contents
- [Key Features](#key-features)
- [Architecture](#architecture)
- [API Endpoints](#api-endpoints)
- [Vector Search (AI-Native)](#vector-search-ai-native)
- [Bucket Branching](#bucket-branching)
- [Browser / P2P](#browser--p2p)
- [Security](#security)
- [Observability](#observability)
- [Getting Started](#getting-started)
- [Enterprise Edition](#enterprise-edition)

---

## Key Features

### ☁️ Core Object Storage
- **100% S3 Wire Protocol Compatibility**: Drop-in replacement for AWS S3 — works with all existing S3 clients (boto3, aws-sdk, etc.)
- **Erasure Coding (Reed-Solomon)**: Configurable data/parity shard ratios for space-efficient fault tolerance
- **Standalone daemon** (`servstored`): Production-ready daemon with graceful shutdown and health probes
- **Dual-CLI**: `servstore` and `serv store` CLI for bucket/object management
- **Multi-language client SDKs**: Go, Python, TypeScript/JS, Rust

### 🔀 Tiering & Replication
- **Multi-cloud S3 bucket tiering**: Hot/warm/cold tier management — auto-migrate objects to cheaper storage tiers based on last-access time
- **Cold archive mirroring**: Mirror rarely-accessed objects to AWS Glacier, Azure Archive, or GCS Nearline
- **Multi-region active-active CRDT replication**: Conflict-free replicated data types for last-write-wins semantics across regions
- **Cross-region active-active bucket replication**: Sync buckets across cloud regions with configurable consistency guarantees

### 🤖 AI-Native Vector Search
- **Automatic embedding generation**: Text objects are automatically embedded on `PUT` using configurable embedding models
- **Hybrid keyword + vector semantic search (RRF)**: Reciprocal Rank Fusion combines BM25 keyword scores with vector similarity for optimal relevance
- **Per-bucket vector index namespace management**: Isolated vector index per bucket; configurable distance metrics (Cosine, Euclidean, DotProduct)
- **ANN query API**: `k`-nearest-neighbor queries with min-score filtering, metadata filters, and hybrid mode
- **Persistent mmap-backed HNSW graph engine**: High-performance Hierarchical Navigable Small World graph with incremental node insertion and mmap persistence for zero-copy access

### 🌿 Bucket Branching (Git-like)
- **Copy-on-Write (CoW) virtual metadata pointer engine**: Branch a bucket in O(1) — no data copy; branches share storage until modified
- **Bucket branch diff & merge**: `servstore diff branch-a branch-b` shows changed objects; merge branches with conflict resolution
- **Isolated virtual namespace router**: Each branch gets its own S3-compatible namespace; branches are fully isolated
- **REST API**: `POST /api/v1/buckets/{name}/branch`, `POST /api/v1/buckets/{name}/merge`
- **CLI**: `servstore branch create`, `servstore branch diff`, `servstore branch merge`

### 🌐 Browser & P2P
- **OPFS local sync** (`@servverse/store-wasm`): Browser-local object storage using Origin Private File System; syncs to server when online
- **WebTorrent P2P chunk seeder**: Seed object chunks via WebTorrent — reduce CDN egress costs
- **WebRTC peer signaling relay**: Broker WebRTC connections between peers for direct chunk transfer
- **P2P SHA-256 integrity verification**: All chunks verified cryptographically before acceptance

### 🔍 S3 Select
- **S3 Select engine**: Query CSV, JSON, and Parquet objects with SQL expressions without downloading entire objects

---

## Architecture

```
┌────────────────────────────────────────────────────────────┐
│                        ServStore                            │
│                                                            │
│  ┌─────────────────────────────────────────────────────┐  │
│  │              S3 Wire Protocol Router                 │  │
│  │  GET/PUT/DELETE/LIST/SELECT compatible with AWS S3   │  │
│  └───────────────────────┬─────────────────────────────┘  │
│                           │                                │
│  ┌────────────┐  ┌────────▼──────┐  ┌────────────────┐   │
│  │ CoW Branch │  │  Object Store │  │  Vector Index  │   │
│  │  Namespaces│  │  (Reed-Solomon│  │  (HNSW + RRF)  │   │
│  └────────────┘  │   Erasure)    │  └────────────────┘   │
│                  └───────┬───────┘                         │
│  ┌────────────┐  ┌───────▼───────┐  ┌────────────────┐   │
│  │   S3 Select│  │  Tiered Store │  │  CRDT Repl.    │   │
│  │   Engine   │  │  Hot/Warm/Cold│  │  Multi-Region  │   │
│  └────────────┘  └───────────────┘  └────────────────┘   │
│                                                            │
│  ┌────────────────────────────────────────────────────┐   │
│  │         P2P / OPFS / WebRTC Layer (Browser)         │   │
│  └────────────────────────────────────────────────────┘   │
└────────────────────────────────────────────────────────────┘
```

---

## API Endpoints

### S3 Compatible (use any S3 client)
| Method | Path | Description |
|--------|------|-------------|
| `PUT` | `/{bucket}/{key}` | Upload object (triggers auto-embedding if text) |
| `GET` | `/{bucket}/{key}` | Download object |
| `DELETE` | `/{bucket}/{key}` | Delete object |
| `GET` | `/{bucket}?list-type=2` | List objects in bucket |
| `POST` | `/{bucket}/{key}?select` | S3 Select query (CSV/JSON/Parquet) |

### ServStore-Specific APIs
| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/buckets` | Create bucket |
| `POST` | `/api/v1/buckets/{name}/branch` | Create a CoW branch |
| `POST` | `/api/v1/buckets/{name}/merge` | Merge a branch back |
| `GET` | `/api/v1/buckets/{name}/diff` | Diff two branches |
| `POST` | `/api/v1/search/vector` | Vector ANN search |
| `POST` | `/api/v1/search/hybrid` | Hybrid keyword+vector search (RRF) |
| `GET` | `/api/v1/search/namespaces` | List vector index namespaces per bucket |
| `GET` | `/api/v1/tiers/{bucket}/policy` | Get tiering policy |
| `PUT` | `/api/v1/tiers/{bucket}/policy` | Set tiering policy |
| `/metrics` | `GET` | Prometheus metrics |

---

## Vector Search (AI-Native)

Objects uploaded to enabled buckets are automatically embedded:

```bash
# Upload a text document — embedding generated automatically
aws s3 cp docs/manual.txt s3://my-bucket/manual.txt \
  --endpoint-url http://servstore:7070

# Hybrid search (keyword + vector, RRF combined)
curl -X POST http://servstore:7070/api/v1/search/hybrid \
  -d '{"bucket": "my-bucket", "query": "installation guide", "k": 5, "metric": "cosine"}'

# Pure vector ANN search
curl -X POST http://servstore:7070/api/v1/search/vector \
  -d '{"bucket": "my-bucket", "vector": [0.12, -0.34, ...], "k": 10, "min_score": 0.8}'
```

### Vector Index Configuration

```json
{
  "bucket": "my-bucket",
  "vector_index": {
    "enabled": true,
    "embedding_model": "text-embedding-3-small",
    "dimensions": 1536,
    "metric": "cosine",
    "hnsw": { "m": 16, "ef_construction": 200 }
  }
}
```

---

## Bucket Branching

```bash
# Create a branch (instant, no data copy)
servstore branch create my-bucket --name feature-x

# Make changes to the branch
aws s3 cp new-file.txt s3://my-bucket@feature-x/new-file.txt

# Diff branch vs main
servstore branch diff my-bucket feature-x

# Merge branch back
servstore branch merge my-bucket --source feature-x --into main
```

---

## Browser / P2P

```bash
npm install @servverse/store-wasm
```

```typescript
import { ServStore } from '@servverse/store-wasm';

const store = new ServStore({ bucket: 'my-bucket', syncUrl: 'https://store.servverse.net' });

// Works offline via OPFS
await store.put('key', new Uint8Array([1, 2, 3]));
const data = await store.get('key');

// P2P chunk seeding (reduces server egress)
await store.enableP2PSeed({ torrentTracker: 'wss://tracker.servverse.net' });
```

---

## Security

| Feature | Description |
|---------|-------------|
| Blind-Store E2EE | Client-side encryption; server never sees plaintext |
| FIPS 140-3 + HSM Key Unsealing | Hardware security module key management |
| WORM Object Lock | Write-Once-Read-Many immutable objects |
| Merkle Immutability Ledger | Tamper-evident audit chain for every object write |
| io_uring + Direct I/O | Bypasses page cache for NVMe-level throughput (EE) |

---

## Observability

- **Prometheus `/metrics`**: Object throughput, IOPS, cache hit rates, vector index query latency, tiering migration stats
- **OTel tracing**: Per-request spans for upload, download, search, and compaction operations
- **ServConsole Inspector**: Bucket browser, vector index namespace management, tiering policy editor

---

## Getting Started

```bash
docker run -p 7070:7070 \
  -e SERVSTORE_DATA_DIR=/data \
  -e SERVSTORE_ERASURE_DATA_SHARDS=6 \
  -e SERVSTORE_ERASURE_PARITY_SHARDS=2 \
  -e SERVSTORE_OTEL_ENDPOINT=http://servtrace:4318 \
  -v store-data:/data \
  ghcr.io/vyuvaraj/servstore:latest

# Use with any S3 client
export AWS_ACCESS_KEY_ID=servstore
export AWS_SECRET_ACCESS_KEY=servstore
aws s3 mb s3://my-bucket --endpoint-url http://localhost:7070
aws s3 cp myfile.txt s3://my-bucket/ --endpoint-url http://localhost:7070
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVSTORE_PORT` | `7070` | HTTP listener port |
| `SERVSTORE_DATA_DIR` | `./data` | Object storage root directory |
| `SERVSTORE_ERASURE_DATA_SHARDS` | `6` | Reed-Solomon data shards |
| `SERVSTORE_ERASURE_PARITY_SHARDS` | `2` | Reed-Solomon parity shards |
| `SERVSTORE_VECTOR_ENABLED` | `false` | Enable auto-embedding & HNSW index |
| `SERVSTORE_EMBEDDING_MODEL` | — | Embedding model endpoint URL |
| `SERVSTORE_OTEL_ENDPOINT` | — | OpenTelemetry collector URL |
| `SERVSTORE_S3_TIER_COLD_URL` | — | Cold tier S3 endpoint |

---

## Enterprise Edition

| Feature | Tier |
|---------|------|
| Blind-Store E2EE & FIPS HSM | EE |
| Cross-Region Active-Active Replication | EE |
| io_uring & Direct I/O NVMe Acceleration | EE |
| WORM Object Lock & Merkle Ledger | EE |
| Enterprise Multi-Tenant CoW Encryption | EE |
| Enterprise P2P Token-Gated Content DRM | EE |
