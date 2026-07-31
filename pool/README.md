# Pranor Pool

Database Connection Proxy with query analytics, read/write splitting, and leak detection.

## Quick Start

```bash
# Run standalone
docker run -p 8080:8080 ghcr.io/vyuvaraj/pranor-pool:latest

# Or build from source
cd pool && go build -o pranor-pool .
```

## Documentation

**Full documentation:** [docs/modules/pool.md](../docs/modules/pool.md)

| Resource | Link |
|----------|------|
| Full docs | [docs/modules/pool.md](../docs/modules/pool.md) |
| API reference | [docs/modules/pool.md#api-endpoints](../docs/modules/pool.md#api-endpoints) |
| Deployment | [docs/deployment/docker.md](../docs/deployment/docker.md) |
| Architecture | [docs/architecture/overview.md](../docs/architecture/overview.md) |

## License

AGPL-3.0 — See [LICENSE](./LICENSE)