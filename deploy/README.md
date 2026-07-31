# Pranor Deploy

Deployment Orchestrator with Docker, Kubernetes, blue/green, and canary deployments.

## Quick Start

```bash
# Run standalone
docker run -p 8080:8080 ghcr.io/vyuvaraj/pranor-deploy:latest

# Or build from source
cd deploy && go build -o pranor-deploy .
```

## Documentation

**Full documentation:** [docs/modules/deploy.md](../docs/modules/deploy.md)

| Resource | Link |
|----------|------|
| Full docs | [docs/modules/deploy.md](../docs/modules/deploy.md) |
| API reference | [docs/modules/deploy.md#api-endpoints](../docs/modules/deploy.md#api-endpoints) |
| Deployment | [docs/deployment/docker.md](../docs/deployment/docker.md) |
| Architecture | [docs/architecture/overview.md](../docs/architecture/overview.md) |

## License

AGPL-3.0 — See [LICENSE](./LICENSE)