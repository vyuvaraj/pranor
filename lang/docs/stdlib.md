# Standard Library

Serv ships with 46 reusable modules in `stdlib/`. Import what you need:

```serv
import { ok, notFound } from "../stdlib/response.pnr"
import { requireAuth } from "../stdlib/auth.pnr"
```

## Quick Reference

### Security
| Module | Key Exports |
|--------|-------------|
| `auth.pnr` | `bearerToken`, `basicAuth`, `requireAuth` |
| `crypto.pnr` | `hashPassword`, `verifyPassword`, `randomToken`, `hmacSign` |
| `jwt.pnr` | `jwtEncode`, `jwtDecode`, `jwtIsExpired` |
| `sanitize.pnr` | `escapeHTML`, `stripTags`, `escapeSQL`, `sanitizeFilename` |
| `ratelimit.pnr` | `createLimiter`, `isAllowed`, `remaining`, `resetLimiter` |
| `mask.pnr` | `maskEmail`, `maskPhone`, `maskCard`, `maskString`, `redact` |
| `ip.pnr` | `extractIP`, `isPrivate`, `isTrustedProxy`, `anonymizeIP` |

### HTTP
| Module | Key Exports |
|--------|-------------|
| `response.pnr` | `ok`, `created`, `badRequest`, `notFound`, `serverError` |
| `pagination.pnr` | `offset`, `pageResponse`, `parsePageParams` |
| `pagination_cursor.pnr` | `encodeCursor`, `decodeCursor`, `cursorResponse` |
| `middleware.pnr` | `corsHeaders`, `requestId`, `logRequest` |
| `http_client.pnr` | `getJSON`, `postJSON`, `isSuccess`, `isClientError` |
| `url.pnr` | `encodeURI`, `parseQuery`, `buildQuery`, `joinPath` |
| `cors.pnr` | `allowOrigin`, `allowAll`, `preflightResponse` |

### Utilities
| Module | Key Exports |
|--------|-------------|
| `datetime.pnr` | `now`, `timestamp`, `isExpired`, `formatDuration` |
| `strings_util.pnr` | `slugify`, `truncate`, `capitalize`, `isEmpty` |
| `math.pnr` | `min`, `max`, `clamp`, `abs`, `percent`, `sum`, `average` |
| `sort.pnr` | `reverse`, `minOf`, `maxOf` |
| `collections.pnr` | `unique`, `flatten`, `chunk`, `first`, `last`, `countWhere` |

### Data
| Module | Key Exports |
|--------|-------------|
| `csv.pnr` | `parseCSV`, `parseRow`, `toCSV` |
| `base64.pnr` | `encode`, `decode`, `isValid` |
| `diff.pnr` | `hasChanged`, `fieldChanged`, `changeRecord` |

### Config
| Module | Key Exports |
|--------|-------------|
| `env.pnr` | `requireEnv`, `envOrDefault`, `envInt`, `envBool` |
| `config.pnr` | `getConfig`, `requireConfig`, `configBool`, `configList` |
| `feature_flags.pnr` | `enableFlag`, `disableFlag`, `isEnabled`, `toggleFlag` |

### Resilience
| Module | Key Exports |
|--------|-------------|
| `retry.pnr` | `backoffDelay`, `defaultMaxRetries` |
| `circuit_breaker.pnr` | `createBreaker`, `isOpen`, `recordSuccess`, `recordFailure` |
| `timeout.pnr` | `withDeadline`, `isTimedOut`, `remainingTime`, `elapsed` |
| `queue.pnr` | `createQueue`, `enqueue`, `dequeue`, `queueSize` |

### Concurrency
| Module | Key Exports |
|--------|-------------|
| `semaphore.pnr` | `createSemaphore`, `tryAcquire`, `release`, `available` |
| `batch.pnr` | `createBatch`, `addToBatch`, `isBatchFull`, `flushBatch` |

### Processing
| Module | Key Exports |
|--------|-------------|
| `job.pnr` | `createJob`, `startJob`, `completeJob`, `failJob` |
| `scheduler.pnr` | `scheduleAfter`, `isScheduled`, `cancelSchedule` |

