# Explorer 3 Survey Handoff Report

## 1. Observation

### Codebase Scope & Directory Locations
- Root repository path: `/home/developer/workspace/serv`
- Pranor Pool package: `/home/developer/workspace/serv/packages/Pranor Pool`
- Pranor Pulse package: `/home/developer/workspace/serv/packages/Pranor Pulse`
- Original Requirements specification: `/home/developer/workspace/serv/.agents/ORIGINAL_REQUEST.md` (Lines 109-143, Requirements R8, R9, R10).

### Key Observations in Pranor Pool
1. **Module & Package Setup (`packages/Pranor Pool/go.mod`)**:
   - Module name: `github.com/vyuvaraj/pranor/packages/Pranor Pool`
   - Go version: `1.23.0`
   - Dependencies: `github.com/vyuvaraj/pranor/packages/Pranor Core v0.0.0`
2. **Pool Interface (`packages/Pranor Pool/pkg/pool/pool.go`, Lines 20-35)**:
   - `DbConn` struct definition:
     ```go
     type DbConn struct {
         ID           int
         CreatedAt    time.Time
         CheckedOutAt time.Time
     }
     ```
   - `Manager` interface definition:
     ```go
     type Manager interface {
         Acquire() (*DbConn, error)
         Release(conn *DbConn)
         IncrementQueries()
         Stats() PoolStats
         Dialect() string
         Shutdown(ctx context.Context) error
     }
     ```
3. **Routing Setup (`packages/Pranor Pool/pkg/routing/routing.go`)**:
   - `Server` struct uses `primaryPool pool.Manager` and `replicaPool pool.Manager`.
   - Currently `pkg/routing/rw_splitter.go` does **not** exist and needs to be created.
   - Currently `pkg/pool/health_checker.go` does **not** exist and needs to be created.

### Key Observations in Pranor Pulse
1. **Module & Package Setup (`packages/Pranor Pulse/go.mod`)**:
   - Module name: `github.com/vyuvaraj/pranor/packages/Pranor Pulse`
   - Go version: `1.25.0`
2. **Core Log Engine (`packages/Pranor Pulse/pkg/core/engine.go`, Lines 10-27, 42-58, 141-183)**:
   - `LogEntry` struct definition:
     ```go
     type LogEntry struct {
         Offset    uint64 `json:"offset"`
         Topic     string `json:"topic"`
         Payload   string `json:"payload"`
         Timestamp int64  `json:"timestamp"`
         Synced    bool   `json:"synced"`
     }
     ```
   - `StorageDriver` interface definition:
     ```go
     type StorageDriver interface {
         Append(topic, payload string) (LogEntry, error)
         ReadRange(topic string, startOffset, limit uint64) ([]LogEntry, error)
         SeekToTime(topic string, targetTimestamp int64) (uint64, error)
         GetUnsynced(limit uint64) ([]LogEntry, error)
         MarkSynced(offsets []uint64) error
         Recover() ([]LogEntry, error)
         Flush() error
         Close() error
     }
     ```
   - `Engine` struct methods: `NewEngine`, `SetEncryptionKey`, `Enqueue`, `Dequeue`, `SeekToTime`, `GetPendingSync`, `AcknowledgeSync`, `Close`.
   - Missing package: `packages/Pranor Pulse/pkg/tracing` directory does not exist yet. `traceparent.go` needs to be created in `pkg/tracing`.
   - `LogEntry` does not currently store trace context. `LogEntry` needs `Traceparent string \`json:"traceparent,omitempty"\`` field.
   - `Engine` needs an `Append(topic, payload string, metadata ...map[string]string) (LogEntry, error)` method.

---

## 2. Logic Chain

