package bloom

import (
	"fmt"
	"sync"
	"testing"

	)

func TestBloom_ZeroFalseNegatives(t *testing.T) {
	capacity := 1000
	fpRate := 0.01
	b := NewBloom(capacity, fpRate)

	// Add 1000 items
	for i := 0; i < capacity; i++ {
		key := fmt.Sprintf("added_key_%d", i)
		b.Add(key)
	}

	// Verify all 1000 added items return MayContain == true
	for i := 0; i < capacity; i++ {
		key := fmt.Sprintf("added_key_%d", i)
		if !b.MayContain(key) {
			t.Fatalf("expected MayContain(%s) to be true, got false (false negative)", key)
		}
	}
}

func TestBloom_FalsePositiveRate(t *testing.T) {
	capacity := 1000
	fpRate := 0.05
	b := NewBloom(capacity, fpRate)

	// Add 1000 items
	for i := 0; i < capacity; i++ {
		b.Add(fmt.Sprintf("inserted_item_%d", i))
	}

	// Query 10,000 un-added items
	numQueries := 10000
	falsePositives := 0
	for i := 0; i < numQueries; i++ {
		key := fmt.Sprintf("unadded_item_%d", i)
		if b.MayContain(key) {
			falsePositives++
		}
	}

	observedFpRate := float64(falsePositives) / float64(numQueries)
	t.Logf("Observed FP rate: %f (%d/%d), target: %f", observedFpRate, falsePositives, numQueries, fpRate)

	// Allow a small tolerance over theoretical FP rate due to hash variance
	maxAllowedFpRate := fpRate + 0.02
	if observedFpRate > maxAllowedFpRate {
		t.Fatalf("false positive rate %f exceeded allowed threshold %f", observedFpRate, maxAllowedFpRate)
	}
}

func TestBloom_EdgeAndInvalidParameters(t *testing.T) {
	// Negative / zero capacity and out-of-range fpRate should fallback safely
	b := NewBloom(0, -0.5)
	if b.Capacity() <= 0 {
		t.Errorf("expected positive fallback capacity, got %d", b.Capacity())
	}
	if b.FalsePositiveRate() <= 0 || b.FalsePositiveRate() >= 1.0 {
		t.Errorf("expected fallback fpRate in (0, 1), got %f", b.FalsePositiveRate())
	}
	if b.M() < 1 || b.K() < 1 {
		t.Errorf("expected m >= 1 and k >= 1, got m=%d, k=%d", b.M(), b.K())
	}

	// Test empty string key and long key
	b.Add("")
	if !b.MayContain("") {
		t.Errorf("expected MayContain(\"\") to be true after Add")
	}

	longKey := "a_very_long_key_with_many_characters_and_special_symbols_!@#$%^&*()_+"
	if b.MayContain(longKey) {
		t.Errorf("expected MayContain to be false for non-inserted long key")
	}
	b.Add(longKey)
	if !b.MayContain(longKey) {
		t.Errorf("expected MayContain to be true after adding long key")
	}
}

func TestBloom_Concurrency(t *testing.T) {
	b := NewBloom(1000, 0.01)
	var wg sync.WaitGroup

	numWriters := 10
	numReaders := 10
	itemsPerWorker := 100

	// Concurrent Adders
	for w := 0; w < numWriters; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < itemsPerWorker; i++ {
				b.Add(fmt.Sprintf("worker_%d_item_%d", workerID, i))
			}
		}(w)
	}

	// Concurrent Readers
	for r := 0; r < numReaders; r++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < itemsPerWorker; i++ {
				_ = b.MayContain(fmt.Sprintf("worker_%d_item_%d", workerID, i))
			}
		}(r)
	}

	wg.Wait()

	// Verify all items written by writers are present
	for w := 0; w < numWriters; w++ {
		for i := 0; i < itemsPerWorker; i++ {
			key := fmt.Sprintf("worker_%d_item_%d", w, i)
			if !b.MayContain(key) {
				t.Fatalf("concurrently added key %s missing from Bloom filter", key)
			}
		}
	}
}
