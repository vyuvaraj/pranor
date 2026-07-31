# Pranor Cache

Distributed Cache Engine with dual-mode memory/Redis, TTL policies, and bloom filters.

## Quick Start

```bash
# Run standalone
docker run -p 8080:8080 ghcr.io/vyuvaraj/pranor-cache:latest

# Or build from source
cd cache && go build -o pranor-cache .
```

## Documentation

**Full documentation:** [docs/modules/cache.md](../docs/modules/cache.md)

| Resource | Link |
|----------|------|
| Full docs | [docs/modules/cache.md](../docs/modules/cache.md) |
| API reference | [docs/modules/cache.md#api-endpoints](../docs/modules/cache.md#api-endpoints) |
| Deployment | [docs/deployment/docker.md](../docs/deployment/docker.md) |
| Architecture | [docs/architecture/overview.md](../docs/architecture/overview.md) |

## License

AGPL-3.0 — See [LICENSE](./LICENSE)