# Pranor Mesh

Service Discovery & Load Balancing with mTLS, circuit breaking, and adaptive routing.

## Quick Start

```bash
# Run standalone
docker run -p 8080:8080 ghcr.io/vyuvaraj/pranor-mesh:latest

# Or build from source
cd mesh && go build -o pranor-mesh .
```

## Documentation

**Full documentation:** [docs/modules/mesh.md](../docs/modules/mesh.md)

| Resource | Link |
|----------|------|
| Full docs | [docs/modules/mesh.md](../docs/modules/mesh.md) |
| API reference | [docs/modules/mesh.md#api-endpoints](../docs/modules/mesh.md#api-endpoints) |
| Deployment | [docs/deployment/docker.md](../docs/deployment/docker.md) |
| Architecture | [docs/architecture/overview.md](../docs/architecture/overview.md) |

## License

AGPL-3.0 — See [LICENSE](./LICENSE)