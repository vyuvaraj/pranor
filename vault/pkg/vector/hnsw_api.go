package import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// VectorStore manages per-bucket HNSW vector index namespaces.
// Each bucket gets its own isolated Index instance.
type VectorStore struct {
	mu         sync.RWMutex
	indexes    map[string]*Index
	defaultCfg Config
}

// NewVectorStore creates a VectorStore with a default Config applied to each new bucket.
func NewVectorStore(cfg Config) *VectorStore {
	return &VectorStore{
		indexes:    make(map[string]*Index),
		defaultCfg: cfg,
	}
}

// indexFor returns (or lazily creates) the Index for the given bucket.
func (vs *VectorStore) indexFor(bucket string) *Index {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	if idx, ok := vs.indexes[bucket]; ok {
		return idx
	}
	idx := NewIndex(vs.defaultCfg)
	vs.indexes[bucket] = idx
	return idx
}

// UpsertVector inserts or replaces a vector in the given bucket's index.
func (vs *VectorStore) UpsertVector(bucket string, id int, vec []float32) error {
	return vs.indexFor(bucket).Insert(id, vec)
}

// SearchBucket runs a top-k ANN search in the given bucket's index.
// Results with score < minScore are filtered out.
func (vs *VectorStore) SearchBucket(bucket string, query []float32, k int, minScore float32) ([]SearchResult, error) {
	vs.mu.RLock()
	idx, ok := vs.indexes[bucket]
	vs.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("bucket %q has no vector index", bucket)
	}
	results, err := idx.Search(query, k)
	if err != nil {
		return nil, err
	}
	if minScore <= 0 {
		return results, nil
	}
	filtered := results[:0]
	for _, r := range results {
		if r.Score >= minScore {
			filtered = append(filtered, r)
		}
	}
	return filtered, nil
}

// Stats returns basic stats about the bucket's index.
func (vs *VectorStore) Stats(bucket string) (count int, dim int, exists bool) {
	vs.mu.RLock()
	idx, ok := vs.indexes[bucket]
	vs.mu.RUnlock()
	if !ok {
		return 0, 0, false
	}
	return idx.Len(), idx.cfg.Dim, true
}

// Handler returns an http.Handler exposing the VectorStore HTTP API:
//
//	POST   /api/v1/vectors/{bucket}         - upsert a vector
//	POST   /api/v1/vectors/{bucket}/search  - search nearest neighbors
//	GET    /api/v1/vectors/{bucket}/stats   - index stats
func (vs *VectorStore) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/vectors/", func(w http.ResponseWriter, r *http.Request) {
		// Parse: /api/v1/vectors/{bucket}[/search|/stats]
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/vectors/")
		parts := strings.SplitN(path, "/", 2)
		bucket := parts[0]
		suffix := ""
		if len(parts) == 2 {
			suffix = parts[1]
		}

		if bucket == "" {
			http.Error(w, "bucket name required", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		switch {
		case suffix == "search" && r.Method == http.MethodPost:
			vs.handleSearch(w, r, bucket)
		case suffix == "stats" && r.Method == http.MethodGet:
			vs.handleStats(w, r, bucket)
		case suffix == "" && r.Method == http.MethodPost:
			vs.handleUpsert(w, r, bucket)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	})
	return mux
}

type upsertRequest struct {
	ID     int       `json:"id"`
	Vector []float32 `json:"vector"`
}

type searchRequest struct {
	Vector   []float32 `json:"vector"`
	K        int       `json:"k"`
	MinScore float32   `json:"min_score"`
}

func (vs *VectorStore) handleUpsert(w http.ResponseWriter, r *http.Request, bucket string) {
	var req upsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if len(req.Vector) == 0 {
		http.Error(w, `{"error":"vector is required"}`, http.StatusBadRequest)
		return
	}

	// If the bucket doesn't have an index yet, create one with the vector's dimension
	vs.mu.Lock()
	if _, exists := vs.indexes[bucket]; !exists {
		cfg := vs.defaultCfg
		cfg.Dim = len(req.Vector)
		vs.indexes[bucket] = NewIndex(cfg)
	}
	vs.mu.Unlock()

	if err := vs.UpsertVector(bucket, req.ID, req.Vector); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusUnprocessableEntity)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": req.ID, "bucket": bucket, "ok": true})
}

func (vs *VectorStore) handleSearch(w http.ResponseWriter, r *http.Request, bucket string) {
	var req searchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if len(req.Vector) == 0 {
		http.Error(w, `{"error":"vector is required"}`, http.StatusBadRequest)
		return
	}
	k := req.K
	if k <= 0 {
		k = 10
	}

	results, err := vs.SearchBucket(bucket, req.Vector, k, req.MinScore)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusNotFound)
		return
	}
	if results == nil {
		results = []SearchResult{}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"results": results, "count": len(results)})
}

func (vs *VectorStore) handleStats(w http.ResponseWriter, r *http.Request, bucket string) {
	count, dim, exists := vs.Stats(bucket)
	if !exists {
		http.Error(w, fmt.Sprintf(`{"error":"bucket %q not found"}`, bucket), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"bucket": bucket,
		"count":  strconv.Itoa(count),
		"dim":    dim,
	})
}
