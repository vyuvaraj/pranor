# Pranor Auth

Identity & Access Control with OAuth2/OIDC, multi-tenant RBAC, MFA, and JWT validation.

## Quick Start

```bash
# Run standalone
docker run -p 8080:8080 ghcr.io/vyuvaraj/pranor-auth:latest

# Or build from source
cd auth && go build -o pranor-auth .
```

## Documentation

**Full documentation:** [docs/modules/auth.md](../docs/modules/auth.md)

| Resource | Link |
|----------|------|
| Full docs | [docs/modules/auth.md](../docs/modules/auth.md) |
| API reference | [docs/modules/auth.md#api-endpoints](../docs/modules/auth.md#api-endpoints) |
| Deployment | [docs/deployment/docker.md](../docs/deployment/docker.md) |
| Architecture | [docs/architecture/overview.md](../docs/architecture/overview.md) |

## License

AGPL-3.0 — See [LICENSE](./LICENSE)