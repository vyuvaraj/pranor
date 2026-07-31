# Observability

Every Pranor module emits traces, metrics, and logs through OpenTelemetry.

## Stack

```
Your Services → Pranor Trace (OTLP collector) → Pranor Console (dashboard)
                                              → Jaeger / Grafana (optional)
```

## Distributed Tracing

All modules propagate W3C `traceparent` headers automatically. A single request generates a connected trace across Gate → Mesh → Service → Pulse → Vault.

### Setup

Set one environment variable on every module:

```bash
PRANOR_OTLP_ENDPOINT=http://pranor-trace:8090
```

### View Traces

Open Pranor Console at `http://localhost:8083`:
- Waterfall trace view
- Service dependency map
- Latency percentiles (p50, p95, p99)
- Error rate tracking
- Trace search by ID, service, duration

## Metrics

Every module exposes Prometheus-compatible metrics at `GET /metrics`:

| Metric | Type | Description |
|--------|------|-------------|
| `pranor_http_requests_total` | Counter | Total HTTP requests by route, method, status |
| `pranor_http_duration_seconds` | Histogram | Request latency distribution |
| `pranor_queue_messages_total` | Counter | Messages published/consumed (Pulse) |
| `pranor_queue_consumer_lag` | Gauge | Consumer group lag (Pulse) |
| `pranor_cache_hits_total` | Counter | Cache hit/miss ratio (Cache) |
| `pranor_pool_connections_active` | Gauge | Active DB connections (Pool) |
| `pranor_vault_objects_total` | Gauge | Stored objects count (Vault) |

## Structured Logging

All modules emit JSON-structured logs with trace correlation:

```json
{
  "level": "info",
  "msg": "request completed",
  "trace_id": "abc123def456",
  "span_id": "789xyz",
  "service": "pranor-gate",
  "duration_ms": 42,
  "status": 200
}
```

## Alerting

Pranor Console supports SLO-based burn rate alerts:

- Define SLOs (99.9% availability, p95 < 200ms)
- Multi-window burn rate detection
- Alert routing to Slack, email, PagerDuty

## Health Checks

Every module exposes:

| Endpoint | Purpose |
|----------|---------|
| `GET /healthz` | Liveness (is the process running?) |
| `GET /readyz` | Readiness (can it serve traffic?) |
| `GET /metrics` | Prometheus metrics |

## Grafana Integration

Pranor Trace implements the Prometheus remote-write receiver protocol. Point your existing Grafana at:

```
Data Source: Prometheus
URL: http://pranor-trace:8090/api/v1/query
```

## Next Steps

- [Security Model](./security.md)
- [Trace Module Docs](../modules/trace.md)
- [Console Module Docs](../modules/console.md)
