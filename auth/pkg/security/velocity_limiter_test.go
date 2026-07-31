package security

import (
	"sync"
	"testing"
	"time"
)

func TestVelocityLimiter_ThresholdBlocking(t *testing.T) {
	window := 1 * time.Second
	maxAttempts := 3
	blockDuration := 500 * time.Millisecond

	vl := NewVelocityLimiter(window, maxAttempts, blockDuration)
	key := "192.168.1.1"

	// 1st failure
	vl.RecordFailure(key)
	if vl.IsBlocked(key) {
		t.Errorf("expected not blocked after 1 failure")
	}

	// 2nd failure
	vl.RecordFailure(key)
	if vl.IsBlocked(key) {
		t.Errorf("expected not blocked after 2 failures")
	}

	// 3rd failure (threshold reached)
	vl.RecordFailure(key)
	if !vl.IsBlocked(key) {
		t.Errorf("expected blocked after 3 failures")
	}
}

func TestVelocityLimiter_Reset(t *testing.T) {
	vl := NewVelocityLimiter(1*time.Minute, 2, 10*time.Minute)
	key := "admin_user"

	vl.RecordFailure(key)
	vl.RecordFailure(key)

	if !vl.IsBlocked(key) {
		t.Fatalf("expected key to be blocked after 2 failures")
	}

	// Reset key
	vl.Reset(key)

	if vl.IsBlocked(key) {
		t.Errorf("expected key to be unblocked after Reset")
	}

	// Single new failure should not block
	vl.RecordFailure(key)
	if vl.IsBlocked(key) {
		t.Errorf("expected key not blocked after 1 failure post-reset")
	}
}

func TestVelocityLimiter_WindowExpiry(t *testing.T) {
	window := 50 * time.Millisecond
	maxAttempts := 2
	blockDuration := 500 * time.Millisecond

	vl := NewVelocityLimiter(window, maxAttempts, blockDuration)
	key := "10.0.0.1"

	// 1st failure
	vl.RecordFailure(key)
	if vl.IsBlocked(key) {
		t.Fatalf("expected not blocked after 1 failure")
	}

	// Wait for window to expire
	time.Sleep(70 * time.Millisecond)

	// 2nd failure (but 1st failure expired, so count in window is 1)
	vl.RecordFailure(key)
	if vl.IsBlocked(key) {
		t.Errorf("expected not blocked because 1st failure expired out of sliding window")
	}
}

func TestVelocityLimiter_BlockDurationExpiry(t *testing.T) {
	window := 1 * time.Second
	maxAttempts := 2
	blockDuration := 50 * time.Millisecond

	vl := NewVelocityLimiter(window, maxAttempts, blockDuration)
	key := "user@example.com"

	vl.RecordFailure(key)
	vl.RecordFailure(key)

	if !vl.IsBlocked(key) {
		t.Fatalf("expected blocked")
	}

	// Wait for block duration to expire
	time.Sleep(70 * time.Millisecond)

	if vl.IsBlocked(key) {
		t.Errorf("expected unblocked after block duration expired")
	}
}

func TestVelocityLimiter_IndependentKeys(t *testing.T) {
	vl := NewVelocityLimiter(1*time.Minute, 2, 5*time.Minute)
	key1 := "192.168.1.1"
	key2 := "192.168.1.2"

	vl.RecordFailure(key1)
	vl.RecordFailure(key1)

	if !vl.IsBlocked(key1) {
		t.Fatalf("expected key1 blocked")
	}
	if vl.IsBlocked(key2) {
		t.Errorf("expected key2 not blocked")
	}
}

func TestVelocityLimiter_DefaultsAndEmptyKeys(t *testing.T) {
	// Zero/negative params should default
	vl := NewVelocityLimiter(0, 0, 0)
	if vl.GetWindowDuration() != 1*time.Minute {
		t.Errorf("expected 1m default window, got %v", vl.GetWindowDuration())
	}
	if vl.GetMaxAttempts() != 5 {
		t.Errorf("expected 5 default max attempts, got %d", vl.GetMaxAttempts())
	}
	if vl.GetBlockDuration() != 15*time.Minute {
		t.Errorf("expected 15m default block duration, got %v", vl.GetBlockDuration())
	}

	// Empty keys should not panic or block
	vl.RecordFailure("")
	if vl.IsBlocked("") {
		t.Errorf("empty key should not be blocked")
	}
	vl.Reset("")
}

func TestVelocityLimiter_Concurrency(t *testing.T) {
	vl := NewVelocityLimiter(500*time.Millisecond, 10, 100*time.Millisecond)
	const numGoroutines = 20
	const numOps = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			key := "shared_key"
			for j := 0; j < numOps; j++ {
				vl.RecordFailure(key)
				_ = vl.IsBlocked(key)
				if j%10 == 0 {
					vl.Reset(key)
				}
			}
		}(i)
	}

	wg.Wait()
}