### Reliability
| Module | Key Exports |
|--------|-------------|
| `idempotency.pnr` | `checkIdempotency`, `markProcessed`, `isProcessed` |
| `dlq.pnr` | `createDLQ`, `sendToDLQ`, `dlqSize`, `clearDLQ` |

### Integration
| Module | Key Exports |
|--------|-------------|
| `webhook.pnr` | `buildPayload`, `sendWebhook`, `verifySignature` |
| `events.pnr` | `on`, `emit`, `hasHandler` |

### Observability
| Module | Key Exports |
|--------|-------------|
| `metrics.pnr` | `counter`, `gauge`, `recordLatency`, `trackRequest` |
| `tracing.pnr` | `traceId`, `startSpan`, `endSpan`, `traceContext` |

### Multi-tenancy
| Module | Key Exports |
|--------|-------------|
| `tenant.pnr` | `extractTenant`, `tenantConfig`, `isTenantActive`, `tenantFilter` |

### Compliance
| Module | Key Exports |
|--------|-------------|
| `audit.pnr` | `auditLog`, `auditAction`, `auditAccess`, `auditAuth`, `auditDenied` |

### Operations
| Module | Key Exports |
|--------|-------------|
| `health.pnr` | `healthy`, `unhealthy`, `degraded`, `buildHealthResponse` |
| `graceful.pnr` | `initShutdown`, `isShuttingDown`, `isDrained` |
| `cache_patterns.pnr` | `cacheKey`, `cacheGet`, `cacheSet`, `invalidate`, `computeIfAbsent` |

### Testing
| Module | Key Exports |
|--------|-------------|
| `testing_helpers.pnr` | `assertEqual`, `assertNotNil`, `assertContains`, `assertTrue` |

## Usage Example

```serv
import { requireAuth, bearerToken } from "../stdlib/auth.pnr"
import { ok, badRequest } from "../stdlib/response.pnr"
import { maskEmail } from "../stdlib/mask.pnr"
import { auditLog } from "../stdlib/audit.pnr"

server "8080"

route "GET" "/api/profile" (req) {
    let authErr = requireAuth(req)
    if authErr != nil { return authErr }

    let token = bearerToken(req)
    auditLog(token, "view", "profile", nil)

    return ok({
        "email": maskEmail("alice@example.com"),
        "role": "admin"
    })
}
```

Full module documentation: see comments at the top of each file in `stdlib/`.

## Deep-dive Module Examples

### 1. Resilience (`circuit_breaker.pnr` & `retry.pnr`)

Implement standard fault tolerance when requesting downstreams:

```serv
import { createBreaker, recordSuccess, recordFailure, isOpen } from "../stdlib/circuit_breaker.pnr"
import { backoffDelay } from "../stdlib/retry.pnr"

let breaker = createBreaker(3, 10s) // 3 failures, 10s timeout window

fn requestExternalAPI(url) {
    if isOpen(breaker) {
        return { "error": "circuit breaker open", "status": "failing" }
    }

    let res = http.get(url)
    if res.status != 200 {
        recordFailure(breaker)
        return nil
    }

    recordSuccess(breaker)
    return res.body
}
```

### 2. Concurrency (`semaphore.pnr`)

Guard resources from over-concurrency:

```serv
import { createSemaphore, tryAcquire, release } from "../stdlib/semaphore.pnr"

let sem = createSemaphore(5) // Max 5 parallel tasks

fn processTask(taskId) {
    if !tryAcquire(sem) {
        log.warn("Rate limited locally — too many concurrent workers")
        return false
    }
    defer release(sem)

    heavyComputation(taskId)
    return true
}
```

### 3. Masking & Compliance (`mask.pnr` & `audit.pnr`)

Sanitize user profiles and write logs:

```serv
import { maskEmail, maskCard } from "../stdlib/mask.pnr"
import { auditLog } from "../stdlib/audit.pnr"

fn getSafeProfile(req) {
    let email = req.body.email
    let card = req.body.card

    auditLog(req.user, "access", "billing-profile", nil)

    return {
        "email": maskEmail(email),
        "card": maskCard(card)
    }
}
```

