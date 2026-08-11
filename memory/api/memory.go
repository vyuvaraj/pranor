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
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	AgentID   string    `json:"agent_id"`
	SessionID string    `json:"session_id"`
	Content   string    `json:"content"`
	Role      string    `json:"role"` // "user", "assistant", "tool"
	Tags      []string  `json:"tags"`
	Vector    []float32 `json:"vector,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Score     float64   `json:"score"` // Computed relevance score during recall
}

// EpisodicMemory provides cross-session structured memory recall.
type EpisodicMemory interface {
	StoreEpisode(ctx context.Context, ec *execctx.ExecutionContext, sessionID, role, content string, tags []string) (MemoryEntry, error)
	StoreEpisodeWithVector(ctx context.Context, ec *execctx.ExecutionContext, sessionID, role, content string, tags []string, vector []float32) (MemoryEntry, error)
	Recall(ctx context.Context, ec *execctx.ExecutionContext, query string, topK int) ([]MemoryEntry, error)
	RecallSemantic(ctx context.Context, ec *execctx.ExecutionContext, queryVector []float32, topK int) ([]MemoryEntry, error)
	Purge(ctx context.Context, ec *execctx.ExecutionContext) error
}

// MemoryEngine combines Working and Episodic memory capabilities.
type MemoryEngine interface {
	Working() WorkingMemory
	Episodic() EpisodicMemory
}

// Sentinel errors
var (
	ErrKeyNotFound   = errors.New("pranor/memory: key not found")
	ErrInvalidMemory = errors.New("pranor/memory: invalid memory parameter")
	ErrEERequired    = errors.New("pranor/memory: vector HNSW semantic memory recall requires Enterprise Edition")
)
