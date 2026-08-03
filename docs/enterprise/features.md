# Pranor Enterprise Edition (EE)

Pranor EE extends the open-source single-binary platform with advanced security, compliance, multi-region high availability, and operational governance for enterprise engineering teams.

---

## Feature Comparison Matrix

| Capability / Feature Area | Community OSS | Enterprise EE |
| :--- | :---: | :---: |
| **API Gateway (Gate)** | Ingress proxy, WASM sandbox, Rate limiting, ACME mTLS | **+ Kernel eBPF XDP DDoS bypass, AI Cost Router, Geo-IP steering, GraphQL Federation** |
| **Object & Vector Storage (Vault)** | AWS S3 API, HNSW vector search, RAG queries, zstd compression | **+ Active-Active multi-region sync, Copy-on-Write bucket branching, Sovereign envelope encryption** |
| **Event Broker & Queue (Pulse)** | Kafka, STOMP, MQTT decoders, DLQ replay, WASM transforms | **+ Multi-region MirrorMaker, Hardware payload encryption at rest, Schema Registry enforcement** |
| **Stateful Workflows (Flow)** | Saga orchestrator, Reverse compensation, Crash checkpointing | **+ Visual drag-and-drop builder, Multi-region coordinator cluster, SLA auto-escalation** |
| **Observability (Trace & Console)** | OTLP trace collector, Flamegraph viewer, SQL workbench | **+ Anomaly-based auto-remediation runbooks, SOC2/HIPAA WORM audit exporter, SAML/OIDC SSO** |
| **Identity & Secrets (Auth & Secret)** | JWKS, TOTP MFA, OAuth, KMS envelope key rotation | **+ Biometric Passkey/WebAuthn server, AD/LDAP Directory sync, Hardware HSM key signing** |
| **Scheduler & PaaS (Chrono & Deploy)** | Cron scheduler, Process runner, Gateway auto-sync | **+ Smart SLA-aware cron window optimizer, Multi-cluster fleet DR failover** |
| **Zero-Trust Mesh (Mesh)** | UDP auto-discovery, In-memory gRPC, mTLS rotation | **+ Hardware TPM/HSM attestation, Multi-cloud cross-VPC mesh peering** |
| **Enterprise Compliance & Support** | AGPL-3.0 Open Source | **+ FIPS 140-3 cryptography, SOC2 Type II audit evidence, 24/7 SLA Support** |

---

## Key Enterprise Capabilities

### 🔒 1. Advanced Security & Zero-Trust
- **Hardware TPM / HSM Attestation**: Hardware-backed cryptographic identity verification across zero-trust enterprise clusters.
- **Biometric Passkey & WebAuthn Server**: FIDO2 / WebAuthn passwordless authentication server eliminating password storage.
- **Active Directory & LDAP Directory Sync**: Direct enterprise directory synchronization with automatic group and role mapping.
- **Sovereign Client Envelope Encryption**: Zero-knowledge envelope encryption ensuring storage nodes never hold raw decryption keys.

### ⚡ 2. Scale & Multi-Region High Availability
- **Kernel eBPF XDP 100Gbps Acceleration**: Linux kernel-bypass packet filtering eliminating OS socket overhead for Pranor Gate.
- **Active-Active Multi-Region Replication**: Cross-cloud object statement replication with Last-Write-Wins (LWW) conflict resolution for Pranor Vault.
- **Multi-Region MirrorMaker Event Sync**: Active-active cross-cloud event topic mirroring with automatic deduplication across AWS, GCP, and Azure for Pranor Pulse.
- **Multi-Cluster Fleet DR Failover**: 1-click active-passive region failover, automatic DNS update sync, and cross-cloud deployment status for Pranor Deploy.

### 🛡️ 3. Governance, Compliance & Automation
- **SOC2 Type II / HIPAA WORM Audit Exporter**: Immutable Write-Once-Read-Many audit logging for all control plane actions, API calls, and queries with automated report generation.
- **Anomaly-Based Auto-Remediation Runbooks**: Automatically executes remediation webhooks or container restarts when trace anomaly models detect SLA breaches.
- **AI Cost Router & Prompt Guard**: Semantic prompt inspection, PII redaction, embedding similarity injection detection, and LLM failover routing.
- **Compliance Schema Registry**: Strict Schema Registry enforcement blocking producers from publishing backward-incompatible event payloads.

---

## Commercial Licensing & Build Tags

Enterprise features are compiled cleanly behind `//go:build enterprise` build tags. They compile into the standard single binary without external sidecars:

```bash
# Compile binary with enterprise feature tags
go build -tags enterprise -o pranord ./cmd/pranord
```

For commercial licensing inquiries, pilot programs, or dedicated enterprise support contracts:
- **Enterprise Repository**: [`github.com/vyuvaraj/pranor-ee`](https://github.com/vyuvaraj/pranor-ee) (Private)
- **Interactive Feature Matrix**: [`https://vyuvaraj.github.io/pranor-platform/features.html`](https://vyuvaraj.github.io/pranor-platform/features.html)

---

## Related Documentation
- [Security Architecture](../architecture/security.md)
- [Kubernetes Deployment](../deployment/kubernetes.md)
- [Ecosystem Integrations](../integrations.md)
