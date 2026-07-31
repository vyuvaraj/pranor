# Survey & Analysis: SP.G1, SP.G2, and SQ.G5

## Executive Summary
This document provides a comprehensive survey and architectural analysis of three OSS roadmap features for the Pranor monorepo (`/home/developer/workspace/pranor`):
1. **SP.G1**: Read/Write split router in `packages/Pranor Pool/pkg/routing/rw_splitter.go` (Requirement R8).
2. **SP.G2**: Pre-checkout connection health validator in `packages/Pranor Pool/pkg/pool/health_checker.go` (Requirement R9).
3. **SQ.G5**: W3C trace context propagation in `packages/Pranor Pulse/pkg/tracing/traceparent.go` and engine integration in `packages/Pranor Pulse/pkg/core/engine.go` (Requirement R10).

All three features require **zero external dependencies**, relying strictly on Go's standard library (`strings`, `sync`, `time`, `crypto/rand`, `encoding/hex`, `fmt`, `errors`).

---

## 1. Existing Monorepo & Package Structure

### 1.1 Pranor Pool Architecture (`packages/Pranor Pool`)
- **`go.mod`**: Module `github.com/vyuvaraj/pranor/packages/Pranor Pool` (Go 1.23.0). Depends on `packages/Pranor Core`.
- **Existing Packages**:
  - `pkg/pool`: Defines `DbConn`, `PoolStats`, `Manager` interface, and `ConnectionPool` struct (`pkg/pool/pool.go`).
  - `pkg/routing`: Defines HTTP handler `Server`, `QueryRequest`, `QueryResponse`, and `QueryOptimizer` interface (`pkg/routing/routing.go`).
  - `pkg/analytics`: Query metrics tracking (`pkg/analytics/analytics.go`).
  - `pkg/migration`: Schema migration management (`pkg/migration/migration.go`).
- **Target Additions**:
  - `pkg/routing/rw_splitter.go` & `rw_splitter_test.go`
  - `pkg/pool/health_checker.go` & `health_checker_test.go`

### 1.2 Pranor Pulse Architecture (`packages/Pranor Pulse`)
- **`go.mod`**: Module `github.com/vyuvaraj/pranor/packages/Pranor Pulse` (Go 1.25.0).
- **Existing Packages**:
  - `pkg/core`: Defines `LogEntry`, `StorageDriver` interface, `MemoryDriver`, `Engine`, `Enqueue`, `Dequeue` (`pkg/core/engine.go`).
  - `pkg/broker`, `pkg/analytics`, `pkg/auth`, `pkg/cdc`, `pkg/opfs`, `pkg/storage`, etc.
- **Target Additions / Modifications**:
  - New package `pkg/tracing`: `pkg/tracing/traceparent.go` & `traceparent_test.go`
  - Modify `pkg/core/engine.go`: Add `Traceparent` to `LogEntry`, implement `Append(topic, payload string, metadata ...map[string]string) (LogEntry, error)` method on `Engine` and update `MemoryDriver`.

---

## 2. Requirement Specifications & Technical Designs

### 2.1 SP.G1 — Read/Write Split Router (`packages/Pranor Pool/pkg/routing/rw_splitter.go`)

#### Purpose
Classify SQL query statements into Read vs Write operations and route them to either a Primary database pool or a set of Replica database pools.

#### Design Specification
```go
package routing

import (
	"strings"
	"sync/atomic"

	"github.com/vyuvaraj/pranor/packages/Pranor Pool/pkg/pool"
)

type QueryType string

const (
	QueryTypeRead  QueryType = "READ"
	QueryTypeWrite QueryType = "WRITE"
)

type RWSplitter struct {
	rrIndex uint64
}

func NewRWSplitter() *RWSplitter {
	return &RWSplitter{}
}

// ClassifyQuery determines whether an SQL string is a read or write query.
// Handles leading whitespace, case-insensitivity, and SQL verb inspection.
func (s *RWSplitter) ClassifyQuery(sql string) QueryType {
	return ClassifyQuery(sql)
}

func ClassifyQuery(sql string) QueryType {
	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return QueryTypeWrite // Safe default
	}

	// Strip leading single-line comments (-- ...) or block comments (/* ... */) if needed
	for strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "/*") {
		if strings.HasPrefix(trimmed, "--") {
			idx := strings.Index(trimmed, "\n")
			if idx == -1 {
				return QueryTypeWrite
			}
			trimmed = strings.TrimSpace(trimmed[idx+1:])
		} else if strings.HasPrefix(trimmed, "/*") {
			idx := strings.Index(trimmed, "*/")
			if idx == -1 {
				return QueryTypeWrite
			}
			trimmed = strings.TrimSpace(trimmed[idx+2:])
		}
	}

	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return QueryTypeWrite
	}

	verb := strings.ToUpper(fields[0])
	switch verb {
	case "SELECT", "WITH", "EXPLAIN", "SHOW":
		return QueryTypeRead
	case "INSERT", "UPDATE", "DELETE", "CREATE", "DROP", "ALTER", "TRUNCATE", "REPLACE":
		return QueryTypeWrite
	default:
		return QueryTypeWrite
	}
}

// Route returns the appropriate pool.Manager based on query classification.
func (s *RWSplitter) Route(sql string, primary pool.Manager, replicas []pool.Manager) pool.Manager {
	qt := s.ClassifyQuery(sql)
	if qt == QueryTypeRead && len(replicas) > 0 {
		idx := atomic.AddUint64(&s.rrIndex, 1) - 1
		return replicas[idx%uint64(len(replicas))]
	}
	return primary
}
```

