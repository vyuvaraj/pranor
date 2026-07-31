package import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type ClientRoute struct {
	Path     string `json:"path"`
	Fallback string `json:"fallback"`
	Offline  bool   `json:"offline"`
}

type BrowserGatewaySDK struct {
	routes map[string]ClientRoute
	mu     sync.RWMutex
}

func NewBrowserGatewaySDK() *BrowserGatewaySDK {
	return &BrowserGatewaySDK{
		routes: make(map[string]ClientRoute),
	}
}

func (b *BrowserGatewaySDK) RegisterClientRoute(path, fallback string, offline bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.routes[path] = ClientRoute{
		Path:     path,
		Fallback: fallback,
		Offline:  offline,
	}
}

func (b *BrowserGatewaySDK) InterceptFetch(ctx context.Context, path string) ([]byte, int, error) {
	b.mu.RLock()
	route, exists := b.routes[path]
	b.mu.RUnlock()

	if !exists {
		return nil, http.StatusNotFound, fmt.Errorf("route %s not registered in browser gateway", path)
	}

	respMap := map[string]interface{}{
		"pranorGate_edge": "browser-wasm",
		"intercepted_path": route.Path,
		"fallback_target":  route.Fallback,
		"offline_capable":  route.Offline,
	}

	data, _ := json.Marshal(respMap)
	return data, http.StatusOK, nil
}
