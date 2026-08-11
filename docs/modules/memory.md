# Memory Engine (`std/memory`)

**Module Path:** `github.com/vyuvaraj/pranor/memory`  
**Introduced:** Phase 92 (Sprint V2.92.1)

---

## Overview

Pranor Memory (`std/memory`) provides governed working and episodic memory for AI agents, operating without external database dependencies.

- **Working Memory:** Volatile, in-session scratchpad scoped to `(TenantID, AgentID, SessionID)`.
- **Episodic Memory:** Cross-session memory recall storing conversation turns and tool outputs with time-decay and keyword relevance scoring algorithms.

---

## Key Interfaces

```go
type WorkingMemory interface {
	Set(ctx context.Context, ec *execctx.ExecutionContext, sessionID, key string, value any) error
	Get(ctx context.Context, ec *execctx.ExecutionContext, sessionID, key string) (any, bool, error)
	Delete(ctx context.Context, ec *execctx.ExecutionContext, sessionID, key string) error
	Flush(ctx context.Context, ec *execctx.ExecutionContext, sessionID string) error
}

type EpisodicMemory interface {
	StoreEpisode(ctx context.Context, ec *execctx.ExecutionContext, sessionID, role, content string, tags []string) (MemoryEntry, error)
	Recall(ctx context.Context, ec *execctx.ExecutionContext, query string, topK int) ([]MemoryEntry, error)
	Purge(ctx context.Context, ec *execctx.ExecutionContext) error
}
```

---

## Data Structures

```go
type MemoryEntry struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	AgentID   string    `json:"agent_id"`
	SessionID string    `json:"session_id"`
	Content   string    `json:"content"`
	Role      string    `json:"role"` // "user", "assistant", "tool"
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	Score     float64   `json:"score"` // Computed relevance score during recall
}
```

---

## Relevance & Time-Decay Scoring Algorithm

During `Recall(ctx, ec, query, topK)`, memory entries are filtered by `ec.TenantID` and `ec.AgentID` (ensuring tenant isolation), then scored:

$$\text{Score} = \text{KeywordMatchCount} \times \left( \frac{1.0}{1.0 + \text{HoursSinceCreation}} \right)$$

Entries are returned sorted by `Score` descending.

---

## Code Example

```go
import (
	"github.com/vyuvaraj/pranor/memory"
)

// Working Memory Scratchpad
wm := memory.Working()
_ = wm.Set(ctx, ec, sessionID, "current_step", "parsing_invoice")

// Episodic Recall
em := memory.Episodic()
entries, _ := em.Recall(ctx, ec, "invoice payment", 5)
```
