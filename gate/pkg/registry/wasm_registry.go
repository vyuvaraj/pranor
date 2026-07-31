package import (
	"fmt"
	"sync"
	"time"
)

type WASMPlugin struct {
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Version   string    `json:"version"`
	LoadedAt  time.Time `json:"loaded_at"`
	ByteCode  []byte    `json:"-"`
}

type WASMHotReloadRegistry struct {
	plugins map[string]*WASMPlugin
	mu      sync.RWMutex
}

func NewWASMHotReloadRegistry() *WASMHotReloadRegistry {
	return &WASMHotReloadRegistry{
		plugins: make(map[string]*WASMPlugin),
	}
}

func (r *WASMHotReloadRegistry) RegisterPlugin(name, url, version string, bytecode []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if name == "" || url == "" {
		return fmt.Errorf("wasm registry: missing plugin name or URL")
	}

	r.plugins[name] = &WASMPlugin{
		Name:     name,
		URL:      url,
		Version:  version,
		LoadedAt: time.Now(),
		ByteCode: bytecode,
	}
	return nil
}

func (r *WASMHotReloadRegistry) GetPlugin(name string) (*WASMPlugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.plugins[name]
	return p, exists
}
