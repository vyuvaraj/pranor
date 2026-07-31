# Pranor Chrono

Distributed Job Scheduler with multi-node leader election, cron expressions, and persistent state.

## Quick Start

```bash
# Run standalone
docker run -p 8080:8080 ghcr.io/vyuvaraj/pranor-chrono:latest

# Or build from source
cd chrono && go build -o pranor-chrono .
```

## Documentation

**Full documentation:** [docs/modules/chrono.md](../docs/modules/chrono.md)

| Resource | Link |
|----------|------|
| Full docs | [docs/modules/chrono.md](../docs/modules/chrono.md) |
| API reference | [docs/modules/chrono.md#api-endpoints](../docs/modules/chrono.md#api-endpoints) |
| Deployment | [docs/deployment/docker.md](../docs/deployment/docker.md) |
| Architecture | [docs/architecture/overview.md](../docs/architecture/overview.md) |

## License

AGPL-3.0 — See [LICENSE](./LICENSE)