# Pranor Lock

Distributed Locking with lease-based acquisition, renewal, and Raft consensus.

## Quick Start

```bash
# Run standalone
docker run -p 8080:8080 ghcr.io/vyuvaraj/pranor-lock:latest

# Or build from source
cd lock && go build -o pranor-lock .
```

## Documentation

**Full documentation:** [docs/modules/lock.md](../docs/modules/lock.md)

| Resource | Link |
|----------|------|
| Full docs | [docs/modules/lock.md](../docs/modules/lock.md) |
| API reference | [docs/modules/lock.md#api-endpoints](../docs/modules/lock.md#api-endpoints) |
| Deployment | [docs/deployment/docker.md](../docs/deployment/docker.md) |
| Architecture | [docs/architecture/overview.md](../docs/architecture/overview.md) |

## License

AGPL-3.0 — See [LICENSE](./LICENSE)