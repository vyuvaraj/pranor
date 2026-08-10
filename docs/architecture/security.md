# Security Architecture

Pranor implements defense-in-depth across all modules.

## Authentication & Authorization

| Layer | Mechanism | Module |
|-------|-----------|--------|
| API clients | JWT (RS256/ES256) + OAuth2/OIDC | Auth |
| Inter-service | mTLS with auto-rotating certificates | Mesh |
| Admin APIs | API key + RBAC | All modules |
| Browser sessions | Secure cookies + MFA (TOTP/WebAuthn) | Auth |

## Zero-Trust Model

Every request between modules is authenticated:

```
Client → Gate (JWT validation & Agent Security Chain)
       → Agent Firewall (Intent, Risk & HITL Approval)
       → Mesh (mTLS between services)
       → Target Service / Tool (Capability execution)
```

No module trusts another implicitly. Mesh provides workload identity via SPIFFE, while Gate enforces Agent Security Chains (`Agent ID -> User ID -> Tenant ID -> Capability ID`).

## AI Agent Security & Governance

| Feature | Mechanism | Scope |
|---------|-----------|-------|
| AI Agent Security Firewall | Inspects tool call intents, arguments & risk scores (`ALLOW/DENY/APPROVE/TRANSFORM`) | Gate |
| Agent Security Chain | First-class `Agent ID -> User ID -> Tenant ID -> Capability` context propagation | Gate / Auth |
| Human-in-the-Loop (HITL) | Asynchronous approval workflows (`Agent -> Gate -> Approval -> Gate -> Tool`) | Gate |
| Trajectory Replay & Simulation | Replays recorded trajectory steps to simulate & diff policy changes | Gate |
| Agent Blast-Radius & Budgets | Session-level and action-specific tool call rate limits | Gate |
| Protocol-Agnostic Exposer | Exposes capabilities across MCP, gRPC, HTTP/REST, and WASM | Gate |

## Encryption

| Scope | Algorithm | Where |
|-------|-----------|-------|
| Data at rest | AES-256-GCM | Vault, Pulse |
| Data in transit | TLS 1.3 | All inter-module traffic |
| Secrets storage | AES-256-GCM + Shamir sharing | Secret |
| Browser queue | AES-256-GCM client-side | Pulse (OPFS) |
| JWT signing | RS256 or ES256 | Auth |

## Enterprise Security (EE)

| Feature | Description |
|---------|-------------|
| FIPS 140-3 mode | HSM-backed key management |
| Post-quantum crypto | X25519 + Kyber hybrid key exchange |
| Byzantine consensus | BFT Raft for tamper-resistant clusters |
| eBPF XDP acceleration | Kernel-level packet filtering |
| Blind broker E2EE | Pulse broker never sees plaintext messages |
| Merkle audit ledger | Tamper-evident append-only audit trail |

## RBAC Model

```pranor
// Define roles in Auth
POST /api/v1/rbac/roles
{
  "name": "editor",
  "permissions": ["read:articles", "write:articles"]
}

// Assign to users
POST /api/v1/rbac/users/user-123/roles
{ "roles": ["editor"] }
```

Gate enforces RBAC policies on every routed request.

## Secret Management

Pranor Secret provides:
- Dynamic secret injection into processes
- Shamir key splitting for master key unsealing
- Automatic rotation with versioning
- Leak detection scanning

```bash
pranor secret inject --env production -- ./my-service
```

## Next Steps

- [Observability](./observability.md)
- [Architecture Overview](./overview.md)
- [Auth Module Docs](../modules/auth.md)
