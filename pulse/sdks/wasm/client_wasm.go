//go:build js && wasm

package wasm

import (
	"syscall/js"
)

func main() {
	c := make(chan struct{})
	js.Global().Set("PranorPulseWasmClient", js.FuncOf(func(this js.Value, args []js.Value) any {
		return map[string]any{
			"status": "ready",
			"driver": "opfs-wasm-embedded",
		}
	}))
	<-c
}
