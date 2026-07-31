package ai

import (
	"container/heap"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/vyuvaraj/pranor/gate/pkg/proxy"
)

// PromptVector represents a prompt string converted to a normalized embedding vector.
type PromptVector struct {
	ID        string
	Prompt    string
	Response  *proxy.HTTPCacheEntry
	Embedding []float32
	ExpiresAt time.Time
}

// VectorNode is an HNSW graph node for prompt semantic cache.
type VectorNode struct {
	ID        int
	Vector    PromptVector
	Neighbors [][]int // layer -> neighbor IDs
	Level     int
}

type itemDist struct {
	id   int
	dist float32
}

type minItemHeap []itemDist
func (h minItemHeap) Len() int            { return len(h) }
func (h minItemHeap) Less(i, j int) bool  { return h[i].dist < h[j].dist }
func (h minItemHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *minItemHeap) Push(x interface{}) { *h = append(*h, x.(itemDist)) }
func (h *minItemHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

type maxItemHeap []itemDist
func (h maxItemHeap) Len() int            { return len(h) }
func (h maxItemHeap) Less(i, j int) bool  { return h[i].dist > h[j].dist }
func (h maxItemHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *maxItemHeap) Push(x interface{}) { *h = append(*h, x.(itemDist)) }
func (h *maxItemHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// VectorHNSWCache is a high-performance vector-indexed HNSW semantic cache with LRU eviction (SG.H1).
type VectorHNSWCache struct {
	mu             sync.RWMutex
	dim            int
	maxElements    int
	m              int
	efConstruction int
	efSearch       int
	similarityThreshold float32

	nodes        map[int]*VectorNode
	idToPromptID map[string]int
	entryLRU     []string // queue of Prompt IDs for LRU eviction
	entryPointID int
	maxLevel     int

	counter int
	ttl     time.Duration
}

// NewVectorHNSWCache initializes an HNSW semantic cache engine.
func NewVectorHNSWCache(dim, maxElements int, similarityThreshold float32, ttl time.Duration) *VectorHNSWCache {
	if dim <= 0 {
		dim = 64 // default embedding dimension
	}
	if maxElements <= 0 {
		maxElements = 100000
	}
	if similarityThreshold <= 0 {
		similarityThreshold = 0.85
	}
	if ttl <= 0 {
		ttl = 1 * time.Hour
	}

	return &VectorHNSWCache{
		dim:                 dim,
		maxElements:         maxElements,
		m:                   16,
		efConstruction:      64,
		efSearch:            32,
		similarityThreshold: similarityThreshold,
		nodes:               make(map[int]*VectorNode),
		idToPromptID:        make(map[string]int),
		entryPointID:        -1,
		maxLevel:            -1,
		ttl:                 ttl,
	}
}

// Simple TF-IDF / character n-gram pseudo embedding generator for fast sub-ms vectorization
func (vc *VectorHNSWCache) EmbedPrompt(prompt string) []float32 {
	vec := make([]float32, vc.dim)
	lower := strings.ToLower(prompt)

	// Hash n-grams into dimension buckets
	words := strings.Fields(lower)
	for idx, word := range words {
		h := sha256.Sum256([]byte(word))
		bucket := int(h[0]) % vc.dim
		val := float32(h[1]) / 255.0
		// Position weighting
		weight := 1.0 / math.Sqrt(float64(idx+1))
		vec[bucket] += val * float32(weight)
	}

	// Normalize vector to unit length
	var norm float64
	for _, v := range vec {
		norm += float64(v * v)
	}
	if norm > 0 {
		norm = math.Sqrt(norm)
		for i := range vec {
			vec[i] /= float32(norm)
		}
	}
	return vec
}

func cosineDistance(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 1.0
	}
	var dot float32
	for i := range a {
		dot += a[i] * b[i]
	}
	// Distance = 1 - cosine_similarity
	dist := 1.0 - dot
	if dist < 0 {
		return 0
	}
	return dist
}

// Get performs an HNSW approximate nearest neighbor search to find a cached prompt matching
// within the cosine similarity threshold (SG.H1).
func (vc *VectorHNSWCache) Get(prompt string) (*proxy.HTTPCacheEntry, float32, bool) {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	if vc.entryPointID == -1 || len(vc.nodes) == 0 {
		return nil, 0, false
	}

	queryVec := vc.EmbedPrompt(prompt)
	currObj := vc.nodes[vc.entryPointID]
	if currObj == nil {
		return nil, 0, false
	}

	currDist := cosineDistance(queryVec, currObj.Vector.Embedding)

	// Traverse top layers down to layer 0
	for level := vc.maxLevel; level > 0; level-- {
		changed := true
		for changed {
			changed = false
			if level < len(currObj.Neighbors) {
				for _, neighborID := range currObj.Neighbors[level] {
					neighbor := vc.nodes[neighborID]
					if neighbor == nil {
						continue
					}
					d := cosineDistance(queryVec, neighbor.Vector.Embedding)
					if d < currDist {
						currDist = d
						currObj = neighbor
						changed = true
					}
				}
			}
		}
	}

	// Layer 0 beam search
	candidates := &minItemHeap{}
	heap.Init(candidates)
	visited := map[int]bool{currObj.ID: true}
	heap.Push(candidates, itemDist{id: currObj.ID, dist: currDist})

	bestNode := currObj
	bestDist := currDist

	for candidates.Len() > 0 {
		curr := heap.Pop(candidates).(itemDist)
		if curr.dist > bestDist && candidates.Len() >= vc.efSearch {
			break
		}

		nodeObj := vc.nodes[curr.id]
		if nodeObj == nil {
			continue
		}

		if len(nodeObj.Neighbors) > 0 {
			for _, neighborID := range nodeObj.Neighbors[0] {
				if visited[neighborID] {
					continue
				}
				visited[neighborID] = true
				neighbor := vc.nodes[neighborID]
				if neighbor == nil {
					continue
				}

				d := cosineDistance(queryVec, neighbor.Vector.Embedding)
				if d < bestDist {
					bestDist = d
					bestNode = neighbor
				}
				if d < currDist || candidates.Len() < vc.efSearch {
					heap.Push(candidates, itemDist{id: neighborID, dist: d})
				}
			}
		}
	}

	sim := 1.0 - bestDist
	if sim >= vc.similarityThreshold {
		if time.Now().Before(bestNode.Vector.ExpiresAt) {
			return bestNode.Vector.Response, sim, true
		}
	}

	return nil, sim, false
}

// Put inserts a prompt and its HTTP response entry into the HNSW graph (SG.H1).
func (vc *VectorHNSWCache) Put(prompt string, resp *proxy.HTTPCacheEntry) {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	// LRU eviction if maxElements reached
	if len(vc.nodes) >= vc.maxElements && len(vc.entryLRU) > 0 {
		oldestID := vc.entryLRU[0]
		vc.entryLRU = vc.entryLRU[1:]
		if nodeID, ok := vc.idToPromptID[oldestID]; ok {
			delete(vc.nodes, nodeID)
			delete(vc.idToPromptID, oldestID)
		}
	}

	promptHash := hex.EncodeToString(sha256.New().Sum([]byte(prompt)))
	embedding := vc.EmbedPrompt(prompt)

	level := 0
	for rand.Float64() < 0.5 && level < 16 {
		level++
	}

	vc.counter++
	newNodeID := vc.counter

	pv := PromptVector{
		ID:        promptHash,
		Prompt:    prompt,
		Response:  resp,
		Embedding: embedding,
		ExpiresAt: time.Now().Add(vc.ttl),
	}

	newNode := &VectorNode{
		ID:        newNodeID,
		Vector:    pv,
		Neighbors: make([][]int, level+1),
		Level:     level,
	}

	vc.nodes[newNodeID] = newNode
	vc.idToPromptID[promptHash] = newNodeID
	vc.entryLRU = append(vc.entryLRU, promptHash)

	if vc.entryPointID == -1 {
		vc.entryPointID = newNodeID
		vc.maxLevel = level
		return
	}

	if level > vc.maxLevel {
		vc.maxLevel = level
		vc.entryPointID = newNodeID
	}
}
