# Pranor Enterprise Edition (EE)

Pranor EE extends the open-source single-binary platform with advanced security, compliance, multi-region high availability, and operational governance capabilities for enterprise engineering teams.

---

## Detailed Enterprise & Open-Source Feature Comparison (By Module)

### 1. 🚪 Pranor Gate (API Gateway & Ingress Router)
| Feature | Community OSS | Enterprise EE |
| :--- | :---: | :---: |
| **Ingress Proxy & WASM Middleware** | ✅ Sub-millisecond HTTP/gRPC proxy & WASM hot-swap | ✅ Zero-downtime TLS hardware PCIe offloading |
| **Kernel eBPF XDP DDoS Bypass** | ❌ | ✅ **100Gbps network packet filtering at Linux kernel level** |
| **AI Agent (MCP) Traffic & Prompt Guard** | ✅ MCP JSON-RPC routing & token cost headers | ✅ **Semantic prompt firewall, PII redaction & injection guard** |
| **Hardware Accelerator & Bandwidth Shaper** | ❌ | ✅ **Direct PCIe GPU/TPU offloader & noisy-neighbor bandwidth shaper** |
| **Geo-IP Anycast & GraphQL Federation** | ❌ | ✅ **Real-time edge Anycast steering & GraphQL schema stitching** |
| **CRDT Rate Limiting & DR Failover** | ❌ | ✅ **Global CRDT rate-limiting grid & 1-click active-passive DR** |

### 2. 🗄️ Pranor Vault (S3 Storage & Vector Search)
| Feature | Community OSS | Enterprise EE |
| :--- | :---: | :---: |
| **AWS S3 API & HNSW Vector Search** | ✅ Full S3 SDK compatibility & native HNSW vector search | ✅ Sovereign vector embedding index isolation |
| **Zero-Knowledge Search & GDPR Purge** | ❌ | ✅ **Encrypted homomorphic search & automated GDPR zeroization** |
| **Audit Trail & Geo-Replication** | ❌ | ✅ **Immutable access audit logs & active-active multi-region sync** |
| **CoW Branching & Masking / MPC** | ❌ | ✅ **Instant bucket branching, dynamic PII masking & MPC secret split** |
| **WORM Retention & Cold Tiering** | ❌ | ✅ **SEC 17a-4 retention lock manager & Glacier automated tiering** |

### 3. ⚡ Pranor Pulse (Async Event Broker & Queue)
| Feature | Community OSS | Enterprise EE |
| :--- | :---: | :---: |
| **Multi-Protocol Engine & DLQ Replay** | ✅ Kafka, STOMP, MQTT decoders & 1-click DLQ replay | ✅ Dedicated per-tenant partition memory pool sharding |
| **Exactly-Once 2PC Transaction Coordinator** | ❌ | ✅ **Two-Phase Commit transaction manager enforcing atomic publish** |
| **MirrorMaker v2 & Hardware WAL** | ❌ | ✅ **Cross-cloud event topic mirroring & zero-copy WAL encryption** |
| **Blind Broker Encryption & SIMD Filter** | ❌ | ✅ **End-to-end payload encryption & SIMD / AVX-512 event filter** |
| **Rebalance Tuning & Schema Guard** | ❌ | ✅ **AI consumer rebalance auto-tuning & breaking-change guard** |

### 4. 🔀 Pranor Flow & Deploy (Workflows & Fleet Orchestration)
| Feature | Community OSS | Enterprise EE |
| :--- | :---: | :---: |
| **Durable Saga Orchestrator** | ✅ Stateful transaction coordinator with compensation rollbacks | ✅ Visual workflow builder, step replay & Raft coordinator |
| **Automated DR Chaos Simulation Suite** | ❌ | ✅ **In-situ chaos engineering testing cross-cloud failover SLAs** |
| **AI FinOps & Blue/Green Promotion** | ❌ | ✅ **Cloud cost guardrails & zero-downtime blue/green cluster promotion** |

