# Pranor Unified Changelog

The Pranor platform and background microservice ecosystem undergo continuous evolution across language tooling, gateways, brokers, storage engines, and observability collectors.

---

## [v1.0.0] - Production Release

### Language & Tooling (Pranor CLI, LSP & IDE Extension)
- **Unified Single-Binary Daemon**: Embedded all 17 background microservices into a single `pranord` executable.
- **Language Server Protocol (LSP)**: Advanced handlers for workspace-wide fuzzy symbol search (`workspace/symbol`), call hierarchy inspection (`textDocument/prepareCallHierarchy`), multi-file symbol renames (`textDocument/rename`), and document highlighting (`textDocument/documentHighlight`).
- **VS Code Control Plane (`pranor-vscode`)**: Registered interactive webview control panels for Gate (API Client), Pulse (Event Stream Tailer), Vault (Vector Explorer), Trace (Flamegraph Viewer), Secret Manager, and Cluster Deployments.
- **Developer Experience**: Interactive CLI setup wizard (`pranor quickstart`), diagnostic verification tool (`pranor doctor`), and devcontainer integration.

### Core Ecosystem Modules

#### Pranor Gate (API Gateway & Ingress Router)
- Edge HTTP/gRPC ingress routing with dynamic mTLS certificate rotation.
- Token bucket rate limiting per IP / API key.
- Sandboxed WebAssembly (Wazero) middleware execution.
- Backpressure load balancing and dynamic IAM token refresh signaling.

#### Pranor Pulse (Async Event Broker & Message Queue)
- Multi-protocol message broker supporting Kafka wire format, STOMP WebSockets, and MQTT 3.1/5.0.
- Automatic Dead Letter Queue (DLQ) isolation with 1-click message replay.

#### Pranor Vault (S3 Storage & HNSW Vector Engine)
- AWS S3 API compatibility (multipart uploads, presigned URLs, bucket policies).
- Native HNSW vector similarity search (Cosine, Euclidean, Dot-product).
- Offline S3 mock mode (`--mock` / `PRANOR_VAULT_MOCK=true`).

#### Pranor Flow (Workflow Engine & Durable Sagas)
- Durable saga orchestrator with HTTP/STOMP compensation rollbacks.
- Asynchronous STOMP compensation notifications over Pulse topics.

#### Pranor Auth, Cache, Mesh & Trace
- **Auth**: Identity server with JWKS, TOTP MFA, Social OAuth, and KMS envelope key rotation.
- **Cache**: Raft-based key-value caching with adaptive connection pool tuning.
- **Mesh**: Zero-trust in-memory service discovery with gRPC JSON-codec transport.
- **Trace**: OTLP/HTTP distributed tracing collector, flamegraph viewer, and Prometheus remote write ingestion.

### External Connectors & Integrations
- **Terraform Provider (`terraform-provider-pranor`)**: Declarative management of buckets, topics, cron jobs, and gateway routes.
- **GitHub Action (`pranor/deploy-action@v1`)**: Automated CI/CD compilation and blue/green deployments.
- **Grafana Datasource Plugin**: Direct visualization of Pranor Trace spans and Pulse queue metrics in Grafana.
