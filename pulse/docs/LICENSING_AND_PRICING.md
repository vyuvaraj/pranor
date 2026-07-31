# Pranor Pulse Licensing & Commercial Pricing Strategy

This document outlines the official licensing strategy, dual-licensing policy, client SDK permissions, and commercial tiering model for **Pranor Pulse**.

For the full detailed document, see [LICENSING_AND_PRICING.md](../../../pranor-repo/docs/LICENSING_AND_PRICING.md) in the ServVerse ecosystem repository.

---

## Executive Summary & Licensing Recommendation

Pranor Pulse uses a **Dual-Licensing Open-Core Model**:

1. **Pranor Pulse Server Engine (`servqueued`)**: **GNU AGPLv3** (Open Source)
   - Protects core intellectual property against cloud hyperscalers stealing work.
   - Requires any network operator modifying `servqueued` for a SaaS offering to open-source their network modifications, or purchase a **Pranor Pulse Enterprise Commercial License**.
2. **Client SDKs & OPFS Browser WASM Package (`sdks/go`, `@pranor/queue-wasm`)**: **Apache 2.0 / MIT**
   - 100% permissive and frictionless integration for all frontend web applications, microservices, and mobile clients without copyleft restrictions.
3. **Pranor Pulse Enterprise Commercial Engine (`serv-ee`)**: **Commercial Proprietary License**
   - Unlocks commercial enterprise features (Geo-replication, Kafka wire compatibility, FIPS 140-3 HSM, eBPF XDP, AI guardrails, K8s federation).

---

## Commercial Tiers & Pricing Overview

- **Community Tier**: **$0** (Free, AGPLv3 Open Source, Unlimited Cores).
- **Enterprise Tier**: **$30 / vCPU Core / Month** (billed annually at **$360 / core / year**).
- **Sovereign & Defense Tier**: **$60 / vCPU Core / Month** (billed annually at **$720 / core / year**).
- **Pranor Pulse Managed Cloud**: Pay-as-you-go ($0.04/GB transferred, $0.025/GB/mo hot buffer).
