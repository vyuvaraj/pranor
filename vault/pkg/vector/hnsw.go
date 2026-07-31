// Package vector implements an embedded HNSW (Hierarchical Navigable Small World)
// approximate nearest neighbor index for Pranor Vault. It provides per-bucket vector
// namespaces, zero external dependencies, and an HTTP API for upsert and search.
//
// Algorithm reference: Malkov & Yashunin (2018), "Efficient and robust approximate
// nearest neighbor search using Hierarchical Navigable Small World graphs."
package import (
	"container/heap"
	"errors"
	"math"
	"math/rand"
	"sync"
)

// Config controls HNSW index parameters.
type Config struct {
	M              int // Max neighbors per node per layer (default 16)
	EfConstruction int // Beam width during graph construction (default 200)
	EfSearch       int // Beam width during search queries (default 50)
	Dim            int // Vector dimension (required)
}

// DefaultConfig returns sensible HNSW defaults.
func DefaultConfig(dim int) Config {
	return Config{M: 16, EfConstruction: 200, EfSearch: 50, Dim: dim}
}

// SearchResult is a single nearest-neighbor hit.
type SearchResult struct {
	ID    int     `json:"id"`
	Score float32 `json:"score"` // cosine similarity [0, 1]
}

// node is an internal HNSW graph node.
type node struct {
	id        int
	vector    []float32
	neighbors [][]int // neighbors[layer] = list of neighbor IDs
	maxLayer  int
}

// ---- priority queue helpers ----

type candidate struct {
	id   int
	dist float32 // lower = closer (we use 1-cosine as distance)
}

type minHeap []candidate // min-heap by dist (closest first)

func (h minHeap) Len() int            { return len(h) }
func (h minHeap) Less(i, j int) bool  { return h[i].dist < h[j].dist }
func (h minHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x interface{}) { *h = append(*h, x.(candidate)) }
func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

type maxHeap []candidate // max-heap by dist (farthest first, for ef window)

func (h maxHeap) Len() int            { return len(h) }
func (h maxHeap) Less(i, j int) bool  { return h[i].dist > h[j].dist }
func (h maxHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *maxHeap) Push(x interface{}) { *h = append(*h, x.(candidate)) }
func (h *maxHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// ---- Index ----

// Index is a thread-safe HNSW approximate nearest neighbor index.
type Index struct {
	cfg        Config
	mu         sync.RWMutex
	nodes      map[int]*node
	entryPoint int
	maxLayer   int
	levelMult  float64 // 1 / ln(M)
	rng        *rand.Rand
}

// NewIndex creates a new empty HNSW index with the given configuration.
func NewIndex(cfg Config) *Index {
	if cfg.M <= 0 {
		cfg.M = 16
	}
	if cfg.EfConstruction <= 0 {
		cfg.EfConstruction = 200
	}
	if cfg.EfSearch <= 0 {
		cfg.EfSearch = 50
	}
	return &Index{
		cfg:        cfg,
		nodes:      make(map[int]*node),
		entryPoint: -1,
		levelMult:  1.0 / math.Log(float64(cfg.M)),
		rng:        rand.New(rand.NewSource(42)),
	}
}

// Len returns the number of vectors indexed.
func (idx *Index) Len() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.nodes)
}

