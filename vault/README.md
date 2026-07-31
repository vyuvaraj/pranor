# Pranor Vault

S3-compatible Object Storage with embedded vector search, time-travel versioning, and erasure coding.

## Quick Start

```bash
# Run standalone
docker run -p 8080:8080 ghcr.io/vyuvaraj/pranor-vault:latest

# Or build from source
cd vault && go build -o pranor-vault .
```

## Documentation

**Full documentation:** [docs/modules/vault.md](../docs/modules/vault.md)

| Resource | Link |
|----------|------|
| Full docs | [docs/modules/vault.md](../docs/modules/vault.md) |
| API reference | [docs/modules/vault.md#api-endpoints](../docs/modules/vault.md#api-endpoints) |
| Deployment | [docs/deployment/docker.md](../docs/deployment/docker.md) |
| Architecture | [docs/architecture/overview.md](../docs/architecture/overview.md) |

## License

AGPL-3.0 — See [LICENSE](./LICENSE)