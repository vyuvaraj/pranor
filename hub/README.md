# Pranor Hub

Package Registry with semver resolution, artifact signing, and OCI backend support.

## Quick Start

```bash
# Run standalone
docker run -p 8080:8080 ghcr.io/vyuvaraj/pranor-hub:latest

# Or build from source
cd hub && go build -o pranor-hub .
```

## Documentation

**Full documentation:** [docs/modules/hub.md](../docs/modules/hub.md)

| Resource | Link |
|----------|------|
| Full docs | [docs/modules/hub.md](../docs/modules/hub.md) |
| API reference | [docs/modules/hub.md#api-endpoints](../docs/modules/hub.md#api-endpoints) |
| Deployment | [docs/deployment/docker.md](../docs/deployment/docker.md) |
| Architecture | [docs/architecture/overview.md](../docs/architecture/overview.md) |

## License

AGPL-3.0 — See [LICENSE](./LICENSE)