# Serv Standard Library

Reusable `.pnr` modules for common service patterns. Import what you need:

```serv
import { ok, notFound } from "stdlib/response.pnr"
import { requireAuth, bearerToken } from "stdlib/auth.pnr"
```

---

## Module Index

| Module | Exports | Category |
|--------|---------|----------|
| `auth.pnr` | `bearerToken`, `basicAuth`, `requireAuth` | Security |
| `crypto.pnr` | `hashPassword`, `verifyPassword`, `randomToken`, `randomHex`, `hmacSign`, `hmacVerify` | Security |
| `jwt.pnr` | `jwtEncode`, `jwtDecode`, `jwtIsExpired` | Security |
| `sanitize.pnr` | `escapeHTML`, `stripTags`, `escapeSQL`, `sanitizeFilename`, `normalizeWhitespace` | Security |
| `ratelimit.pnr` | `createLimiter`, `isAllowed`, `remaining`, `resetLimiter` | Security |
| `validation.pnr` | `required`, `isEmail`, `isURL`, `minLength`, `maxLength`, `validateFields` | Input |
| `response.pnr` | `ok`, `created`, `badRequest`, `notFound`, `serverError`, `errorResponse` | HTTP |
| `pagination.pnr` | `offset`, `pageResponse`, `parsePageParams` | HTTP |
| `middleware.pnr` | `corsHeaders`, `requestId`, `logRequest`, `isPreflight` | HTTP |
| `http_client.pnr` | `getJSON`, `postJSON`, `isSuccess`, `isClientError`, `isServerError` | HTTP |
| `url.pnr` | `encodeURI`, `parseQuery`, `buildQuery`, `joinPath`, `extractPath` | HTTP |
| `datetime.pnr` | `now`, `timestamp`, `isExpired`, `formatDuration`, `sleep` | Utilities |
| `strings_util.pnr` | `slugify`, `truncate`, `capitalize`, `isEmpty`, `repeat`, `matches` | Utilities |
| `math.pnr` | `min`, `max`, `clamp`, `abs`, `percent`, `between`, `sum`, `average` | Utilities |
| `sort.pnr` | `sortAsc`, `sortDesc`, `reverse`, `minOf`, `maxOf` | Utilities |
| `collections.pnr` | `groupBy`, `unique`, `flatten`, `chunk`, `first`, `last`, `countWhere` | Data |
| `csv.pnr` | `parseCSV`, `parseRow`, `toRow`, `toCSV` | Data |
| `diff.pnr` | `hasChanged`, `fieldChanged`, `changeRecord` | Data |
| `env.pnr` | `requireEnv`, `envOrDefault`, `envInt`, `envBool`, `envExists` | Config |
| `retry.pnr` | `backoffDelay`, `defaultMaxRetries`, `defaultBaseDelay` | Resilience |
| `circuit_breaker.pnr` | `createBreaker`, `isOpen`, `recordSuccess`, `recordFailure`, `resetBreaker`, `failureCount` | Resilience |
| `queue.pnr` | `createQueue`, `enqueue`, `dequeue`, `queueSize`, `queueIsEmpty` | Resilience |
| `events.pnr` | `on`, `emit`, `hasHandler` | Messaging |
| `metrics.pnr` | `counter`, `counterWithLabel`, `gauge`, `recordLatency`, `trackRequest` | Observability |
| `testing_helpers.pnr` | `assertEqual`, `assertNotNil`, `assertNil`, `assertContains`, `assertTrue`, `assertFalse`, `assertLength` | Testing |
| `health.pnr` | `healthy`, `unhealthy`, `degraded`, `buildHealthResponse` | Ops |
| `scheduler.pnr` | `scheduleAfter`, `isScheduled`, `cancelSchedule`, `getDelay` | Scheduling |
| `webhook.pnr` | `buildPayload`, `sendWebhook`, `verifySignature`, `retryRecord` | Integration |
| `cors.pnr` | `allowOrigin`, `allowAll`, `preflightResponse`, `isOriginAllowed` | HTTP |
| `graceful.pnr` | `initShutdown`, `isShuttingDown`, `connectionOpened`, `connectionClosed`, `isDrained` | Ops |
| `tracing.pnr` | `traceId`, `spanId`, `startSpan`, `endSpan`, `addTag`, `traceContext` | Observability |
| `semaphore.pnr` | `createSemaphore`, `tryAcquire`, `release`, `available`, `utilization` | Concurrency |
| `batch.pnr` | `createBatch`, `addToBatch`, `batchSize`, `isBatchFull`, `flushBatch` | Processing |
| `idempotency.pnr` | `checkIdempotency`, `markProcessed`, `isProcessed`, `getProcessedResult` | Reliability |
| `job.pnr` | `createJob`, `startJob`, `completeJob`, `failJob`, `jobStatus` | Processing |
| `feature_flags.pnr` | `enableFlag`, `disableFlag`, `isEnabled`, `toggleFlag`, `initFlag` | Config |
| `config.pnr` | `getConfig`, `requireConfig`, `configInt`, `configBool`, `configList`, `hasConfig` | Config |
| `tenant.pnr` | `extractTenant`, `tenantConfig`, `isTenantActive`, `tenantCacheKey`, `tenantFilter` | Multi-tenancy |
| `dlq.pnr` | `createDLQ`, `sendToDLQ`, `dlqSize`, `dlqHasItems`, `clearDLQ` | Reliability |
| `audit.pnr` | `auditLog`, `auditAction`, `auditAccess`, `auditAuth`, `auditDenied` | Compliance |
| `cache_patterns.pnr` | `cacheKey`, `cacheGet`, `cacheSet`, `invalidate`, `invalidatePrefix`, `cacheTTL`, `computeIfAbsent` | Caching |
| `pagination_cursor.pnr` | `encodeCursor`, `decodeCursor`, `hasCursor`, `extractCursor`, `cursorResponse`, `cursorResponseWith` | HTTP |
| `timeout.pnr` | `withDeadline`, `isTimedOut`, `remainingTime`, `startTimer`, `elapsed`, `hasExceeded` | Resilience |
| `ip.pnr` | `extractIP`, `isPrivate`, `isTrustedProxy`, `rateLimitKey`, `anonymizeIP` | Security |
| `mask.pnr` | `maskEmail`, `maskPhone`, `maskCard`, `maskString`, `redact` | Security |