#### Test Coverage Requirements (`rw_splitter_test.go`)
- Verbs: `SELECT`, `WITH`, `INSERT`, `UPDATE`, `DELETE`, `CREATE`, `DROP`, `ALTER`.
- Case insensitivity: `select * from users`, `InSeRt InTo...`.
- Whitespace/newlines: `\t\n  SELECT * FROM users`.
- Routing: Read query routes to replica pool (round-robin across multiple replicas); fallback to primary when replicas slice is empty; Write query routes to primary pool.

---

### 2.2 SP.G2 — Connection Health Validation (`packages/Pranor Pool/pkg/pool/health_checker.go`)

#### Purpose
Wrap a `pool.Manager` to perform a pre-checkout validation check on every connection before returning it from `Acquire()`.

#### Design Specification
```go
package pool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

type ValidateFn func(*DbConn) bool

type HealthStats struct {
	HealthyAcquires int64 `json:"healthy_acquires"`
	StaleDiscarded  int64 `json:"stale_discarded"`
}

type HealthChecker struct {
	target          Manager
	validateFn      ValidateFn
	healthyAcquires int64
	staleDiscarded  int64
}

func NewHealthChecker(target Manager, validateFn ValidateFn) *HealthChecker {
	if validateFn == nil {
		validateFn = func(conn *DbConn) bool { return true }
	}
	return &HealthChecker{
		target:     target,
		validateFn: validateFn,
	}
}

func (h *HealthChecker) Acquire() (*DbConn, error) {
	for attempt := 0; attempt < 3; attempt++ {
		conn, err := h.target.Acquire()
		if err != nil {
			return nil, err
		}
		if h.validateFn(conn) {
			atomic.AddInt64(&h.healthyAcquires, 1)
			return conn, nil
		}
		// Invalid connection: discard and retry
		atomic.AddInt64(&h.staleDiscarded, 1)
		h.target.Release(conn)
	}
	return nil, errors.New("failed to acquire healthy connection after 3 attempts")
}

func (h *HealthChecker) Release(conn *DbConn) {
	h.target.Release(conn)
}

func (h *HealthChecker) IncrementQueries() {
	h.target.IncrementQueries()
}

func (h *HealthChecker) Stats() PoolStats {
	return h.target.Stats()
}

func (h *HealthChecker) HealthStats() HealthStats {
	return HealthStats{
		HealthyAcquires: atomic.LoadInt64(&h.healthyAcquires),
		StaleDiscarded:  atomic.LoadInt64(&h.staleDiscarded),
	}
}

func (h *HealthChecker) Dialect() string {
	return h.target.Dialect()
}

func (h *HealthChecker) Shutdown(ctx context.Context) error {
	return h.target.Shutdown(ctx)
}
```

#### Test Coverage Requirements (`health_checker_test.go`)
- **Healthy connection**: `ValidateFn` returns `true`. Connection acquired successfully on 1st attempt. `HealthyAcquires=1`, `StaleDiscarded=0`.
- **Stale/unhealthy connection retried**: `ValidateFn` returns `false` on 1st call, `true` on 2nd call. 1st connection is released, 2nd acquired. `HealthyAcquires=1`, `StaleDiscarded=1`.
- **All unhealthy exhaustion**: `ValidateFn` returns `false` continuously. Fails after 3 attempts, returns error. `HealthyAcquires=0`, `StaleDiscarded=3`.

---

### 2.3 SQ.G5 — W3C Trace Context Propagation (`packages/Pranor Pulse/pkg/tracing/traceparent.go` & `pkg/core/engine.go`)

#### Purpose
Format, parse, inject, and extract W3C Trace Context (`traceparent` header per W3C Recommendation) and record tracing info on `LogEntry` upon log append.

