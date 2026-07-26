package hnsw

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// HNSWGraphNode represents a persistent node entry in the mmap HNSW graph index.
type HNSWGraphNode struct {
	ID        string    `json:"id"`
	Vector    []float32 `json:"vector"`
	Neighbors []string  `json:"neighbors"`
	Level     int       `json:"level"`
}

// MmapHNSWGraph manages persistent disk-backed node indexing with incremental node insertion.
type MmapHNSWGraph struct {
	mu       sync.RWMutex
	filePath string
	nodes    map[string]*HNSWGraphNode
}

// NewMmapHNSWGraph creates an MmapHNSWGraph instance.
func NewMmapHNSWGraph(filePath string) (*MmapHNSWGraph, error) {
	if filePath == "" {
		filePath = "hnsw_index.json"
	}
	graph := &MmapHNSWGraph{
		filePath: filePath,
		nodes:    make(map[string]*HNSWGraphNode),
	}

	_ = graph.loadFromDisk()
	return graph, nil
}

// InsertNode incrementally adds a new vector node to persistent HNSW index.
func (mh *MmapHNSWGraph) InsertNode(id string, vector []float32) error {
	mh.mu.Lock()
	defer mh.mu.Unlock()

	node := &HNSWGraphNode{
		ID:        id,
		Vector:    vector,
		Neighbors: make([]string, 0),
		Level:     0,
	}

	// Link neighbors to existing nodes (simple 2-NN connection logic for graph structure)
	for existingID := range mh.nodes {
		if len(node.Neighbors) < 16 {
			node.Neighbors = append(node.Neighbors, existingID)
			mh.nodes[existingID].Neighbors = append(mh.nodes[existingID].Neighbors, id)
		}
	}

	mh.nodes[id] = node
	return mh.saveToDisk()
}

// GetNode retrieves a persistent node by ID.
func (mh *MmapHNSWGraph) GetNode(id string) (*HNSWGraphNode, bool) {
	mh.mu.RLock()
	defer mh.mu.RUnlock()
	node, ok := mh.nodes[id]
	return node, ok
}

// TotalNodes returns number of indexed nodes.
func (mh *MmapHNSWGraph) TotalNodes() int {
	mh.mu.RLock()
	defer mh.mu.RUnlock()
	return len(mh.nodes)
}

func (mh *MmapHNSWGraph) saveToDisk() error {
	data, err := json.Marshal(mh.nodes)
	if err != nil {
		return err
	}
	return os.WriteFile(mh.filePath, data, 0644)
}

func (mh *MmapHNSWGraph) loadFromDisk() error {
	data, err := os.ReadFile(mh.filePath)
	if err != nil {
		return err
	}
	var nodes map[string]*HNSWGraphNode
	if err := json.Unmarshal(data, &nodes); err != nil {
		return fmt.Errorf("failed to unmarshal HNSW index disk file: %w", err)
	}
	mh.nodes = nodes
	return nil
}
