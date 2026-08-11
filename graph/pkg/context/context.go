package graphctx

import (
	"context"

	"github.com/vyuvaraj/pranor/graph/api"
)

type EntityContext struct {
	EntityID  string
	TenantID  string
	AgentID   string
	UserID    string
	Tier      string
	Data      map[string]any
	LatencyMs int64
	CacheHit  bool
}

type Assembler interface {
	Assemble(ctx context.Context, q api.ContextQuery) (api.ContextResult, error)
}
