package wasm

import (
	"context"

	"github.com/tetratelabs/wazero"
)

type WASMPredictor struct {
	runtime wazero.Runtime
}

func NewWASMPredictor() *WASMPredictor {
	return &WASMPredictor{
		runtime: wazero.NewRuntime(context.Background()),
	}
}
