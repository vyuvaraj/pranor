package import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type CachedFile struct {
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	LastSync    time.Time `json:"last_sync"`
	SyncedRemote bool     `json:"synced_remote"`
}

type BrowserStoreSDK struct {
	cache map[string]CachedFile
	mu    sync.RWMutex
}

func NewBrowserStoreSDK() *BrowserStoreSDK {
	return &BrowserStoreSDK{
		cache: make(map[string]CachedFile),
	}
}

func (b *BrowserStoreSDK) SaveToOPFS(ctx context.Context, path string, data []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.cache[path] = CachedFile{
		Path:        path,
		Size:        int64(len(data)),
		LastSync:    time.Now(),
		SyncedRemote: false,
	}
	return nil
}

func (b *BrowserStoreSDK) SyncToPranorVault(ctx context.Context, path string) ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	f, exists := b.cache[path]
	if !exists {
		return nil, fmt.Errorf("file %s not found in browser OPFS cache", path)
	}

	f.SyncedRemote = true
	b.cache[path] = f

	respMap := map[string]interface{}{
		"status":      "SYNCED",
		"opfs_path":   path,
		"size_bytes":  f.Size,
		"synced_time": time.Now(),
	}
	return json.Marshal(respMap)
}
