# Docker Deployment Guide

Run the full Pranor platform or individual modules using Docker Compose.

## Quick Start — Full Platform

```bash
git clone https://github.com/vyuvaraj/pranor.git
cd pranor
docker compose up -d
```

This starts all modules:

| Service | Port | Description |
|---------|------|-------------|
| pranor-gate | 8080 | API Gateway |
| pranor-vault | 8081 | Object Storage (S3) |
| pranor-pulse | 8082 | Message Broker (STOMP) |
| pranor-console | 8083 | Dashboard UI |
| pranor-deploy | 8085 | Deployment Orchestrator |
| pranor-cache | 8086 | Cache Engine |
| pranor-chrono | 8087 | Job Scheduler |
| pranor-hub | 8088 | Package Registry |
| pranor-mesh | 8089 | Service Mesh |
| pranor-trace | 8090 | Tracing Collector |
| pranor-notify | 8094 | Notification Gateway |
| pranor-flow | 8096 | Workflow Engine |
| pranor-pool | 8097 | DB Connection Pool |
| pranor-auth | 8098 | Auth Provider |
| pranor-tunnel | 8443 | Dev Tunnel |

## Single Module

Run any module standalone:

```bash
# Just the API Gateway
docker run -p 8080:8080 ghcr.io/vyuvaraj/pranor-gate:latest

# Just the Object Storage
docker run -p 8081:8081 -v vault-data:/data ghcr.io/vyuvaraj/pranor-vault:latest

# Just the Message Broker
docker run -p 8082:8082 -p 61613:61613 ghcr.io/vyuvaraj/pranor-pulse:latest
```

## Environment Variables

All modules accept:

| Variable | Description |
|----------|-------------|
| `PRANOR_OTLP_ENDPOINT` | OpenTelemetry collector URL |
| `PRANOR_DISCOVERY` | JSON map of module URLs for service discovery |

Module-specific variables are documented in each [module's docs](../modules/).

## Docker Compose (Production)

```yaml
services:
  pranor-gate:
    image: ghcr.io/vyuvaraj/pranor-gate:latest
    ports:
      - "8080:8080"
    environment:
      - PRANOR_OTLP_ENDPOINT=http://pranor-trace:8090
    depends_on:
      pranor-trace:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "--spider", "http://localhost:8080/healthz"]
      interval: 5s
      timeout: 3s
      retries: 5

  pranor-vault:
    image: ghcr.io/vyuvaraj/pranor-vault:latest
    ports:
      - "8081:8081"
    volumes:
      - vault-data:/data
    environment:
      - PRANOR_OTLP_ENDPOINT=http://pranor-trace:8090

  pranor-trace:
    image: ghcr.io/vyuvaraj/pranor-trace:latest
    ports:
      - "8090:8090"

networks:
  default:
    name: pranor-net

volumes:
  vault-data:
```

## Health Checks

Every module exposes `GET /healthz` returning:

```json
{"status": "UP", "service": "pranor", "version": "1.0.0"}
```

## Observability

Connect all modules to Pranor Trace for distributed tracing:

```bash
PRANOR_OTLP_ENDPOINT=http://pranor-trace:8090
```

View traces in Pranor Console at `http://localhost:8083` or forward to Jaeger/Grafana.

## Next Steps

- [Kubernetes Deployment](./kubernetes.md)
- [Running Standalone Binaries](./standalone.md)
- [Architecture Overview](../architecture/overview.md)
