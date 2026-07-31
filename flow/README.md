# Pranor Flow

Workflow Engine with DAG execution, saga orchestrator, and human approval gates.

## Quick Start

```bash
# Run standalone
docker run -p 8080:8080 ghcr.io/vyuvaraj/pranor-flow:latest

# Or build from source
cd flow && go build -o pranor-flow .
```

## Documentation

**Full documentation:** [docs/modules/flow.md](../docs/modules/flow.md)

| Resource | Link |
|----------|------|
| Full docs | [docs/modules/flow.md](../docs/modules/flow.md) |
| API reference | [docs/modules/flow.md#api-endpoints](../docs/modules/flow.md#api-endpoints) |
| Deployment | [docs/deployment/docker.md](../docs/deployment/docker.md) |
| Architecture | [docs/architecture/overview.md](../docs/architecture/overview.md) |

## License

AGPL-3.0 — See [LICENSE](./LICENSE)