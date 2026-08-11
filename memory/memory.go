package memory

import (
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
