//go:build !enterprise

package decision

import (
	"github.com/vyuvaraj/pranor/decision/pkg/engine"
)

func init() {
	// Initialize default engine for OSS
	DefaultEngine = engine.NewVetoLadderEngine(GraphProvider)
}
