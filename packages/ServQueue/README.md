# ServQueue

```bash
docker run -p 9090:9090 ghcr.io/vyuvaraj/servqueue:latest
```

`ServQueue` is a full-featured, enterprise-grade message broker for the **Servverse** ecosystem. It supports server-side STOMP brokering, browser-local OPFS-backed queueing, multi-protocol adapters (Kafka wire, MQTT v5), and advanced security (FIPS 140-3, post-quantum cryptography, blind E2EE).

---

## Table of Contents
- [Key Features](#key-features)
- [Architecture](#architecture)
- [API Endpoints](#api-endpoints)
- [Protocols Supported](#protocols-supported)
- [Browser / OPFS Features](#browser--opfs-features)
- [Security](#security)
- [Observability](#observability)
- [Kubernetes & Distribution](#kubernetes--distribution)
- [Getting Started](#getting-started)
- [Enterprise Edition](#enterprise-edition)

---

## Key Features

### 📨 Core Broker
- **STOMP 1.2 message broker**: Topic/queue routing with fan-out, competing consumers, and durable subscriptions
- **Exactly-once delivery semantics**: Idempotent message IDs with deduplication window
- **DLQ + Exponential Backoff Engine**: Failed messages automatically moved to Dead Letter Queue with configurable retry policies; exponential backoff with jitter
- **Point-in-time event replay**: Replay messages from any historical offset on demand
- **Schema Registry & Validation**: Embedded schema registry for message contract enforcement (Avro/JSON Schema/Protobuf); schema evolution with compatibility checks
- **Atomic Multi-Topic Transactions**: ACID-style multi-topic publish/consume transactions
- **Cooperative Consumer Rebalancing**: Sticky partition assignment with graceful rebalance on consumer join/leave

### 🌐 Browser & OPFS (Local-First)
- **OPFS Storage Driver** (`pkg/opfs`): Full browser-native persistent queue using Origin Private File System
- **WASM/JS FFI bindings** (`@servverse/queue-wasm`): Use ServQueue from the browser with a TypeScript SDK
- **SharedWorker multi-tab coordination**: Single broker across all browser tabs via SharedWorker
- **Multi-tab OPFS leader election**: `navigator.locks`-based lease protocol ensures only one tab acts as queue leader at a time
- **Client-side AES-256-GCM encryption at rest**: Messages encrypted before writing to OPFS
- **WebTransport HTTP/3 QUIC relay**: Browser outbox relay over QUIC for low-latency connectivity
- **Offline outbox & reconnect relay**: Queue messages offline; auto-relay when connectivity restores
- **Auto-compaction & quota manager**: Automatic OPFS quota management with configurable size limits
- **Client-side WASM stream filters**: Run sandboxed WASM modules to transform/filter messages in-browser
- **Persistent storage eviction safeguard**: Priority-based eviction prevents silent data loss at storage limits

### 🗜️ Storage & Compaction
- **Write-Ahead Log (WAL)** with corruption recovery and CRC checksums
- **Topic Log Compaction Policy Engine**: Key-based compaction (retain only latest value per key), tombstone purging, TTL-based retention
- **Tiered cloud storage offloading**: Hot/warm/cold tier management with S3-compatible backend
- **Automated storage tiering & compaction**: Background compaction scheduler with configurable policies

### 📡 Protocol Adapters
- **Kafka Wire Protocol Compatibility**: Drop-in replacement for Kafka consumers/producers (Kafka binary protocol)
- **MQTT v5.0 IoT Gateway**: Full MQTT v5 adapter — QoS 0/1/2, retain, session persistence, will messages

### 🔁 Streaming & CDC
- **Change Data Capture (CDC) Engine**: Database change event streaming (row-level insert/update/delete events)
- **Real-time Stream SQL Windowing**: Tumbling, sliding, and session windows with aggregations (COUNT, SUM, AVG)

### 🏢 Multi-Tenant
- **Multi-tenant VHosts & rate quotas**: Isolated virtual hosts per tenant with per-tenant rate limits and storage quotas
- **Zero-Trust OAuth2 & SPIFFE auth**: Per-connection authentication with SPIFFE workload identity attestation

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                       ServQueue                              │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ STOMP Broker │  │ Kafka Compat │  │  MQTT v5 Gateway │  │
│  └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘  │
│         └─────────────────┼──────────────────-─┘            │
│                           ▼                                 │
│  ┌────────────────────────────────────────────────────────┐ │
│  │              Schema Registry & Validation              │ │
│  └────────────────────────────┬───────────────────────────┘ │
│                               ▼                             │
│  ┌────────────────────────────────────────────────────────┐ │
│  │    WAL Storage Engine │ Compaction │ Tiered Offload     │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                             │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────────┐ │
│  │ DLQ Engine  │  │ CDC Streamer │  │ SQL Window Engine  │ │
│  └─────────────┘  └──────────────┘  └────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

---

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/topics` | Create a topic |
| `GET` | `/api/v1/topics` | List all topics |
| `POST` | `/api/v1/publish` | Publish a message to a topic |
| `POST` | `/api/v1/subscribe` | Subscribe to a topic (SSE or WebSocket) |
| `GET` | `/api/v1/consumers` | List consumer groups |
| `GET` | `/api/v1/consumers/{group}/lag` | Consumer group lag per partition |
| `POST` | `/api/v1/schemas` | Register a message schema |
| `GET` | `/api/v1/schemas/{topic}` | Get schema for a topic |
| `GET` | `/api/v1/dlq/{topic}` | Browse DLQ for a topic |
| `POST` | `/api/v1/dlq/{topic}/replay` | Replay DLQ messages |
| `POST` | `/api/v1/replay` | Point-in-time replay from offset |
| `POST` | `/api/v1/compact/{topic}` | Trigger log compaction for topic |
| `GET` | `/api/v1/transactions/{id}` | Query atomic transaction status |
| `/metrics` | `GET` | Prometheus metrics (per-topic lag, throughput, error rates) |

---

## Protocols Supported

| Protocol | Transport | Notes |
|----------|-----------|-------|
| STOMP 1.2 | TCP / WebSocket | Primary protocol |
| Kafka Binary | TCP | Wire-compatible; use existing Kafka clients |
| MQTT v5.0 | TCP / WebSocket | IoT device support, QoS 0/1/2 |
| OPFS (browser) | WASM | Local-first browser queue |
| WebTransport | HTTP/3 QUIC | Browser outbox relay |

---

## Browser / OPFS Features

Install the browser SDK:

```bash
npm install @servverse/queue-wasm
```

```typescript
import { ServQueue } from '@servverse/queue-wasm';

const queue = new ServQueue({ encryption: 'aes-256-gcm' });
await queue.publish('orders', { id: 1, item: 'Widget' });
await queue.subscribe('orders', (msg) => console.log(msg));

// Auto-syncs to server when online; stores locally when offline
await queue.enableOfflineSync({ serverUrl: 'wss://queue.servverse.net' });
```

---

## Security

| Feature | Description |
|---------|-------------|
| FIPS 140-3 & HSM key unsealing | HSM-backed key management for regulated environments |
| Blind Broker E2EE | End-to-end encryption — broker never sees plaintext |
| Post-Quantum Hybrid Crypto (PQC) | X25519+Kyber hybrid key exchange |
| Tamper-Evident Merkle Audit Ledger | Append-only Merkle tree audit log for every message event |
| Inline WASM AI Guardrails | Sandboxed WASM interceptors on message payloads |
| Byzantine Fault Tolerant Consensus | BFT Raft variant for tamper-resistant cluster consensus |
| Zero-Trust OAuth2 & SPIFFE | Per-connection workload identity attestation |
| AES-256-GCM (OPFS) | Client-side encryption for browser-local messages |

---

## Observability

- **Prometheus `/metrics`**: Per-topic message rate, consumer lag, DLQ depth, compaction stats
- **OTel W3C Trace Context**: `traceparent` header propagated per message through full pipeline
- **ServConsole Queue Inspector**: Live topic browser, consumer group lag dashboard, DLQ browser with one-click replay, schema registry browser

---

## Kubernetes & Distribution

```bash
# Standalone daemon
servqueued --port 9090 --storage ./data --tls

# CLI
servqueue publish orders '{"id": 1}'
servqueue consume orders --group my-service
serv queue publish orders '{"id": 1}'   # Serv-lang integration

# Kubernetes Operator
kubectl apply -f servqueuecluster.yaml

# KEDA auto-scaling
kubectl apply -f keda-scaledobject.yaml   # Scale consumers on lag
```

Multi-language client SDKs: Go, TypeScript/JS, Python, Rust, Java.

Cross-cloud active-active geo-replication with automated failover and conflict resolution.

---

## Getting Started

```bash
docker run -p 9090:9090 \
  -e SERVQUEUE_STORAGE_PATH=/data \
  -e SERVQUEUE_OTEL_ENDPOINT=http://servtrace:4318 \
  -v queue-data:/data \
  ghcr.io/vyuvaraj/servqueue:latest
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVQUEUE_PORT` | `9090` | Listener port |
| `SERVQUEUE_STORAGE_PATH` | `./data` | WAL and segment storage directory |
| `SERVQUEUE_OTEL_ENDPOINT` | — | OpenTelemetry collector URL |
| `SERVQUEUE_S3_BUCKET` | — | S3 bucket for tiered offloading |
| `SERVQUEUE_KAFKA_COMPAT` | `false` | Enable Kafka wire protocol adapter |
| `SERVQUEUE_MQTT_PORT` | — | MQTT listener port |
| `SERVQUEUE_FIPS` | `false` | Enable FIPS 140-3 mode (EE) |

---

## Enterprise Edition

| Feature | Tier |
|---------|------|
| Geo-Replication across clouds | EE |
| Kafka Protocol Adapter | EE |
| FIPS 140-3 HSM & Sovereign Security | EE |
| Inline WASM AI Guardrails | EE |
| eBPF Kernel Bypass & XDP Acceleration | EE |
| Multi-Cloud Tiered Storage Compaction | EE |
| AWS EventBridge & Enterprise Webhooks Connector | EE |
| Multi-Cluster Kubernetes Federation | EE |
| SIMD/AVX-512 Vectorized Filter Engine | EE |
| Byzantine Fault Tolerant Consensus | EE |
| Post-Quantum Hybrid Cryptography | EE |
