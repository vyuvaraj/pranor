# Pranor Notify

Notification Gateway for transactional email, Slack, and SMS with HTML templating.

## Quick Start

```bash
# Run standalone
docker run -p 8080:8080 ghcr.io/vyuvaraj/pranor-notify:latest

# Or build from source
cd notify && go build -o pranor-notify .
```

## Documentation

**Full documentation:** [docs/modules/notify.md](../docs/modules/notify.md)

| Resource | Link |
|----------|------|
| Full docs | [docs/modules/notify.md](../docs/modules/notify.md) |
| API reference | [docs/modules/notify.md#api-endpoints](../docs/modules/notify.md#api-endpoints) |
| Deployment | [docs/deployment/docker.md](../docs/deployment/docker.md) |
| Architecture | [docs/architecture/overview.md](../docs/architecture/overview.md) |

## License

AGPL-3.0 — See [LICENSE](./LICENSE)