#### Design Specification (`traceparent.go`)
```go
package tracing

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

// Inject formats traceID and spanID into W3C traceparent header: 00-{traceID}-{spanID}-01
func Inject(headers map[string]string, traceID, spanID string) {
	if headers == nil {
		return
	}
	headers["traceparent"] = fmt.Sprintf("00-%s-%s-01", traceID, spanID)
}

// Extract parses traceparent header from headers map.
func Extract(headers map[string]string) (traceID, spanID string, sampled bool, ok bool) {
	if headers == nil {
		return "", "", false, false
	}

	var val string
	for k, v := range headers {
		if strings.EqualFold(k, "traceparent") {
			val = v
			break
		}
	}
	if val == "" {
		return "", "", false, false
	}

	parts := strings.Split(val, "-")
	if len(parts) != 4 {
		return "", "", false, false
	}

	version, tID, sID, flags := parts[0], parts[1], parts[2], parts[3]
	if version != "00" {
		return "", "", false, false
	}
	if len(tID) != 32 || tID == "00000000000000000000000000000000" {
		return "", "", false, false
	}
	if len(sID) != 16 || sID == "0000000000000000" {
		return "", "", false, false
	}
	if len(flags) != 2 {
		return "", "", false, false
	}

	if _, err := hex.DecodeString(tID); err != nil {
		return "", "", false, false
	}
	if _, err := hex.DecodeString(sID); err != nil {
		return "", "", false, false
	}

	flagBytes, err := hex.DecodeString(flags)
	if err != nil {
		return "", "", false, false
	}

	sampled = (flagBytes[0] & 0x01) != 0
	return tID, sID, sampled, true
}

func NewTraceID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func NewSpanID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
```

#### Engine Integration Specification (`pkg/core/engine.go`)
1. Extend `LogEntry`:
```go
type LogEntry struct {
	Offset      uint64 `json:"offset"`
	Topic       string `json:"topic"`
	Payload     string `json:"payload"`
	Timestamp   int64  `json:"timestamp"`
	Synced      bool   `json:"synced"`
	Traceparent string `json:"traceparent,omitempty"`
}
```
2. Add `Append` method to `Engine` accepting optional metadata:
```go
func (e *Engine) Append(topic, payload string, metadata ...map[string]string) (LogEntry, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	writePayload := payload
	if len(e.encryptionKey) == 32 {
		enc, err := EncryptPayload(payload, e.encryptionKey)
		if err != nil {
			return LogEntry{}, fmt.Errorf("engine: encryption failed: %w", err)
		}
		writePayload = "ENC:" + enc
	}

	entry, err := e.driver.Append(topic, writePayload)
	if err != nil {
		return LogEntry{}, err
	}

	if len(metadata) > 0 && metadata[0] != nil {
		var tp string
		for k, v := range metadata[0] {
			if strings.EqualFold(k, "traceparent") {
				tp = v
				break
			}
		}
		if tp != "" {
			entry.Traceparent = tp
			// Update in memory driver storage if applicable
			if md, ok := e.driver.(*MemoryDriver); ok {
				md.mu.Lock()
				if len(md.entries) > 0 {
					md.entries[len(md.entries)-1].Traceparent = tp
				}
				md.mu.Unlock()
			}
		}
	}

	return entry, nil
}
```

#### Test Coverage Requirements (`traceparent_test.go` & `engine_test.go`)
- `Inject`/`Extract` round-trip verification.
- Rejection of invalid headers (malformed version, invalid hex, invalid length, all zeros).
- Uniqueness and length verification for `NewTraceID` (32 hex characters) and `NewSpanID` (16 hex characters).
- Integration test: Calling `engine.Append(topic, payload, map[string]string{"traceparent": header})` stores `Traceparent` on the returned `LogEntry` and subsequent `Dequeue` calls preserve the `Traceparent`.

---

## 3. Component & File Mapping Matrix

| Requirement | Target File Path | Action | Key Dependencies |
|-------------|------------------|--------|------------------|
| **SP.G1** | `packages/Pranor Pool/pkg/routing/rw_splitter.go` | Create | Standard lib (`strings`, `sync/atomic`), `pkg/pool` |
| **SP.G1** | `packages/Pranor Pool/pkg/routing/rw_splitter_test.go` | Create | `testing`, `pkg/pool` |
| **SP.G2** | `packages/Pranor Pool/pkg/pool/health_checker.go` | Create | Standard lib (`sync/atomic`, `errors`, `fmt`, `context`), `pkg/pool` |
| **SP.G2** | `packages/Pranor Pool/pkg/pool/health_checker_test.go` | Create | `testing`, `pkg/pool` |
| **SQ.G5** | `packages/Pranor Pulse/pkg/tracing/traceparent.go` | Create | Standard lib (`crypto/rand`, `encoding/hex`, `fmt`, `strings`) |
| **SQ.G5** | `packages/Pranor Pulse/pkg/tracing/traceparent_test.go` | Create | `testing` |
| **SQ.G5** | `packages/Pranor Pulse/pkg/core/engine.go` | Modify | Update `LogEntry`, add `Append(...)` to `Engine` |
| **SQ.G5** | `packages/Pranor Pulse/pkg/core/engine_test.go` | Modify / Extend | Add trace context integration tests |

---

## 4. Test Suite Setup & Verification Plan

### Test Command Targets
- Pranor Pool: `cd /home/developer/workspace/pranor/packages/Pranor Pool && go test ./...`
- Pranor Pulse: `cd /home/developer/workspace/pranor/packages/Pranor Pulse && go test ./...`
- Build Check: `go build ./...` in both package directories.

### Verification Conditions
1. Exit code 0 for both `go build ./...` and `go test ./...`.
2. No new external dependencies added in `go.mod`.
3. Each new package has at least 3 test functions.
4. No test uses `t.Skip()`.
