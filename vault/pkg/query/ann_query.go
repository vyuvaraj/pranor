package import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"sync"
)

// ANNQueryRequest defines parameters for approximate nearest neighbor vector search.
type ANNQueryRequest struct {
	BucketName string                 `json:"bucket_name"`
	Vector     []float32              `json:"vector"`
	TopK       int                    `json:"top_k"`
	MinScore   float64                `json:"min_score"` // e.g. 0.75
	Filter     map[string]interface{} `json:"filter,omitempty"`
}

// VectorMatchItem represents a matching object with similarity score.
type VectorMatchItem struct {
	ObjectID string                 `json:"object_id"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ANNQueryEngine executes ANN vector query evaluations against bucket index nodes.
type ANNQueryEngine struct {
	mu     sync.RWMutex
	nodes  map[string]map[string][]float32              // bucket -> (objectID -> vector)
	meta   map[string]map[string]map[string]interface{} // bucket -> (objectID -> metadata)
}

// NewANNQueryEngine creates an ANNQueryEngine instance.
func NewANNQueryEngine() *ANNQueryEngine {
	return &ANNQueryEngine{
		nodes: make(map[string]map[string][]float32),
		meta:  make(map[string]map[string]map[string]interface{}),
	}
}

// InsertVector index object vector and metadata.
func (aqe *ANNQueryEngine) InsertVector(bucket, objectID string, vec []float32, metadata map[string]interface{}) {
	aqe.mu.Lock()
	defer aqe.mu.Unlock()

	if _, ok := aqe.nodes[bucket]; !ok {
		aqe.nodes[bucket] = make(map[string][]float32)
		aqe.meta[bucket] = make(map[string]map[string]interface{})
	}

	aqe.nodes[bucket][objectID] = vec
	aqe.meta[bucket][objectID] = metadata
}

// QueryANN evaluates nearest neighbor vectors matching threshold and metadata filters.
func (aqe *ANNQueryEngine) QueryANN(req ANNQueryRequest) ([]VectorMatchItem, error) {
	if req.TopK <= 0 {
		req.TopK = 10
	}

	aqe.mu.RLock()
	bucketNodes, ok := aqe.nodes[req.BucketName]
	bucketMeta := aqe.meta[req.BucketName]
	aqe.mu.RUnlock()

	if !ok {
		return []VectorMatchItem{}, nil
	}

	var matches []VectorMatchItem

	for objID, vec := range bucketNodes {
		// Evaluate metadata filter if specified
		if len(req.Filter) > 0 {
			meta := bucketMeta[objID]
			if !matchMetadataFilter(meta, req.Filter) {
				continue
			}
		}

		score := cosineSimilarity(req.Vector, vec)
		if score >= req.MinScore {
			matches = append(matches, VectorMatchItem{
				ObjectID: objID,
				Score:    score,
				Metadata: bucketMeta[objID],
			})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	if len(matches) > req.TopK {
		matches = matches[:req.TopK]
	}

	return matches, nil
}

func cosineSimilarity(v1, v2 []float32) float64 {
	if len(v1) != len(v2) || len(v1) == 0 {
		return 0.0
	}
	var dot, n1, n2 float64
	for i := range v1 {
		dot += float64(v1[i] * v2[i])
		n1 += float64(v1[i] * v1[i])
		n2 += float64(v2[i] * v2[i])
	}
	if n1 == 0 || n2 == 0 {
		return 0.0
	}
	return dot / (math.Sqrt(n1) * math.Sqrt(n2))
}

func matchMetadataFilter(meta, filter map[string]interface{}) bool {
	if meta == nil {
		return false
	}
	for k, targetVal := range filter {
		actualVal, exists := meta[k]
		if !exists || fmt.Sprintf("%v", actualVal) != fmt.Sprintf("%v", targetVal) {
			return false
		}
	}
	return true
}

// HTTPHandler exposes REST endpoint `/api/v1/store/vector/query`.
func (aqe *ANNQueryEngine) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req ANNQueryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}

		matches, err := aqe.QueryANN(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"count":   len(matches),
			"matches": matches,
		})
	})
}
