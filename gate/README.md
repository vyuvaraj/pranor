# Pranor Gate

API Gateway & Ingress Router with AI Guard, WASM plugins, rate limiting, and circuit breaking.

## Quick Start

```bash
# Run standalone
docker run -p 8080:8080 ghcr.io/vyuvaraj/pranor-gate:latest

# Or build from source
cd gate && go build -o pranor-gate .
```

## Documentation

**Full documentation:** [docs/modules/gate.md](../docs/modules/gate.md)

| Resource | Link |
|----------|------|
| Full docs | [docs/modules/gate.md](../docs/modules/gate.md) |
| API reference | [docs/modules/gate.md#api-endpoints](../docs/modules/gate.md#api-endpoints) |
| Deployment | [docs/deployment/docker.md](../docs/deployment/docker.md) |
| Architecture | [docs/architecture/overview.md](../docs/architecture/overview.md) |

## License

AGPL-3.0 — See [LICENSE](./LICENSE)