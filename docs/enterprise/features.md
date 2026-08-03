# Pranor Enterprise Edition (EE)

Pranor EE extends the open-source single-binary platform with 32 advanced security, compliance, multi-region high availability, and operational governance capabilities for enterprise engineering teams.

---

## Detailed Enterprise Feature Comparison (By Module)

### 1. 🚪 Pranor Gate (API Gateway & Ingress Router)
| Feature | Community OSS | Enterprise EE |
| :--- | :---: | :---: |
| **Ingress Proxy** | HTTP/1.1, HTTP/2, gRPC | ✅ Sub-millisecond edge proxy with header matching & path rewriting |
| **Kernel eBPF XDP DDoS Bypass** | ❌ | ✅ **100Gbps network packet filtering at Linux kernel level** |
| **AI Cost Router & Prompt Firewall** | ❌ | ✅ **Semantic prompt inspection, PII redaction & LLM failover** |
| **Geo-IP Latency Anycast Steering** | ❌ | ✅ **Real-time edge routing to lowest-latency datacenter** |
| **GraphQL Schema Federation** | ❌ | ✅ **Stitches multiple backend GraphQL subgraphs into single schema** |
| **Enterprise WAF Rule Generator** | ❌ | ✅ **Auto-generates OWASP Top-10 protection rules** |
| **TLS Hardware Offloading** | ❌ | ✅ **Offloads TLS handshakes to PCIe crypto accelerator cards** |

### 2. 🗄️ Pranor Vault (S3 Storage & Vector Search)
| Feature | Community OSS | Enterprise EE |
| :--- | :---: | :---: |
| **AWS S3 API & HNSW Vector Search** | ✅ | ✅ AWS S3 SDK compatibility & HNSW vector search engine |
| **Immutable Access Audit Trail** | ❌ | ✅ **Append-only log recording object reads/writes with identity** |
| **Active-Active Multi-Region Sync** | ❌ | ✅ **Cross-cloud object replication with LWW conflict resolution** |
| **CoW Bucket Branching** | ❌ | ✅ **Instant copy-on-write bucket branching for sandbox dev** |
| **Sovereign Client Envelope Encryption** | ❌ | ✅ **Zero-knowledge client encryption (storage nodes hold no keys)** |
| **Automatic Cold-Tier Archival** | ❌ | ✅ **Lifecycle migration to AWS Glacier, Azure Cool & GCS Nearline** |
| **WORM Retention Lock Manager** | ❌ | ✅ **SEC 17a-4 / HIPAA immutable retention windows & legal holds** |
| **Embedded Analytics & Torrent** | ❌ | ✅ **DuckDB SQL over S3 parquet objects & P2P WebTorrent seeding** |

### 3. ⚡ Pranor Pulse (Async Event Broker & Queue)
| Feature | Community OSS | Enterprise EE |
| :--- | :---: | :---: |
| **Multi-Protocol Engine** | Kafka, STOMP, MQTT 3.1/5.0 | ✅ Unified multi-protocol message decoders & WASM transforms |
| **Multi-Tenant Partition Sharding** | ❌ | ✅ **Isolated queue memory pools & dedicated partition routing** |
| **Multi-Region MirrorMaker Sync** | ❌ | ✅ **Active-active cross-cloud topic mirroring (AWS, GCP, Azure)** |
| **Hardware Payload Encryption at Rest** | ❌ | ✅ **KMS/HSM envelope encryption before writing to disk WAL** |
| **Schema Breaking-Change Guard** | ❌ | ✅ **BACKWARD/FULL compatibility blocking breaking event payloads** |
| **Rebalance Auto-Tuning** | ❌ | ✅ **AI-assisted partition rebalancing eliminating consumer starvation** |

### 4. 🔀 Pranor Flow (Stateful Workflows & Sagas)
| Feature | Community OSS | Enterprise EE |
| :--- | :---: | :---: |
| **Durable Saga Orchestrator** | ✅ | ✅ Stateful transaction coordinator with compensation rollbacks |
| **Visual Workflow Builder & Replay** | ❌ | ✅ **Drag-and-drop designer with step-by-step state diff playback** |
| **HA Coordinator Cluster** | ❌ | ✅ **Multi-region workflow state replication across Raft clusters** |
| **SLA Deadline Enforcer** | ❌ | ✅ **Monitors instance SLAs & fires escalation webhooks (PagerDuty/Slack)** |

### 5. 📡 Pranor Trace & Console (Observability & Control)
| Feature | Community OSS | Enterprise EE |
| :--- | :---: | :---: |
| **OTLP Tracing & SQL Workbench** | ✅ | ✅ Full OTel span collector, flamegraphs & SQL workbench |
| **Anomaly Auto-Remediation Runbooks** | ❌ | ✅ **Fires remediation webhooks or container restarts on trace anomaly** |
| **Long-Term Tail Sampling Storage** | ❌ | ✅ **Archives 100% error traces + 0.1% normal spans to cold S3** |
| **SOC2 / HIPAA Immutable Audit** | ❌ | ✅ **WORM audit logging with SHA-256 tamper-proof proof generation** |
| **SAML 2.0 / OIDC Enterprise SSO & RBAC** | ❌ | ✅ **Enterprise SSO (Okta, Azure AD) with fine-grained RBAC** |

### 6. 🔐 Pranor Auth & Secret (Security & Keys)
| Feature | Community OSS | Enterprise EE |
| :--- | :---: | :---: |
| **Identity & Access Management** | JWKS, MFA, OAuth | ✅ IAM server with adaptive risk-based authentication |
| **SCIM 2.0 User Provisioning** | ❌ | ✅ **User/group sync from Okta, Azure AD & Google Workspace** |
| **Credential Stuffing Detector** | ❌ | ✅ **Distributed velocity tracking blocking credential stuffing** |
| **Biometric Passkey / WebAuthn** | ❌ | ✅ **FIDO2 / Passkey passwordless authentication server** |
| **VS Code Visual Secret Console** | ❌ | ✅ **Unseal vault stores & manage secret maps inside VS Code** |

### 7. ⏰ Pranor Chrono, Deploy & Mesh (Orchestration & Network)
| Feature | Community OSS | Enterprise EE |
| :--- | :---: | :---: |
| **Smart SLA Cron Optimizer** | ❌ | ✅ **Predictive cron optimizer shifting jobs to off-peak cost windows** |
| **Multi-Cluster Fleet DR Failover** | ❌ | ✅ **1-click active-passive region failover with DNS update sync** |
| **Multi-Cloud Cross-VPC Mesh Peering**| ❌ | ✅ **Encrypted WireGuard mesh tunnels across AWS EKS, GCP & on-prem** |

---

## Commercial Licensing & Build Tags

Enterprise features compile cleanly behind `//go:build enterprise` build tags into the standard single binary:

```bash
# Compile binary with all 32 Enterprise Edition capabilities enabled
go build -tags enterprise -o pranord ./cmd/pranord
```

For commercial licensing inquiries, pilot programs, or dedicated support contracts:
- **Enterprise Repository**: [`github.com/vyuvaraj/pranor-ee`](https://github.com/vyuvaraj/pranor-ee) (Private)
- **Interactive Feature Matrix**: [`https://vyuvaraj.github.io/pranor-platform/features.html`](https://vyuvaraj.github.io/pranor-platform/features.html)

---

## Related Documentation
- [Security Architecture](../architecture/security.md)
- [Kubernetes Deployment](../deployment/kubernetes.md)
- [Ecosystem Integrations](../integrations.md)
