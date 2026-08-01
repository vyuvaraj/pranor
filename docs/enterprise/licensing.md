# Enterprise Licensing & Edition Split

Pranor is distributed under a dual-licensing model designed for both open-source developers and enterprise organizations.

---

## Edition Matrix

| Capability | Open-Source Edition (OSS) | Enterprise Edition (EE) |
|------------|---------------------------|-------------------------|
| **Core Monorepo Modules** | 16 Modules Included | 16 Modules Included |
| **License** | Apache 2.0 / MIT | Commercial Enterprise License |
| **High Availability & Clustering** | Standalone & Basic Mesh | Active-Active Multi-Region, Raft Consensus |
| **Security & Compliance** | TLS 1.3, JWT, RBAC | FIPS 140-3, HSM Key Unsealing, PQC (Kyber) |
| **Observability** | OTel Tracing, Metrics | eBPF Continuous Profiling, Flamegraphs |
| **Support** | Community (GitHub / Discord) | 24/7 SLA, Dedicated Solutions Engineer |

---

## Licensing Terms

### Open-Source Edition (OSS)
The open-source components in `github.com/vyuvaraj/pranor` are available for free use, modification, and self-hosted deployment under standard open-source licenses.

### Enterprise Edition (EE)
Enterprise overlay modules in `github.com/vyuvaraj/pranor-ee` require a commercial license key issued by Pranor Inc. Features are gated at compile time using Go build tags (`//go:build enterprise`).
