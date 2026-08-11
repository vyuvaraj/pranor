package api

import (
	"context"
	"errors"
	"time"

	"github.com/vyuvaraj/pranor/core/pkg/execctx"
)

// WorkingMemory provides volatile, in-session scratchpad storage.
type WorkingMemory interface {
	Set(ctx context.Context, ec *execctx.ExecutionContext, sessionID, key string, value any) error
	Get(ctx context.Context, ec *execctx.ExecutionContext, sessionID, key string) (any, bool, error)
	Delete(ctx context.Context, ec *execctx.ExecutionContext, sessionID, key string) error
	Flush(ctx context.Context, ec *execctx.ExecutionContext, sessionID string) error
}

// MemoryEntry represents a single stored episodic memory item.
type MemoryEntry struct {
	ID        string
	TenantID  string
	AgentID   string
	SessionID string
	Content   string
	Role      string            // user, assistant, tool
	Tags      []string
	CreatedAt time.Time
	Score     float64           // computed relevance score during recall
}

// EpisodicMemory provides cross-session structured memory recall.
type EpisodicMemory interface {
	StoreEpisode(ctx context.Context, ec *execctx.ExecutionContext, sessionID, role, content string, tags []string) (MemoryEntry, error)
	Recall(ctx context.Context, ec *execctx.ExecutionContext, query string, topK int) ([]MemoryEntry, error)
	Purge(ctx context.Context, ec *execctx.ExecutionContext) error
}

// MemoryEngine combines Working and Episodic memory capabilities.
type MemoryEngine interface {
	Working() WorkingMemory
	Episodic() EpisodicMemory
}

// Sentinel errors
var (
	ErrKeyNotFound     = errors.New("pranor/memory: key not found")
	ErrInvalidMemory   = errors.New("pranor/memory: invalid memory parameter")
	ErrEERequired      = errors.New("pranor/memory: vector HNSW semantic memory recall requires Enterprise Edition")
)