---

## Categories

### Security
- **auth.pnr** — Token extraction, bearer/basic auth, auth guards
- **crypto.pnr** — Password hashing, HMAC signing, token generation
- **jwt.pnr** — JWT encode/decode/expiry (lightweight; use `serv add github.com/golang-jwt/jwt/v5` for production)

### HTTP
- **response.pnr** — Standard HTTP response builders (ok, notFound, etc.)
- **pagination.pnr** — Page offset calculation, response envelope
- **middleware.pnr** — CORS headers, request ID generation, preflight detection
- **http_client.pnr** — JSON GET/POST wrappers, status code helpers

### Utilities
- **datetime.pnr** — Timestamps, expiry checks, duration formatting
- **strings_util.pnr** — Slugify, truncate, capitalize, pattern matching
- **collections.pnr** — Array utilities (unique, flatten, chunk, first/last)

### Config & Environment
- **env.pnr** — Required env vars, defaults, type coercion (int/bool)

### Resilience
- **retry.pnr** — Exponential backoff calculation

### Messaging
- **events.pnr** — In-process event bus (emit/on pattern)

### Testing
- **testing_helpers.pnr** — Expressive assertions for test blocks

### Operations
- **health.pnr** — Custom health check builders
- **graceful.pnr** — Shutdown state, connection draining, drain detection

### Scheduling
- **scheduler.pnr** — Dynamic runtime scheduling beyond `every`/`cron`

### Integration
- **webhook.pnr** — Outgoing webhook payloads, signature verification, retry records
- **cors.pnr** — CORS header generation, origin checking, preflight responses

### Concurrency
- **semaphore.pnr** — Named semaphores with slot tracking and utilization metrics

### Processing
- **batch.pnr** — Accumulate-and-flush batch pattern with size tracking
- **job.pnr** — Background job lifecycle (pending → running → completed/failed)

### Reliability
- **idempotency.pnr** — Idempotency key pattern for deduplication
- **dlq.pnr** — Dead letter queue for failed message tracking

### Multi-tenancy
- **tenant.pnr** — Tenant extraction from requests, scoped config/cache/DB keys

### Compliance
- **audit.pnr** — Structured audit trail (actions, access, auth, denied events)

---

## Usage Example

```serv
import { requireAuth, bearerToken } from "stdlib/auth.pnr"
import { ok, badRequest } from "stdlib/response.pnr"
import { required, isEmail } from "stdlib/validation.pnr"
import { envOrDefault } from "stdlib/env.pnr"

server envOrDefault("PORT", "8080")

route "POST" "/api/users" (req) {
    let authErr = requireAuth(req)
    if authErr != nil {
        return authErr
    }

    let errors = validate(req.body, {
        "email": "required,email",
        "name": "required"
    })
    if errors != nil {
        return badRequest(errors)
    }

    return ok({ "created": true })
}
```

---

## Contributing

Add new modules as `stdlib/<name>.pnr`. Export functions with `export fn`. Follow existing patterns:
- Pure functions where possible
- No side effects unless explicitly documented
- Use `interface{}` params (no type annotations) for maximum flexibility
