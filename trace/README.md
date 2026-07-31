# Pranor Trace

Distributed Tracing Engine with OTLP collection, waterfall UI, and anomaly detection.

## Quick Start

```bash
# Run standalone
docker run -p 8080:8080 ghcr.io/vyuvaraj/pranor-trace:latest

# Or build from source
cd trace && go build -o pranor-trace .
```

## Documentation

**Full documentation:** [docs/modules/trace.md](../docs/modules/trace.md)

| Resource | Link |
|----------|------|
| Full docs | [docs/modules/trace.md](../docs/modules/trace.md) |
| API reference | [docs/modules/trace.md#api-endpoints](../docs/modules/trace.md#api-endpoints) |
| Deployment | [docs/deployment/docker.md](../docs/deployment/docker.md) |
| Architecture | [docs/architecture/overview.md](../docs/architecture/overview.md) |

## License

AGPL-3.0 — See [LICENSE](./LICENSE)