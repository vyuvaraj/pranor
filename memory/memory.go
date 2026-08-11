package memory

import (
	"context"
	"github.com/vyuvaraj/pranor/core/pkg/execctx"
	"github.com/vyuvaraj/pranor/memory/api"
	"github.com/vyuvaraj/pranor/memory/pkg/engine"
)

var DefaultEngine api.MemoryEngine = engine.NewOSSEngine()

func Working() api.WorkingMemory {
	return DefaultEngine.Working()
}

func Episodic() api.EpisodicMemory {
	return DefaultEngine.Episodic()
}

func RecallSemantic(ctx context.Context, ec *execctx.ExecutionContext, queryVector []float32, topK int) ([]api.MemoryEntry, error) {
	return DefaultEngine.Episodic().RecallSemantic(ctx, ec, queryVector, topK)
}