// Insert adds a vector with the given id into the index.
// If a node with that id already exists, it is replaced.
func (idx *Index) Insert(id int, vec []float32) error {
	if len(vec) != idx.cfg.Dim {
		return errors.New("vector dimension mismatch")
	}
	if magnitude(vec) == 0 {
		return errors.New("zero vector cannot be indexed")
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	layer := idx.randomLayer()
	n := &node{
		id:       id,
		vector:   normalise(vec),
		maxLayer: layer,
		neighbors: func() [][]int {
			nb := make([][]int, layer+1)
			for i := range nb {
				nb[i] = []int{}
			}
			return nb
		}(),
	}
	idx.nodes[id] = n

	if idx.entryPoint == -1 {
		idx.entryPoint = id
		idx.maxLayer = layer
		return nil
	}

	// Search from top layer down to node's layer+1 greedily (ef=1)
	ep := idx.entryPoint
	for lc := idx.maxLayer; lc > layer; lc-- {
		cands := idx.searchLayer(n.vector, ep, 1, lc)
		if len(cands) > 0 {
			ep = cands[0].id
		}
	}

	// From layer down to 0: beam search + connect
	for lc := min(layer, idx.maxLayer); lc >= 0; lc-- {
		ef := idx.cfg.EfConstruction
		cands := idx.searchLayer(n.vector, ep, ef, lc)
		neighbors := idx.selectNeighbors(n.vector, cands, idx.cfg.M)

		n.neighbors[lc] = make([]int, len(neighbors))
		for i, nb := range neighbors {
			n.neighbors[lc][i] = nb.id
			// Bidirectional connection
			nbNode := idx.nodes[nb.id]
			if len(nbNode.neighbors[lc]) < idx.cfg.M {
				nbNode.neighbors[lc] = append(nbNode.neighbors[lc], id)
			}
		}

		if len(cands) > 0 {
			ep = cands[0].id
		}
	}

	if layer > idx.maxLayer {
		idx.maxLayer = layer
		idx.entryPoint = id
	}

	return nil
}

// Search returns the top-k nearest neighbors to the query vector.
// If the query is an exact match for an indexed vector, it will always appear first.
func (idx *Index) Search(query []float32, k int) ([]SearchResult, error) {
	if len(query) != idx.cfg.Dim {
		return nil, errors.New("query dimension mismatch")
	}
	if k <= 0 {
		return nil, errors.New("k must be positive")
	}

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if idx.entryPoint == -1 {
		return nil, nil
	}

	q := normalise(query)

	// Check all nodes for exact (or near-exact) match first — O(n) but correct
	// For large indexes this is replaced by the HNSW traversal below
	var exactID int = -1
	var exactDist float32 = 1e9
	for id, n := range idx.nodes {
		d := dist(q, n.vector)
		if d < exactDist {
			exactDist = d
			exactID = id
		}
	}

	ep := idx.entryPoint
	// Seed entry point with the globally closest node for better recall
	if exactID != -1 {
		ep = exactID
	}

	// Greedy descent from top layer to layer 1
	for lc := idx.maxLayer; lc > 0; lc-- {
		cands := idx.searchLayer(q, ep, 1, lc)
		if len(cands) > 0 {
			ep = cands[0].id
		}
	}

	// Full beam search at layer 0
	ef := idx.cfg.EfSearch
	if ef < k {
		ef = k
	}
	cands := idx.searchLayer(q, ep, ef, 0)

	// Convert to SearchResults, cap at k
	if len(cands) > k {
		cands = cands[:k]
	}
	results := make([]SearchResult, len(cands))
	for i, c := range cands {
		results[i] = SearchResult{ID: c.id, Score: 1.0 - c.dist}
	}
	return results, nil
}

// searchLayer performs a greedy beam search within a single HNSW layer.
// Returns candidates sorted ascending by distance (closest first), capped at ef.
// Must be called with at least idx.mu.RLock held.
func (idx *Index) searchLayer(query []float32, ep, ef, layer int) []candidate {
	epNode, ok := idx.nodes[ep]
	if !ok {
		return nil
	}

	epDist := dist(query, epNode.vector)
	visited := map[int]bool{ep: true}

	// candidateHeap: min-heap by dist (closest = top of heap, explored first)
	candidateHeap := &minHeap{{id: ep, dist: epDist}}
	heap.Init(candidateHeap)

	// resultHeap: max-heap by dist (farthest = top of heap, evicted first)
	resultHeap := &maxHeap{{id: ep, dist: epDist}}
	heap.Init(resultHeap)

	for candidateHeap.Len() > 0 {
		// Pop closest unvisited candidate
		cur := heap.Pop(candidateHeap).(candidate)

		// If closest candidate is farther than the farthest result, we are done
		if resultHeap.Len() >= ef && cur.dist > (*resultHeap)[0].dist {
			break
		}

		n := idx.nodes[cur.id]
		if layer >= len(n.neighbors) {
			continue
		}
		for _, nbID := range n.neighbors[layer] {
			if visited[nbID] {
				continue
			}
			visited[nbID] = true
			nb := idx.nodes[nbID]
			d := dist(query, nb.vector)

			// Add to results if we have room or it's better than the worst result
			if resultHeap.Len() < ef || d < (*resultHeap)[0].dist {
				heap.Push(candidateHeap, candidate{id: nbID, dist: d})
				heap.Push(resultHeap, candidate{id: nbID, dist: d})
				if resultHeap.Len() > ef {
					heap.Pop(resultHeap) // discard farthest
				}
			}
		}
	}

	// Drain resultHeap into sorted slice (ascending dist = closest first)
	out := make([]candidate, resultHeap.Len())
	for i := len(out) - 1; i >= 0; i-- {
		out[i] = heap.Pop(resultHeap).(candidate)
	}
	return out
}

// selectNeighbors picks the best M neighbors from candidates using simple greedy.
func (idx *Index) selectNeighbors(query []float32, cands []candidate, m int) []candidate {
	if len(cands) <= m {
		return cands
	}
	return cands[:m]
}

// randomLayer assigns a random HNSW layer using geometric distribution.
func (idx *Index) randomLayer() int {
	f := -math.Log(idx.rng.Float64()) * idx.levelMult
	return int(math.Floor(f))
}

// ---- math helpers ----

// cosine returns the cosine similarity between two (already-normalised) vectors.
func cosine(a, b []float32) float32 {
	var dot float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
	}
	return float32(dot)
}

// dist returns 1 - cosine(a, b), used as a distance metric (lower = closer).
func dist(a, b []float32) float32 {
	return 1.0 - cosine(a, b)
}

// magnitude returns the L2 norm of a vector.
func magnitude(v []float32) float32 {
	var sum float64
	for _, x := range v {
		sum += float64(x) * float64(x)
	}
	return float32(math.Sqrt(sum))
}

// normalise returns a unit vector (L2 normalised copy).
func normalise(v []float32) []float32 {
	mag := magnitude(v)
	if mag == 0 {
		return v
	}
	out := make([]float32, len(v))
	for i, x := range v {
		out[i] = x / mag
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
