# Architecture Overview

Pranor is a modular backend infrastructure engine. Each module runs independently or together as a unified platform.

## System Diagram

```
                         ┌─────────────────────┐
                         │     Clients         │
                         │  (Web, Mobile, API) │
                         └──────────┬──────────┘
                                    │
                         ┌──────────▼──────────┐
                         │    Pranor Gate      │
                         │   API Gateway &     │
                         │   Ingress Router    │
                         └──────────┬──────────┘
                                    │
              ┌─────────────────────┼─────────────────────┐
              │                     │                     │
    ┌─────────▼────────┐  ┌────────▼────────┐  ┌────────▼────────┐
    │   Pranor Auth    │  │  Pranor Mesh    │  │  Pranor Cache   │
    │  Identity/RBAC   │  │ Service Discovery│  │  Redis/Memory   │
    └──────────────────┘  └────────┬────────┘  └─────────────────┘
                                   │
         ┌─────────────────────────┼─────────────────────────┐
         │                         │                         │
┌────────▼────────┐  ┌────────────▼────────────┐  ┌─────────▼────────┐
│  Your Services  │  │     Pranor Pulse        │  │   Pranor Vault   │
│  (.pnr files)   │  │  Async Event Broker     │  │  Object Storage  │
└────────┬────────┘  └────────────┬────────────┘  └──────────────────┘
         │                        │
         │            ┌───────────┼───────────┐
         │            │           │           │
┌────────▼───────┐ ┌──▼──┐ ┌─────▼─────┐ ┌───▼───┐
│ Pranor Chrono  │ │Flow │ │  Notify   │ │ Pool  │
│  Scheduler     │ │     │ │ Email/SMS │ │  DB   │
└────────────────┘ └─────┘ └───────────┘ └───────┘
         │
┌────────▼───────────────────────────────────┐
│              Pranor Trace                   │
│         Distributed Tracing (OTLP)         │
└────────────────────────────────────────────┘
         │
┌────────▼───────────────────────────────────┐
│            Pranor Console                   │
│       Observability Dashboard UI           │
└────────────────────────────────────────────┘
```

## How Modules Connect

| From | To | Protocol | Purpose |
|------|-----|----------|---------|
| Gate → Services | `pranor://` | HTTP/gRPC via Mesh | Route requests to backends |
| Services → Pulse | STOMP/TCP | Async messaging | Publish events, consume queues |
| Services → Vault | S3 HTTP API | Object storage | Store files, vectors, configs |
| Services → Cache | Redis protocol | Caching | TTL-based key-value cache |
| Services → Pool | PostgreSQL wire | DB proxy | Connection pooling, read/write split |
| All → Trace | OTLP HTTP | Telemetry | Spans, metrics, logs |
| Chrono → Services | HTTP webhook | Scheduling | Trigger jobs on cron schedule |
| Flow → Services | HTTP | Orchestration | DAG workflow execution |
| Auth → Gate | JWT validation | Security | Token verification on every request |

## Module Independence

Each module is:
- A standalone Go binary with zero external dependencies
- Independently deployable (Docker, K8s, bare metal)
- Horizontally scalable
- Observable via standard OTLP

No module requires any other module to function. The Pranor language compiler orchestrates them together when running a unified `.pnr` service, but each works alone.

## Data Flow Example

A typical request through the full platform:

1. Client sends `POST /api/orders` to **Gate** (port 8080)
2. Gate validates JWT via **Auth**, applies rate limiting
3. Gate routes to your order service via **Mesh** discovery
4. Your service writes order to **Vault** (S3 storage)
5. Your service publishes `order.created` event to **Pulse**
6. **Chrono** triggers a delayed notification job
7. **Notify** sends confirmation email/SMS
8. **Trace** captures the full request waterfall
9. **Console** displays the trace in real-time

## Next Steps

- [Security Model](./security.md)
- [Observability](./observability.md)
- [Module Docs](../modules/) — Detailed per-module documentation
