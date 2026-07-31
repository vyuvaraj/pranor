//go:build js && wasm

package wasm

import (
	"encoding/json"
	"syscall/js"

	"github.com/vyuvaraj/pranor/pulse/pkg/core"
	"github.com/vyuvaraj/pranor/pulse/pkg/opfs"
	"github.com/vyuvaraj/pranor/pulse/pkg/relay"
)

var (
	globalEngine *core.Engine
	globalRelay  *relay.OutboxRelay
)

func main() {
	driver, err := opfs.NewOPFSDriver("")
	if err != nil {
		driver = nil
	}
	globalEngine = core.NewEngine(driver)

	// Register JS global functions for Pranor Pulse Embedded
	js.Global().Set("PranorPulseEnqueue", js.FuncOf(jsEnqueue))
	js.Global().Set("PranorPulseDequeue", js.FuncOf(jsDequeue))
	js.Global().Set("PranorPulseSyncOutbox", js.FuncOf(jsSyncOutbox))
	js.Global().Set("PranorPulseGetPendingSync", js.FuncOf(jsGetPendingSync))

	// Keep WASM process alive in browser event loop
	select {}
}

func jsEnqueue(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf(map[string]interface{}{"error": "Usage: PranorPulseEnqueue(topic, payload)"})
	}
	topic := args[0].String()
	payload := args[1].String()

	entry, err := globalEngine.Enqueue(topic, payload)
	if err != nil {
		return js.ValueOf(map[string]interface{}{"error": err.Error()})
	}

	data, _ := json.Marshal(entry)
	return js.ValueOf(string(data))
}

func jsDequeue(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		return js.ValueOf(map[string]interface{}{"error": "Usage: PranorPulseDequeue(topic, startOffset, limit)"})
	}
	topic := args[0].String()
	startOffset := uint64(args[1].Int())
	limit := uint64(args[2].Int())

	entries, err := globalEngine.Dequeue(topic, startOffset, limit)
	if err != nil {
		return js.ValueOf(map[string]interface{}{"error": err.Error()})
	}

	data, _ := json.Marshal(entries)
	return js.ValueOf(string(data))
}

func jsSyncOutbox(this js.Value, args []js.Value) interface{} {
	serverURL := "http://localhost:8080/api/v1/queue/stream"
	if len(args) >= 1 && args[0].String() != "" {
		serverURL = args[0].String()
	}

	if globalRelay == nil {
		globalRelay = relay.NewOutboxRelay(globalEngine, serverURL)
	}

	err := globalRelay.SyncNow()
	if err != nil {
		return js.ValueOf(map[string]interface{}{"error": err.Error()})
	}
	return js.ValueOf(map[string]interface{}{"status": "success"})
}

func jsGetPendingSync(this js.Value, args []js.Value) interface{} {
	limit := uint64(50)
	if len(args) >= 1 {
		limit = uint64(args[0].Int())
	}

	entries, err := globalEngine.GetPendingSync(limit)
	if err != nil {
		return js.ValueOf(map[string]interface{}{"error": err.Error()})
	}

	data, _ := json.Marshal(entries)
	return js.ValueOf(string(data))
}
