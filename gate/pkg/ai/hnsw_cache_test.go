package import (
	"net/http"
	"testing"
	"time"

	"github.com/vyuvaraj/pranor/gate/pkg/proxy"
)

func TestVectorHNSWCache(t *testing.T) {
	cache := NewVectorHNSWCache(16, 100, 0.80, 10*time.Minute)

	prompt1 := "What is the capital of France?"
	entry1 := &proxy.HTTPCacheEntry{
		Body:       []byte(`{"reply":"Paris"}`),
		StatusCode: http.StatusOK,
		Headers:    make(http.Header),
		ExpiresAt:  time.Now().Add(10 * time.Minute),
	}

	cache.Put(prompt1, entry1)

	// Exact query hit
	res, sim, ok := cache.Get("What is the capital of France?")
	if !ok || res == nil || sim < 0.80 {
		t.Fatalf("Expected exact hit in vector cache, got ok=%v, sim=%.2f", ok, sim)
	}

	// Semantically similar query hit
	similarPrompt := "What is the capital city of France?"
	resSimilar, simSimilar, okSimilar := cache.Get(similarPrompt)
	if !okSimilar || resSimilar == nil || simSimilar < 0.70 {
		t.Fatalf("Expected similar query hit in HNSW vector cache, got ok=%v, sim=%.2f", okSimilar, simSimilar)
	}

	// Completely unrelated prompt miss
	unrelatedPrompt := "How to bake a chocolate chip cookie at home?"
	_, simUnrelated, okUnrelated := cache.Get(unrelatedPrompt)
	if okUnrelated || simUnrelated > 0.80 {
		t.Fatalf("Expected miss for unrelated prompt, got ok=%v, sim=%.2f", okUnrelated, simUnrelated)
	}
}
