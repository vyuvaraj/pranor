package import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type WASMFilterAction int

const (
	ActionAllow WASMFilterAction = iota
	ActionBlock
	ActionModifyHeaders
)

type WASMEdgeFilter struct {
	Name    string
	Code    []byte
	Actions WASMFilterAction
}

type WASMEdgeEngine struct {
	filters map[string]*WASMEdgeFilter
	mu      sync.RWMutex
}

func NewWASMEdgeEngine() *WASMEdgeEngine {
	return &WASMEdgeEngine{
		filters: make(map[string]*WASMEdgeFilter),
	}
}

func (e *WASMEdgeEngine) RegisterFilter(name string, code []byte) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.filters[name] = &WASMEdgeFilter{
		Name:    name,
		Code:    code,
		Actions: ActionModifyHeaders,
	}
}

func (e *WASMEdgeEngine) ProcessRequest(ctx context.Context, r *http.Request) (WASMFilterAction, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Simulate WASM filter evaluation
	authHeader := r.Header.Get("Authorization")
	if strings.Contains(authHeader, "blocked") {
		return ActionBlock, fmt.Errorf("wasm edge filter: authorization token blocked by rule")
	}

	r.Header.Set("X-WASM-Edge-Processed", "true")
	r.Header.Set("X-PranorGateway-Engine", "wasmtime-edge-v2")
	return ActionAllow, nil
}
