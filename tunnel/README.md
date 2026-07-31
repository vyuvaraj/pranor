# Pranor Tunnel

WebSocket Dev Tunneling with custom subdomains, HTTP inspector, and auth gating.

## Quick Start

```bash
# Run standalone
docker run -p 8080:8080 ghcr.io/vyuvaraj/pranor-tunnel:latest

# Or build from source
cd tunnel && go build -o pranor-tunnel .
```

## Documentation

**Full documentation:** [docs/modules/tunnel.md](../docs/modules/tunnel.md)

| Resource | Link |
|----------|------|
| Full docs | [docs/modules/tunnel.md](../docs/modules/tunnel.md) |
| API reference | [docs/modules/tunnel.md#api-endpoints](../docs/modules/tunnel.md#api-endpoints) |
| Deployment | [docs/deployment/docker.md](../docs/deployment/docker.md) |
| Architecture | [docs/architecture/overview.md](../docs/architecture/overview.md) |

## License

AGPL-3.0 — See [LICENSE](./LICENSE)