package eval

import (
	"context"

	"github.com/vyuvaraj/pranor/eval/api"
	"github.com/vyuvaraj/pranor/eval/pkg/engine"
)

// DefaultEngine is the package-level EvalEngine. Set by init() in oss/ee file.
var DefaultEngine api.EvalEngine = engine.NewOSSEngine()

// Register adds an Evaluator to DefaultEngine.
func Register(e api.Evaluator) {
	DefaultEngine.Register(e)
}

// Run evaluates a Trajectory through all registered evaluators.
func Run(ctx context.Context, t api.Trajectory) (api.EvalResult, error) {
	return DefaultEngine.Run(ctx, t)
}

// Replay re-emits a trajectory's spans in order for analysis.
func Replay(ctx context.Context, t api.Trajectory) (api.Trajectory, error) {
	return DefaultEngine.Replay(ctx, t)
}
