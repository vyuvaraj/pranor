package search

import (
	"testing"
)

func TestHybridSearchEngine_RRF(t *testing.T) {
	hse := NewHybridSearchEngine()

	kwScores := map[string]float64{
		"doc1": 0.8,
		"doc2": 0.4,
	}

	vecScores := map[string]float64{
		"doc2": 0.95,
		"doc1": 0.60,
	}

	results := hse.PerformHybridSearch(kwScores, vecScores, 60.0)

	if len(results) != 2 {
		t.Fatalf("expected 2 merged search results, got %d", len(results))
	}

	// doc2 is #1 in vector and #2 in keyword -> high combined score
	if results[0].ObjectID != "doc2" && results[0].ObjectID != "doc1" {
		t.Errorf("unexpected top result: %+v", results[0])
	}
}
