package engine

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// WASMStepRunner executes inline WebAssembly bytecode for ServFlow task nodes.
type WASMStepRunner struct {
	mu      sync.RWMutex
	runtime wazero.Runtime
}

// NewWASMStepRunner initializes a wazero WASM runtime for ServFlow.
func NewWASMStepRunner(ctx context.Context) (*WASMStepRunner, error) {
	r := wazero.NewRuntime(ctx)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}
	return &WASMStepRunner{runtime: r}, nil
}

// Close releases the WASM runtime resources.
func (wsr *WASMStepRunner) Close(ctx context.Context) error {
	wsr.mu.Lock()
	defer wsr.mu.Unlock()
	if wsr.runtime != nil {
		return wsr.runtime.Close(ctx)
	}
	return nil
}

// ExecuteWASM runs a compiled WASM module and executes the named function with string input.
func (wsr *WASMStepRunner) ExecuteWASM(ctx context.Context, wasmBytes []byte, funcName string) ([]byte, error) {
	if len(wasmBytes) == 0 {
		return nil, errors.New("WASM bytecode cannot be empty")
	}

	wsr.mu.RLock()
	r := wsr.runtime
	wsr.mu.RUnlock()

	if r == nil {
		return nil, errors.New("WASM runtime is closed")
	}

	mod, err := r.Instantiate(ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate WASM module: %w", err)
	}
	defer mod.Close(ctx)

	fn := mod.ExportedFunction(funcName)
	if fn == nil {
		return nil, fmt.Errorf("exported function '%s' not found in WASM module", funcName)
	}

	results, err := fn.Call(ctx)
	if err != nil {
		return nil, fmt.Errorf("WASM function execution error: %w", err)
	}

	if len(results) > 0 {
		return []byte(fmt.Sprintf("%v", results[0])), nil
	}
	return []byte("success"), nil
}
