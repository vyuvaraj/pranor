package core

import (
	"testing"
)

func TestSIMDVectorFilterEngine(t *testing.T) {
	simd := NewSIMDVectorFilterEngine("URGENT")

	payloads := []string{
		`{"id": 1, "msg": "URGENT order cancel"}`,
		`{"id": 2, "msg": "Normal info"}`,
		`{"id": 3, "msg": "URGENT system alert"}`,
		`{"id": 4, "msg": "Routine batch process"}`,
	}

	results := simd.BatchMatch(payloads)
	if len(results) != 4 {
		t.Fatalf("Expected 4 match results, got %d", len(results))
	}

	if !results[0] || results[1] || !results[2] || results[3] {
		t.Errorf("Unexpected SIMD match results: %v", results)
	}
}
