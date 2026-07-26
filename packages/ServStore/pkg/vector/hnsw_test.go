package vector

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
)

// randVec produces a random float32 slice of length dim.
func randVec(dim int) []float32 {
	v := make([]float32, dim)
	for i := range v {
		v[i] = rand.Float32()*2 - 1 // [-1, 1]
	}
	return v
}

func TestHNSWCosineSimilarity(t *testing.T) {
	// Identical vectors should have cosine similarity 1.0
	v := []float32{1, 0, 0, 0}
	nv := normalise(v)
	sim := cosine(nv, nv)
	if sim < 0.999 {
		t.Errorf("identical vector cosine should be ~1.0, got %f", sim)
	}

	// Orthogonal vectors should have cosine similarity 0.0
	a := normalise([]float32{1, 0, 0, 0})
	b := normalise([]float32{0, 1, 0, 0})
	sim2 := cosine(a, b)
	if sim2 > 0.001 || sim2 < -0.001 {
		t.Errorf("orthogonal vectors cosine should be ~0.0, got %f", sim2)
	}
}

func TestHNSWInsertAndSearch(t *testing.T) {
	const dim = 8
	idx := NewIndex(Config{M: 16, EfConstruction: 100, EfSearch: 50, Dim: dim})

	// Insert a known target vector as ID 42
	target := []float32{0.9, 0.1, 0.05, 0.02, 0.01, 0.0, 0.0, 0.0}

	// Insert 100 random vectors; slot 42 is the target
	for i := 0; i < 100; i++ {
		var v []float32
		if i == 42 {
			v = make([]float32, dim)
			copy(v, target)
		} else {
			v = randVec(dim)
		}
		if err := idx.Insert(i, v); err != nil {
			t.Fatalf("Insert(%d) failed: %v", i, err)
		}
	}

	if idx.Len() != 100 {
		t.Errorf("expected 100 nodes, got %d", idx.Len())
	}

	// Search for the target — ID 42 must appear in top-5 with score near 1
	results, err := idx.Search(target, 5)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}

	found := false
	for _, r := range results {
		if r.ID == 42 && r.Score > 0.99 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected vector 42 in top-5 results with score>0.99, got: %+v", results)
	}
}

func TestHNSWDimensionMismatch(t *testing.T) {
	idx := NewIndex(DefaultConfig(4))
	err := idx.Insert(1, []float32{1, 2, 3}) // wrong dim
	if err == nil {
		t.Error("expected dimension mismatch error")
	}
}

func TestVectorStoreHTTP(t *testing.T) {
	cfg := DefaultConfig(4)
	vs := NewVectorStore(cfg)
	h := vs.Handler()

	// Upsert 5 vectors
	vecs := [][]float32{
		{1, 0, 0, 0},
		{0, 1, 0, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
		{1, 1, 0, 0},
	}
	for i, v := range vecs {
		body, _ := json.Marshal(map[string]interface{}{"id": i, "vector": v})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/vectors/testbucket", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("upsert id=%d: expected 201, got %d body=%s", i, w.Code, w.Body.String())
		}
	}

	// Search for vector closest to [1, 0, 0, 0] — should return ID 0
	searchBody, _ := json.Marshal(map[string]interface{}{
		"vector":    []float32{1, 0, 0, 0},
		"k":         3,
		"min_score": 0.5,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/vectors/testbucket/search", bytes.NewReader(searchBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("search: expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	results, _ := resp["results"].([]interface{})
	if len(results) == 0 {
		t.Error("expected at least 1 search result")
	}

	// Stats endpoint
	statsReq := httptest.NewRequest(http.MethodGet, "/api/v1/vectors/testbucket/stats", nil)
	sw := httptest.NewRecorder()
	h.ServeHTTP(sw, statsReq)
	if sw.Code != http.StatusOK {
		t.Fatalf("stats: expected 200, got %d", sw.Code)
	}
	var stats map[string]interface{}
	json.NewDecoder(sw.Body).Decode(&stats)
	if stats["bucket"] != "testbucket" {
		t.Errorf("stats bucket mismatch: %v", stats)
	}
}
