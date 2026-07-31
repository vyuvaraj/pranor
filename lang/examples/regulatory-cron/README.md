# Regulatory Cron Service — Pranor Port

A port of the Java regulatory-cron-service (Spring Boot 3.4 / Java 21) to Pranor, demonstrating how a complex production microservice maps to Pranor's declarative syntax.

## Quick Start

### Prerequisites
- Go 1.18+ (for building the Pranor compiler)
- MongoDB running on `localhost:27017`
- Pranor compiler built (`go build -o pranor.exe main.go` from Pranor root)

### Build & Run

```bash
# From the Pranor root directory
.\pranor.exe build examples\regulatory-cron\main.pnr -o examples\regulatory-cron\cron-service.exe

# Run from the regulatory-cron directory (so config.yml is found)
cd examples\regulatory-cron
.\cron-service.exe
```

### Configuration

The service reads `config.yml` from the working directory. Alternatively:

```bash
# Override with env vars
set database.uri=mongodb://localhost:27017/cron_service
set server.port=8087

# Or point to config explicitly
.\cron-service.exe --config C:\path\to\config.yml

# Or use PRANOR_CONFIG env var
set PRANOR_CONFIG=C:\path\to\config.yml
.\cron-service.exe
```

### Verify

```bash
# Health check
curl http://localhost:8087/health

# Readiness (checks MongoDB connectivity)
curl http://localhost:8087/ready

# List seeded jobs (12 sample jobs on first run)
curl http://localhost:8087/api/v1/jobs

# Runtime stats
curl http://localhost:8087/api/v1/stats

# Prometheus metrics
curl http://localhost:8087/metrics
```

## Project Structure

```
regulatory-cron/
├── main.pnr                         # Entry point (imports + seed call)
├── config.pnr                       # Server/DB config (reads config.yml)
├── config.yml                       # YAML configuration
├── scheduler/
│   ├── engine.pnr                   # Job queue, workers, dispatch, schedulers
│   └── seeder.pnr                   # Seeds 12 sample jobs if DB is empty
├── jobs/
│   ├── data_sync.pnr               # External API sync pattern
│   ├── batch_processing.pnr        # Chunked DB processing pattern
│   └── cleanup.pnr                 # Retention-based cleanup pattern
├── api/
│   ├── middleware.pnr              # X-Actor header middleware
│   ├── jobs.pnr                    # CRUD + pause/resume/trigger
│   ├── dlq.pnr                     # Dead Letter Queue endpoints
│   ├── bulk.pnr                    # Bulk operations
│   └── stats.pnr                   # Observability stats
├── models/
│   └── domain.pnr                  # Structs & enums
├── services/
│   ├── executor.pnr                # Retry engine with backoff
│   ├── lock.pnr                    # MongoDB distributed locks
│   ├── dlq.pnr                     # DLQ service logic
│   └── alert.pnr                   # Webhook alerts
└── README.md
```

## REST API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check (auto-generated) |
| GET | `/ready` | Readiness check with DB connectivity |
| GET | `/metrics` | Prometheus-format metrics |
| GET | `/api/v1/jobs` | List jobs (paginated) |
| GET | `/api/v1/jobs/:id` | Get job details |
| POST | `/api/v1/jobs` | Create job |
| PUT | `/api/v1/jobs/:id` | Update job |
| DELETE | `/api/v1/jobs/:id` | Delete job |
| POST | `/api/v1/jobs/:id/pause` | Pause job |
| POST | `/api/v1/jobs/:id/resume` | Resume job |
| POST | `/api/v1/jobs/:id/trigger` | Manual trigger (queues for execution) |
| GET | `/api/v1/jobs/:id/executions` | Execution history |
| GET | `/api/v1/dlq` | Dead Letter Queue entries |
| POST | `/api/v1/dlq/:id/replay` | Replay failed execution |
| POST | `/api/v1/jobs/bulk/pause` | Bulk pause |
| POST | `/api/v1/jobs/bulk/resume` | Bulk resume |
| GET | `/api/v1/stats` | Runtime stats (running jobs, queue depth) |

## Architecture Comparison

| Component | Java (Spring Boot) | Pranor |
|-----------|-------------------|------|
| Scheduler | PriorityBlockingQueue + virtual thread | `every 5s` + `channel` + worker pool |
| Job execution | Spring bean resolution + Future.get | `spawn` + channel workers |
| Distributed lock | MongoDB insert + DuplicateKeyException | `db.query("insert", "locks", ...)` |
| Retry logic | For loop + Thread.sleep(backoff) | For loop + `time.sleep(delay)` |
| REST API | @RestController (12 endpoints) | `route` declarations |
| DLQ | MongoDB + @Scheduled purge | `db.query` + `cron "0 0 2 * * *"` |
| Alerts | HttpClient + retry + rate limit | `http.post` + `cache.set` rate limit |
| Metrics | Micrometer + Prometheus | `metric.inc()` + `/metrics` |
| Health | Spring Actuator | Auto-generated `/health` + `/ready` |
| Concurrency | AtomicBoolean + ConcurrentHashMap | `atomic.new/inc/dec/get` |
| Worker pool | Virtual thread executor | `spawn` + `channel.receive` loop |
| Middleware | Spring Security filter | `middleware xActor(req) { }` |
| Logging | Logback + Logstash JSON | `log.with()` + `LOG_FORMAT=json` |
| Config | application.yml + Spring profiles | `config.yml` + `config("key")` |
| Data seeding | CommandLineRunner + @Profile("dev") | `seedIfEmpty()` at startup |

## Line Count

| | Java | Pranor | Reduction |
|---|------|------|-----------|
| Total source | ~1350 lines | ~430 lines | **68%** |
| Config/boilerplate | ~300 lines | 3 lines | **99%** |
| Build config | ~50 lines (Gradle) | 0 lines | **100%** |

## Sample Jobs (Seeded on First Run)

| Job | Schedule | Jurisdiction | Pattern |
|-----|----------|-------------|---------|
| kyc-reminder-batch-it | Every 30 min | Italy | Batch processing |
| aams-sync-daily | 3 AM daily | Italy | External API sync |
| gamstop-check-uk | Every 15 min | UK | External API sync |
| cnj-verification-brazil | Every 2 hours | Brazil | External API sync (PAUSED) |
| tsupis-sync-greece | 4 AM daily | Greece | External API sync |
| account-closure-batch | 2 AM daily | IT/UK/DE | Batch processing |
| audit-log-cleanup | 1 AM daily | All | Cleanup |
| cache-invalidation-all | Every 5 min | All | Cleanup |
| kyc-reminder-batch-de | 8 AM daily | Germany | Batch processing |
| jumio-verification-check | Every 10 min | IT/UK/DE/GR | External API sync |
| player-dormancy-check-es | 6 AM daily | Spain | Batch processing (PAUSED) |
| regulatory-report-ro | 5 AM Monday | Romania | Batch processing |

## Troubleshooting

**"unsupported database schema" error:**
- The service can't find `config.yml`. Run from the `regulatory-cron/` directory, or set `PRANOR_CONFIG` env var.

**"I/O error on GET request... null":**
- The service isn't running or is on a different port. Check console output for "listening on port 8087".

**Port mismatch:**
- Default port is in `config.yml` (`server.port`). Override with env: `set server.port=8087`

**MongoDB connection failed:**
- Ensure MongoDB is running on `localhost:27017`
- Override URI: `set database.uri=mongodb://localhost:27017/cron_service`
