# Pranor Vault

[![S3 Conformance](https://img.shields.io/badge/S3_Conformance-96%2F96_Operations_Pass-10b981?style=for-the-badge&logo=amazons3)](pkg/s3/s3_compliance_test.go)
[![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=for-the-badge&logo=go)](go.mod)

```bash
docker compose up -d
```

`Pranor Vault` is a high-performance, S3-compatible distributed object storage system for the **Pranor** ecosystem. It combines classical cloud storage (erasure coding, multi-region replication) with advanced capabilities: AI-native semantic vector search, browser-local OPFS sync, P2P chunk seeding, and Git-like bucket branching.

---

## Quickstart (S3 & AI Vector Search in 30 Seconds)

### 1. Launch Pranor Vault Standalone Daemon & Admin Console
```bash
docker compose up -d
# S3 API listening at http://localhost:9000
# Admin Console UI listening at http://localhost:9001/ui/
```

### 2. Standard S3 Operations (via AWS S3 CLI or `servstore` CLI)
```bash
export AWS_ACCESS_KEY_ID=minioadmin
export AWS_SECRET_ACCESS_KEY=minioadmin

# Create a bucket and upload a document via AWS CLI
aws s3 mb s3://knowledge --endpoint-url http://localhost:9000
aws s3 cp ./deploy/helm/servstore/README.md s3://knowledge/deploy-guide.md --endpoint-url http://localhost:9000

# Or use the unified servstore CLI
servstore mb s3://knowledge
servstore put knowledge deploy-guide.md ./deploy/helm/servstore/README.md
servstore ls knowledge
```

### 3. AI-Native Semantic Vector Search (End-to-End)
Text uploaded to Pranor Vault is automatically indexed and vectorized on `PUT`. Query semantically without external vector databases:

```bash
curl -X POST http://localhost:9000/api/v1/search/hybrid \
  -H "Content-Type: application/json" \
  -d '{
    "bucket": "knowledge",
    "query": "how to deploy helm chart to Kubernetes",
    "k": 5
  }'
```

---

## Table of Contents
- [Key Features](#key-features)
- [Architecture](#architecture)
- [API Endpoints](#api-endpoints)
- [Unified CLI Reference](#unified-cli-reference)
- [Vector Search (AI-Native)](#vector-search-ai-native)
- [Bucket Branching](#bucket-branching)
- [Browser / P2P](#browser--p2p)
- [Security](#security)
- [Observability](#observability)
- [Getting Started & Docker Compose](#getting-started--docker-compose)
- [Enterprise Edition](#enterprise-edition)

---

## Key Features

### ☁️ Core Object Storage
- **100% S3 Wire Protocol Compatibility**: Drop-in replacement for AWS S3 — works with all existing S3 clients (aws-cli, boto3, aws-sdk-js, etc.)
- **Erasure Coding (Reed-Solomon)**: Configurable data/parity shard ratios for space-efficient fault tolerance
- **Standalone daemon** (`servstored`): Production-ready daemon serving S3 API (`:9000`) and Admin Console (`:9001`)
- **Unified CLI (`servstore`)**: Single CLI for object storage management, IAM policies, and cluster administration
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
- **OPFS local sync** (`@pranor/store-wasm`): Browser-local object storage using Origin Private File System; syncs to server when online
- **WebTorrent P2P chunk seeder**: Seed object chunks via WebTorrent — reduce CDN egress costs
- **WebRTC peer signaling relay**: Broker WebRTC connections between peers for direct chunk transfer
- **P2P SHA-256 integrity verification**: All chunks verified cryptographically before acceptance

### 🔍 S3 Select
- **S3 Select engine**: Query CSV, JSON, and Parquet objects with SQL expressions without downloading entire objects

---

## Architecture

```
┌────────────────────────────────────────────────────────────┐
│                        Pranor Vault                            │
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

### Pranor Vault-Specific APIs
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

## Unified CLI Reference (`servstore`)

Pranor Vault ships a single, unified CLI tool (`servstore`) that connects to both the S3 API endpoint and the Admin management API:

```bash
# Global flags
servstore --endpoint http://localhost:9000 --admin-endpoint http://localhost:9001 <command>

# S3 & Data Management
servstore mb s3://my-bucket                    # Make bucket
servstore rb s3://my-bucket                    # Remove bucket
servstore ls s3://my-bucket                    # List bucket contents
servstore put my-bucket photo.jpg ./photo.jpg  # Upload object
servstore get my-bucket photo.jpg ./dest.jpg   # Download object
servstore rm my-bucket photo.jpg               # Delete object
servstore lock my-bucket photo.jpg 30d         # WORM Object Lock (30 days)

# Admin & Server Health
servstore status                               # Daemon status & uptime
servstore admin-buckets                        # List buckets via Admin API
```

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
npm install @pranor/store-wasm
```

```typescript
import { Pranor Vault } from '@pranor/store-wasm';

const store = new Pranor Vault({ bucket: 'my-bucket', syncUrl: 'https://store.pranor.net' });

// Works offline via OPFS
await store.put('key', new Uint8Array([1, 2, 3]));
const data = await store.get('key');

// P2P chunk seeding (reduces server egress)
await store.enableP2PSeed({ torrentTracker: 'wss://tracker.pranor.net' });
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
- **Pranor Console Inspector**: Bucket browser, vector index namespace management, tiering policy editor

---

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
