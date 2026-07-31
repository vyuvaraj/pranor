# Pranor Secret

Secret Management with dynamic injection, Shamir key splitting, and rotation.

## Quick Start

```bash
# Run standalone
docker run -p 8080:8080 ghcr.io/vyuvaraj/pranor-secret:latest

# Or build from source
cd secret && go build -o pranor-secret .
```

## Documentation

**Full documentation:** [docs/modules/secret.md](../docs/modules/secret.md)

| Resource | Link |
|----------|------|
| Full docs | [docs/modules/secret.md](../docs/modules/secret.md) |
| API reference | [docs/modules/secret.md#api-endpoints](../docs/modules/secret.md#api-endpoints) |
| Deployment | [docs/deployment/docker.md](../docs/deployment/docker.md) |
| Architecture | [docs/architecture/overview.md](../docs/architecture/overview.md) |

## License

AGPL-3.0 — See [LICENSE](./LICENSE)