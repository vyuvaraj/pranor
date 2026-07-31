package import (
	"crypto/sha256"
	"math"
	"strings"
	"sync"
)

// AutoEmbeddingGenerator generates dense vector embeddings for text objects upon PUT operations.
type AutoEmbeddingGenerator struct {
	mu        sync.RWMutex
	dimension int
}

// NewAutoEmbeddingGenerator creates an AutoEmbeddingGenerator instance.
func NewAutoEmbeddingGenerator(dimension int) *AutoEmbeddingGenerator {
	if dimension <= 0 {
		dimension = 128
	}
	return &AutoEmbeddingGenerator{dimension: dimension}
}

// GenerateEmbedding creates a deterministic normalized dense vector for text input.
func (aeg *AutoEmbeddingGenerator) GenerateEmbedding(text string) []float32 {
	if text == "" {
		return make([]float32, aeg.dimension)
	}

	vec := make([]float32, aeg.dimension)
	words := strings.Fields(text)

	for i, word := range words {
		hash := sha256.Sum256([]byte(word))
		for d := 0; d < aeg.dimension; d++ {
			val := float32(hash[d%32]) / 255.0
			vec[d] += val * float32(i+1)
		}
	}

	// L2 Normalize
	var norm float64
	for _, v := range vec {
		norm += float64(v * v)
	}
	norm = math.Sqrt(norm)
	if norm > 0 {
		for i := range vec {
			vec[i] = float32(float6