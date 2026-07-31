package import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestEmpirical_VelocityLimiter_ExactThresholdBlocking tests exact blocking behavior for N thresholds.
func TestEmpirical_VelocityLimiter_ExactThresholdBlocking(t *testing.T) {
	thresholds := []int{1, 3, 5, 10}

	for _, maxAttempts := range thresholds {
		t.Run(fmt.Sprintf("MaxAttempts_%d", maxAttempts), func(t *testing.T) {
			vl := NewVelocityLimiter(1*time.Minute, maxAttempts, 5*time.Minute)
			key := fmt.Sprintf("test_ip_%d", maxAttempts)

			// Record failures 1 to maxAttempts - 1
			for i := 1; i < maxAttempts; i++ {
				vl.RecordFailure(key)
				if vl.IsBlocked(key) {
					t.Fatalf("key blocked early at attempt %d (max: %d)", i, maxAttempts)
				}
			}

			// Record maxAttempts-th failure
			vl.RecordFailure(key)
			if !vl.IsBlocked(key) {
				t.Fatalf("key NOT blocked after reaching maxAttempts (%d)", maxAttempts)
			}

			// Record extra failure while blocked
			vl.RecordFailure(key)
			if !vl.IsBlocked(key) {
				t.Fatalf("key unblocked after additional failure while blocked")
			}
		})
	}
}

// TestEmpirical_VelocityLimiter_ResetCorrectness tests clearing failure history and block state across keys.
func TestEmpirical_VelocityLimiter_ResetCorrectness(t *testing.T) {
	vl := NewVelocityLimiter(1*time.Minute, 3, 5*time.Minute)
	const totalKeys = 50

	for i := 0; i < totalKeys; i++ {
		key := fmt.Sprintf("ip_%d", i)
		for f := 0; f < 3; f++ {
			vl.RecordFailure(key)
		}
		if !vl.IsBlocked(key) {
			t.Fatalf("key %s should be blocked", key)
		}
	}

	// Reset half the keys
	for i := 0; i < totalKeys/2; i++ {
		key := fmt.Sprintf("ip_%d", i)
		vl.Reset(key)
	}

	// Verify reset keys are unblocked and counters cleared
	for i := 0; i < totalKeys/2; i++ {
		key := fmt.Sprintf("ip_%d", i)
		if vl.IsBlocked(key) {
			t.Errorf("reset key %s is still blocked", key)
		}
		// Record 1 failure - should not block
		vl.RecordFailure(key)
		if vl.IsBlocked(key) {
			t.Errorf("reset key %s blocked after only 1 failure post-reset", key)
		}
	}

	// Verify un-reset keys remain blocked
	for i := totalKeys / 2; i < totalKeys; i++ {
		key := fmt.Sprintf("ip_%d", i)
		if !vl.IsBlocked(key) {
			t.Errorf("non-reset key %s should remain blocked", key)
		}
	}
}

// TestEmpirical_VelocityLimiter_SlidingWindowExpiryAndPrecision tests precision of failure window expiration.
func TestEmpirical_VelocityLimiter_SlidingWindowExpiryAndPrecision(t *testing.T) {
	window := 80 * time.Millisecond
	maxAttempts := 3
	blockDuration := 500 * time.Millisecond

	vl := NewVelocityLimiter(window, maxAttempts, blockDuration)
	key := "window_test_key"

	// Record 2 failures
	vl.RecordFailure(key)
	vl.RecordFailure(key)
	if vl.IsBlocked(key) {
		t.Fatalf("blocked after 2 failures")
	}

	// Wait past window duration (100ms > 80ms)
	time.Sleep(100 * time.Millisecond)

	// Record 1 failure. The 2 prior failures should have expired.
	vl.RecordFailure(key)
	if vl.IsBlocked(key) {
		t.Fatalf("blocked because old failures were not evicted from sliding window")
	}

	// Record 2 more failures (total 3 in current window)
	vl.RecordFailure(key)
	vl.RecordFailure(key)
	if !vl.IsBlocked(key) {
		t.Fatalf("failed to block after 3 failures in fresh window")
	}
}

// TestEmpirical_VelocityLimiter_BlockDurationExpiry tests automatic unblocking after blockDuration.
func TestEmpirical_VelocityLimiter_BlockDurationExpiry(t *testing.T) {
	window := 1 * time.Second
	maxAttempts := 2
	blockDuration := 80 * time.Millisecond

	vl := NewVelocityLimiter(window, maxAttempts, blockDuration)
	key := "block_duration_key"

	vl.RecordFailure(key)
	vl.RecordFailure(key)

	if !vl.IsBlocked(key) {
		t.Fatalf("key should be blocked")
	}

	// Wait past blockDuration (110ms > 80ms)
	time.Sleep(110 * time.Millisecond)

	if vl.IsBlocked(key) {
		t.Fatalf("key should be unblocked after blockDuration expired")
	}

	// Recording a new failure should start new window with count 1
	vl.RecordFailure(key)
	if vl.IsBlocked(key) {
		t.Fatalf("key blocked immediately on 1st attempt after block expiration")
	}
}

// TestEmpirical_VelocityLimiter_HighConcurrency tests concurrent access across 100 goroutines.
func TestEmpirical_VelocityLimiter_HighConcurrency(t *testing.T) {
	vl := NewVelocityLimiter(500*time.Millisecond, 5, 200*time.Millisecond)
	const numGoroutines = 100
	const opsPerGoroutine = 300

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	var failuresRecorded int64
	var checksPerformed int64
	var resetsPerformed int64

	for i := 0; i < numGoroutines; i++ {
		go func(gID int) {
			defer wg.Done()
			key := fmt.Sprintf("ip_conc_%d", gID%10) // 10 shared keys across 100 goroutines

			for j := 0; j < opsPerGoroutine; j++ {
				vl.RecordFailure(key)
				atomic.AddInt64(&failuresRecorded, 1)

				_ = vl.IsBlocked(key)
				atomic.AddInt64(&checksPerformed, 1)

				if j%20 == 0 {
					vl.Reset(key)
					atomic.AddInt64(&resetsPerformed, 1)
				}
			}
		}(i)
	}

	wg.Wait()

	if failuresRecorded != int64(numGoroutines*opsPerGoroutine) {
		t.Errorf("expected %d failures recorded, got %d", numGoroutines*opsPerGoroutine, failuresRecorded)
	}
}
