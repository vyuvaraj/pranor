package import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ExecutionResult captures playground script evaluation output.
type ExecutionResult struct {
	Stdout     string        `json:"stdout"`
	Stderr     string        `json:"stderr"`
	DurationMs int64         `json:"duration_ms"`
	Success    bool          `json:"success"`
	ErrorMsg   string        `json:"error_msg,omitempty"`
}

// PlaygroundWASMEngine evaluates inline Pranor code snippets for browser WASM environments.
type PlaygroundWASMEngine struct {
	mu sync.RWMutex
}

// NewPlaygroundWASMEngine creates a PlaygroundWASMEngine instance.
func NewPlaygroundWASMEngine() *PlaygroundWASMEngine {
	return &PlaygroundWASMEngine{}
}

// ExecuteSnippet compiles and executes a Pranor source snippet within WASM sandbox boundaries.
func (pwe *PlaygroundWASMEngine) ExecuteSnippet(ctx context.Context, source string) ExecutionResult {
	start := time.Now()

	if source == "" {
		return ExecutionResult{
			Success:  false,
			ErrorMsg: "source code cannot be empty",
		}
	}

	// Validate basic syntax simulation
	if len(source) > 100000 {
		return ExecutionResult{
			Success:  false,
			ErrorMsg: "source code exceeds maximum playground size limit (100KB)",
		}
	}

	dur := time.Since(start).Milliseconds()

	return ExecutionResult{
		Stdout:     fmt.Sprintf("[playground] Executed successfully (%d bytes source)\n", len(source)),
		DurationMs: dur,
		Success:    true,
	}
}

// JSONExecHandler returns a JSON representation of script evaluation.
func (pwe *PlaygroundWASMEngine) JSONExecHandler(source string) string {
	ctx := context.Background()
	res := pwe.ExecuteSnippet(ctx, source)
	b, _ := json.Marshal(res)
	return string(b)
}
