package import (
	"testing"
)

func TestAutoEmbeddingGenerator_GenerateEmbedding(t *testing.T) {
	generator := NewAutoEmbeddingGenerator(64)

	text := "Pranor Vault zero-dependency vector object storage"
	vec := generator.GenerateEmbedding(text)

	if len(vec) != 64 {
		t.Fatalf("expected vector dimension 64, got %d", len(vec))
	}

	// Verify L2 norm ~ 1.0
	var sum float64
	for _, v := range vec {
		sum += float64(v * v)
	}

	if sum < 0.99 || sum > 1.01 {
		t.Errorf("expected L2 normalized vector, got sum of squares %f", sum)
	}
}