### 5. 📡 Pranor Trace & Console (Observability & Governance)
| Feature | Community OSS | Enterprise EE |
| :--- | :---: | :---: |
| **OTLP Tracing & SQL Workbench** | ✅ Full OTel span collector, flamegraphs & SQL workbench | ✅ Anomaly auto-remediation runbooks & tail trace sampling |
| **AI Anomaly Auto-Tuner & SIEM Streamer** | ❌ | ✅ **Self-learning anomaly baseline & SIEM audit log streamer** |
| **Regulatory WORM Log & Compliance** | ❌ | ✅ **SEC Rule 17a-4 WORM vault & real-time compliance inspector** |
| **Incident Postmortem & VIP Support** | ❌ | ✅ **Automated postmortem synthesizer & 15-min emergency SLA support** |

### 6. 🔐 Pranor Auth, Secret, Core & Hub (Security & Identity)
| Feature | Community OSS | Enterprise EE |
| :--- | :---: | :---: |
| **IAM & KMS Envelope Encryption** | ✅ JWKS rotation, MFA, OAuth & KMS key rotation worker | ✅ Hardware HSM offloading & Vault Transit integration |
| **Confidential Computing Enclave** | ❌ | ✅ **Hardware memory enclave (AMD SEV / Intel SGX) isolation** |
| **Multi-Cloud KMS Federation Sync** | ❌ | ✅ **Key synchronization across AWS KMS, Azure Key Vault & GCP KMS** |
| **Enterprise Identity & Passkey** | ❌ | ✅ **SCIM 2.0 provisioning, FIDO2/WebAuthn & IdP claim mapping** |
| **FIPS 140-3 & Post-Quantum SPIFFE** | ❌ | ✅ **FIPS 140-3 Level 3 engine, Kyber768 PQC & SPIFFE token exchange** |
| **Air-Gapped Private Artifact Registry** | ❌ | ✅ **Offline package registry & RSA-4096 license key verifier** |

### 7. ⏰ Pranor Cache, Pool, Chrono, Tunnel & Mesh (Infrastructure)
| Feature | Community OSS | Enterprise EE |
| :--- | :---: | :---: |
| **Raft KV Cache & Developer Tunnel** | ✅ Distributed in-memory cache & multiplexed local tunnel | ✅ Sub-millisecond SIMD vector cache & zero-trust private relay |
| **Zero-Downtime DB Schema Migration** | ❌ | ✅ **Online DDL schema migration proxy & replica failover coordinator** |
| **Smart Cron & Fencing Tokens** | ✅ Mono-lock distributed cron execution with fencing tokens | ✅ AI off-peak cron window optimizer & multi-region fencing |
| **Zero-Trust Mesh & Microsegmentation** | ✅ Library-level sidecarless mTLS mesh & auto-discovery | ✅ Hardware TPM attestation, cross-VPC peering & L7 microsegmentation |

---

## Complete Interactive Features Matrix & Licensing

For the full interactive table listing **all 137+ Community OSS and Enterprise EE capabilities** with search filtering:
- **Interactive Web Matrix**: [`https://vyuvaraj.github.io/pranor-platform/features.html`](https://vyuvaraj.github.io/pranor-platform/features.html)

Enterprise features compile cleanly behind `//go:build enterprise` build tags into the standard single binary:

```bash
# Compile single binary with all Enterprise Edition capabilities enabled
go build -tags enterprise -o pranord ./cmd/pranord
```

For commercial licensing inquiries, pilot programs, or dedicated support contracts:
- **Enterprise Repository**: [`github.com/vyuvaraj/pranor-ee`](https://github.com/vyuvaraj/pranor-ee) (Private)
- **Open-Source Repository**: [`github.com/vyuvaraj/pranor`](https://github.com/vyuvaraj/pranor)

---

## Related Documentation
- [Security Architecture](../architecture/security.md)
- [Kubernetes Deployment](../deployment/kubernetes.md)
- [Ecosystem Integrations](../integrations.md)
