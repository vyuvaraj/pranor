# Pranor Pulse

Async Event Broker & Message Queue with STOMP, Kafka wire protocol, MQTT v5, and browser OPFS support.

## Quick Start

```bash
# Run standalone
docker run -p 8080:8080 ghcr.io/vyuvaraj/pranor-pulse:latest

# Or build from source
cd pulse && go build -o pranor-pulse .
```

## Documentation

**Full documentation:** [docs/modules/pulse.md](../docs/modules/pulse.md)

| Resource | Link |
|----------|------|
| Full docs | [docs/modules/pulse.md](../docs/modules/pulse.md) |
| API reference | [docs/modules/pulse.md#api-endpoints](../docs/modules/pulse.md#api-endpoints) |
| Deployment | [docs/deployment/docker.md](../docs/deployment/docker.md) |
| Architecture | [docs/architecture/overview.md](../docs/architecture/overview.md) |

## License

AGPL-3.0 — See [LICENSE](./LICENSE)