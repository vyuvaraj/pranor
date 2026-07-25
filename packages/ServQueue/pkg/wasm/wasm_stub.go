//go:build !js || !wasm

package wasm

import "fmt"

// Stub main for non-WASM targets (e.g. linux/amd64, windows, darwin)
func main() {
	fmt.Println("ServQueue WASM Bridge: Native build stub. Compile with GOOS=js GOARCH=wasm for WebAssembly browser target.")
}
