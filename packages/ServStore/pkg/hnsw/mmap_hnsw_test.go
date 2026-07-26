package hnsw

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMmapHNSWGraph_InsertAndReload(t *testing.T) {
	tempDir := t.TempDir()
	indexPath := filepath.Join(tempDir, "test_hnsw.json")

	graph, err := NewMmapHNSWGraph(indexPath)
	if err != nil {
		t.Fatalf("NewMmapHNSWGraph failed: %v", err)
	}

	// Insert nodes
	_ = graph.InsertNode("node-1", []float32{1.0, 0.0, 0.0})
	_ = graph.InsertNode("node-2", []float32{0.0, 1.0, 0.0})

	if graph.TotalNodes() != 2 {
		t.Fatalf("expected 2 total nodes, got %d", graph.TotalNodes())
	}

	// Reload from disk
	reloaded, err := NewMmapHNSWGraph(indexPath)
	if err != nil {
		t.Fatalf("failed to reload graph from disk: %v", err)
	}

	if reloaded.TotalNodes() != 2 {
		t.Errorf("expected 2 reloaded nodes, got %d", reloaded.TotalNodes())
	}

	node1, found := reloaded.GetNode("node-1")
	if !found || len(node1.Neighbors) != 1 || node1.Neighbors[0] != "node-2" {
		t.Errorf("unexpected reloaded node1 data: %+v", node1)
	}

	_ = os.Remove(indexPath)
}
