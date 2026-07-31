package import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// SearchableResource represents an ecosystem item indexed for ⌘K global search.
type SearchableResource struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"` // "service", "route", "queue", "bucket", "workflow", "trace"
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	URL         string            `json:"url"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// UnifiedGlobalSearchEngine indexes all Pranor resources for fast prefix-matching (⌘K shortcut).
type UnifiedGlobalSearchEngine struct {
	mu        sync.RWMutex
	resources map[string]*SearchableResource // ID -> resource
}

// NewUnifiedGlobalSearchEngine creates a UnifiedGlobalSearchEngine instance.
func NewUnifiedGlobalSearchEngine() *UnifiedGlobalSearchEngine {
	return &UnifiedGlobalSearchEngine{
		resources: make(map[string]*SearchableResource),
	}
}

// IndexResource indexes or updates a resource in the global search index.
func (ugse *UnifiedGlobalSearchEngine) IndexResource(res SearchableResource) {
	if res.ID == "" {
		return
	}
	ugse.mu.Lock()
	defer ugse.mu.Unlock()
	ugse.resources[res.ID] = &res
}

// Search performs multi-field prefix matching across indexed Pranor resources.
func (ugse *UnifiedGlobalSearchEngine) Search(query string, maxResults int) []SearchableResource {
	if maxResults <= 0 {
		maxResults = 20
	}
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return []SearchableResource{}
	}

	ugse.mu.RLock()
	defer ugse.mu.RUnlock()

	var matches []SearchableResource

	for _, res := range ugse.resources {
		if strings.Contains(strings.ToLower(res.Title), q) ||
			strings.Contains(strings.ToLower(res.ID), q) ||
			strings.Contains(strings.ToLower(res.Type), q) ||
			strings.Contains(strings.ToLower(res.Description), q) {
			matches = append(matches, *res)
			if len(matches) >= maxResults {
				break
			}
		}
	}

	return matches
}

// HTTPHandler exposes `/api/v1/console/search` for ⌘K search bar input.
func (ugse *UnifiedGlobalSearchEngine) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		results := ugse.Search(q, 20)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"query":   q,
			"count":   len(results),
			"results": results,
		})
	})
}
