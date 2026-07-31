# Pranor Enterprise Edition

Pranor EE extends the open-source platform with features for regulated, high-scale, and multi-tenant environments.

## Feature Comparison

| Feature | OSS | Enterprise |
|---------|:---:|:----------:|
| API Gateway (Gate) | ✅ | ✅ + WAF, GraphQL federation, eBPF XDP |
| Message Broker (Pulse) | ✅ | ✅ + Geo-replication, Kafka wire, BFT consensus |
| Object Storage (Vault) | ✅ | ✅ + Multi-cloud tiering, WORM compliance |
| Auth (OAuth2/OIDC/RBAC) | ✅ | ✅ + SAML, credential stuffing detection |
| Distributed Tracing (Trace) | ✅ | ✅ + NL query, cold-tier archival |
| Service Mesh | ✅ | ✅ + WireGuard overlay, adaptive LB |
| Workflow Engine (Flow) | ✅ | ✅ + Saga orchestrator, ML cost predictor |
| FIPS 140-3 / HSM | ❌ | ✅ |
| Post-Quantum Cryptography | ❌ | ✅ |
| eBPF Kernel Bypass | ❌ | ✅ |
| Multi-tenant isolation | Basic | Full namespace + quota |
| SLA: 99.99% uptime | ❌ | ✅ |
| Priority support | Community | 24/7 dedicated |

## Key Enterprise Capabilities

### Security
- **FIPS 140-3 mode** — HSM-backed key management for regulated industries
- **Post-quantum hybrid crypto** — X25519 + Kyber key exchange
- **Blind broker E2EE** — Pulse broker never sees message plaintext
- **Byzantine Fault Tolerant consensus** — Tamper-resistant Raft clustering
- **Merkle audit ledger** — Cryptographic proof of every operation

### Scale
- **Geo-replication** — Active-active multi-region for Vault and Pulse
- **eBPF XDP acceleration** — Kernel-bypass packet processing for Gate
- **SIMD/AVX-512 filters** — Vectorized message filtering in Pulse
- **Multi-cloud tiering** — Automatic hot/warm/cold storage lifecycle

### Compliance
- **SOC 2 Type II** evidence generation
- **GDPR** data residency controls
- **WORM storage** — Write-once-read-many for regulatory archives
- **Audit trails** — Every operation logged with tamper-proof integrity

## Licensing

Enterprise features are gated behind `//go:build enterprise` build tags. They compile into the same binary — no separate installation needed.

```bash
# Build with enterprise features
go build -tags enterprise -o pranor-gate .
```

## Contact

Enterprise Repo: [`github.com/vyuvaraj/pranor-ee`](https://github.com/vyuvaraj/pranor-ee) (Private)

## Next Steps

- [Security Architecture](../architecture/security.md)
- [Deployment Guide](../deployment/kubernetes.md)
- [Module Docs](../modules/)
