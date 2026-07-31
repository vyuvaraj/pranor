package import (
	"hash/fnv"
	"math"
	"sync"
)

// Bloom represents a thread-safe probabilistic Bloom filter.
type Bloom struct {
	mu       sync.RWMutex
	bitset   []uint64
	m        uint64  // total number of bits
	k        uint64  // number of hash functions
	capacity int     // expected capacity
	fpRate   float64 // target false positive rate
}

// NewBloom creates a new Bloom filter optimized for the given capacity and false positive rate.
func NewBloom(capacity int, falsePositiveRate float64) *Bloom {
	if capacity <= 0 {
		capacity = 1000
	}
	if falsePositiveRate <= 0 || falsePositiveRate >= 1.0 {
		falsePositiveRate = 0.01
	}

	// m = ceil(-capacity * ln(p) / (ln(2)^2))
	mFloat := math.Ceil(-float64(capacity) * math.Log(falsePositiveRate) / (math.Ln2 * math.Ln2))
	m := uint64(mFloat)
	if m < 1 {
		m = 1
	}

	// k = ceil((m / capacity) * ln(2))
	kFloat := math.Ceil((float64(m) / float64(capacity)) * math.Ln2)
	k := uint64(kFloat)
	if k < 1 {
		k = 1
	}

	numWords := (m + 63) / 64

	return &Bloom{
		bitset:   make([]uint64, numWords),
		m:        m,
		k:        k,
		capacity: capacity,
		fpRate:   falsePositiveRate,
	}
}

// hashFNV computes two 64-bit FNV-1a hashes for double hashing (Kirsch-Mitzenmacher optimization).
func hashFNV(key string) (uint64, uint64) {
	data := []byte(key)

	h1 := fnv.New64a()
	h1.Write(data)
	v1 := h1.Sum64()

	h2 := fnv.New64a()
	h2.Write([]byte{0x01})
	h2.Write(data)
	v2 := h2.Sum64()
	if v2 == 0 {
		v2 = 1
	}

	return v1, v2
}

// getIndices returns the k bit indices for a given key.
func (b *Bloom) getIndices(key string) []uint64 {
	h1, h2 := hashFNV(key)
	indices := make([]uint64, b.k)
	for i := uint64(0); i < b.k; i++ {
		indices[i] = (h1 + i*h2) % b.m
	}
	return indices
}

// Add inserts a key into the Bloom filter.
func (b *Bloom) Add(key string) {
	indices := b.getIndices(key)

	b.mu.Lock()
	defer b.mu.Unlock()

	for _, idx := range indices {
		wordIdx := idx / 64
		bitIdx := idx % 64
		b.bitset[wordIdx] |= (uint64(1) << bitIdx)
	}
}

// MayContain checks if a key might be in the Bloom filter.
// Returns true if key might be present (with possible false positive), or false if definitely absent (zero false negatives).
func (b *Bloom) MayContain(key string) bool {
	indices := b.getIndices(key)

	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, idx := range indices {
		wordIdx := idx / 64
		bitIdx := idx % 64
		if (b.bitset[wordIdx] & (uint64(1) << bitIdx)) == 0 {
			return false
		}
	}
	return true
}

// Capacity returns the expected capacity of the Bloom filter.
func (b *Bloom) Capacity() int {
	return b.capacity
}

// FalsePositiveRate returns the configured target false positive rate.
func (b *Bloom) FalsePositiveRate() float64 {
	return b.fpRate
}

// M returns the total number of bits in the bit array.
func (b *Bloom) M() uint64 {
	return b.m
}

// K returns the number of hash functions used.
func (b *Bloom) K() uint64 {
	return b.k
}