1. **SP.G1 Read/Write Split Router**:
   - Query routing relies on classifying SQL queries. SQL verbs `SELECT`, `WITH`, `EXPLAIN`, `SHOW` represent read operations, whereas `INSERT`, `UPDATE`, `DELETE`, `CREATE`, `DROP`, `ALTER`, `TRUNCATE`, `REPLACE` represent write/mutation operations.
   - SQL queries may contain leading whitespace, newlines, or comments. Stripping whitespace/comments and converting the leading verb to uppercase ensures robust classification regardless of query formatting or casing.
   - Read queries route to replica pools when available, using round-robin distribution (`atomic.AddUint64(&s.rrIndex, 1) - 1 % len(replicas)`). If no replicas are available, fallback to the primary pool. Write queries always route to the primary pool.

2. **SP.G2 Connection Health Validation**:
   - Connection pools can hand out dead/stale socket handles if network drops or database restarts occur while idle.
   - Implementing `HealthChecker` as a decorator pattern wrapping `pool.Manager` allows seamless drop-in connection validation.
   - When `Acquire()` is called, `HealthChecker` acquires a connection from the underlying pool and invokes `ValidateFn(conn)`. If invalid, `StaleDiscarded` is incremented, `Release(conn)` discards the bad handle, and up to 3 retry attempts are made before returning an error. Valid handles increment `HealthyAcquires` and are returned to the caller.

3. **SQ.G5 W3C Trace Context Propagation**:
   - Distributed tracing requires format compliance with the W3C Trace Context specification (`00-{traceID}-{spanID}-{flags}`).
   - `Inject(headers, traceID, spanID)` formats the header string `00-{traceID}-{spanID}-01` and writes it to `headers["traceparent"]`.
   - `Extract(headers)` searches case-insensitively for the `traceparent` key, parses the 4 hyphen-delimited fields, validates length (32 hex characters for `traceID`, 16 hex characters for `spanID`), verifies hex decoding, and checks the sampling bit (`flags & 0x01`).
   - `NewTraceID()` and `NewSpanID()` use `crypto/rand` to generate cryptographically secure 16-byte and 8-byte values encoded as hex strings.
   - Integrating into `pkg/core/engine.go`'s `Append` method allows callers to pass an optional metadata map `map[string]string{"traceparent": header}`. `Engine.Append` extracts `traceparent` and sets `LogEntry.Traceparent`.

---

## 3. Caveats

- **No external dependencies**: Implementations must rely only on standard Go packages (`strings`, `sync`, `crypto/rand`, `encoding/hex`, `fmt`, `errors`, `context`). No external YAML, SQL parser, or OTel libraries may be imported.
- **SQL Comment Parsing**: Simple comment stripping (handling `--` and `/* ... */`) is sufficient for standard SQL query inputs, but complex nested comments should default safely to `QueryTypeWrite`.
- **Concurrency & In-Memory Drivers**: When `Engine.Append` updates `LogEntry.Traceparent`, `MemoryDriver.entries` must be updated under mutex lock so `Dequeue` returns the populated `Traceparent`.

---

## 4. Conclusion

The technical design for SP.G1, SP.G2, and SQ.G5 is fully specified and aligned with `ORIGINAL_REQUEST.md`. Implementation can proceed cleanly by creating:
1. `packages/Pranor Pool/pkg/routing/rw_splitter.go` & `rw_splitter_test.go`
2. `packages/Pranor Pool/pkg/pool/health_checker.go` & `health_checker_test.go`
3. `packages/Pranor Pulse/pkg/tracing/traceparent.go` & `traceparent_test.go`
4. Modifying `packages/Pranor Pulse/pkg/core/engine.go` and `engine_test.go`.

---

## 5. Verification Method

### Step-by-Step Build & Test Verification
```bash
# 1. Verify Pranor Pool
cd /home/developer/workspace/serv/packages/Pranor Pool
go build ./...
go test ./... -v

# 2. Verify Pranor Pulse
cd /home/developer/workspace/serv/packages/Pranor Pulse
go build ./...
go test ./... -v
```

### Invalidation Conditions
- Any build failure or non-zero exit code on `go build ./...`.
- Any failing unit test or missing test cases for edge/error conditions.
- Addition of any new module in `go.mod`.
