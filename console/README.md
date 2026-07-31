# Pranor Console

Observability Dashboard with metrics visualization, SQL workbench, and incident management.

## Quick Start

```bash
# Run standalone
docker run -p 8080:8080 ghcr.io/vyuvaraj/pranor-console:latest

# Or build from source
cd console && go build -o pranor-console .
```

## Documentation

**Full documentation:** [docs/modules/console.md](../docs/modules/console.md)

| Resource | Link |
|----------|------|
| Full docs | [docs/modules/console.md](../docs/modules/console.md) |
| API reference | [docs/modules/console.md#api-endpoints](../docs/modules/console.md#api-endpoints) |
| Deployment | [docs/deployment/docker.md](../docs/deployment/docker.md) |
| Architecture | [docs/architecture/overview.md](../docs/architecture/overview.md) |

## License

AGPL-3.0 — See [LICENSE](./LICENSE)