package sessions

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestEmpirical_TokenEntropyAndUniqueness generates 10,000 tokens to test
// cryptographic randomness, token format (64 hex chars), 0 collisions, and bit distribution.
func TestEmpirical_TokenEntropyAndUniqueness(t *testing.T) {
	ts := NewTokenStore()
	const totalTokens = 10000
	tokens := make(chan string, totalTokens)

	const numWorkers = 20
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < totalTokens/numWorkers; j++ {
				userID := fmt.Sprintf("user-%d-%d", workerID, j)
				tok, err := ts.Issue(userID)
				if err != nil {
					t.Errorf("worker %d failed to issue token: %v", workerID, err)
					return
				}
				tokens <- tok
			}
		}(i)
	}

	wg.Wait()
	close(tokens)

	seen := make(map[string]bool, totalTokens)
	charCounts := make(map[rune]int)
	var totalChars int64

	for tok := range tokens {
		// 1. Length check: 32 bytes hex = 64 characters
		if len(tok) != 64 {
			t.Fatalf("expected token length 64, got %d for token %s", len(tok), tok)
		}

		// 2. Hex decode check
		decoded, err := hex.DecodeString(tok)
		if err != nil {
			t.Fatalf("failed to decode hex token %s: %v", tok, err)
		}
		if len(decoded) != 32 {
			t.Fatalf("expected 32 decoded bytes, got %d", len(decoded))
		}

		// 3. Uniqueness check
		if seen[tok] {
			t.Fatalf("DUPLICATE TOKEN DETECTED: %s", tok)
		}
		seen[tok] = true

		// 4. Character distribution check
		for _, ch := range tok {
			charCounts[ch]++
			totalChars++
		}
	}

	if len(seen) != totalTokens {
		t.Fatalf("expected %d unique tokens, got %d", totalTokens, len(seen))
	}

	// 5. Entropy & uniformity distribution test for 16 hex chars (0-9, a-f)
	expectedPerChar := float64(totalChars) / 16.0
	hexChars := "0123456789abcdef"
	for _, ch := range hexChars {
		count := charCounts[ch]
		diffRatio := math.Abs(float64(count)-expectedPerChar) / expectedPerChar
		if diffRatio > 0.15 { // Max 15% deviation allowed for 640k sample size
			t.Errorf("char %c count %d deviates by %.2f%% from expected %.0f", ch, count, diffRatio*100, expectedPerChar)
		}
	}
}

// TestEmpirical_TokenStore_HighConcurrency tests concurrent operations across 100 goroutines.
func TestEmpirical_TokenStore_HighConcurrency(t *testing.T) {
	ts := NewTokenStore(2 * time.Second)
	const numGoroutines = 100
	const opsPerGoroutine = 200

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	var issuedCount int64
	var validatedCount int64
	var revokedCount int64

	for i := 0; i < numGoroutines; i++ {
		go func(gID int) {
			defer wg.Done()
			var localTokens []string
			for j := 0; j < opsPerGoroutine; j++ {
				userID := fmt.Sprintf("user-conc-%d-%d", gID, j)
				tok, err := ts.Issue(userID)
				if err != nil {
					t.Errorf("Issue failed: %v", err)
					return
				}
				atomic.AddInt64(&issuedCount, 1)
				localTokens = append(localTokens, tok)

				// Validate previous token if available
				if len(localTokens) > 0 {
					targetTok := localTokens[j%len(localTokens)]
					uid, valErr := ts.Validate(targetTok)
					if valErr == nil && uid != "" {
						atomic.AddInt64(&validatedCount, 1)
					}
				}

				// Periodically revoke
				if j%3 == 0 && len(localTokens) > 0 {
					targetTok := localTokens[j%len(localTokens)]
					_ = ts.Revoke(targetTok)
					atomic.AddInt64(&revokedCount, 1)
				}

				// Periodically clean expired
				if j%50 == 0 {
					_ = ts.CleanExpired()
				}
			}
		}(i)
	}

	wg.Wait()

	if issuedCount != int64(numGoroutines*opsPerGoroutine) {
		t.Errorf("expected %d total issued, got %d", numGoroutines*opsPerGoroutine, issuedCount)
	}
}

// TestEmpirical_TokenStore_RevocationCorrectness verifies revocation isolation and error behavior.
func TestEmpirical_TokenStore_RevocationCorrectness(t *testing.T) {
	ts := NewTokenStore()
	const total = 1000
	activeTokens := make([]string, total/2)
	revokedTokens := make([]string, total/2)

	for i := 0; i < total/2; i++ {
		tActive, err := ts.Issue(fmt.Sprintf("user-active-%d", i))
		if err != nil {
			t.Fatalf("failed to issue active token: %v", err)
		}
		activeTokens[i] = tActive

		tRevoked, err := ts.Issue(fmt.Sprintf("user-revoked-%d", i))
		if err != nil {
			t.Fatalf("failed to issue revoked token: %v", err)
		}
		revokedTokens[i] = tRevoked
	}

	// Revoke half
	for _, tok := range revokedTokens {
		if err := ts.Revoke(tok); err != nil {
			t.Fatalf("failed to revoke token: %v", err)
		}
	}

	// Validate active tokens succeed
	for i, tok := range activeTokens {
		uid, err := ts.Validate(tok)
		if err != nil {
			t.Errorf("expected valid active token, got err: %v", err)
		}
		expectedUID := fmt.Sprintf("user-active-%d", i)
		if uid != expectedUID {
			t.Errorf("expected uid %s, got %s", expectedUID, uid)
		}
	}

	// Validate revoked tokens fail with ErrTokenRevoked
	for _, tok := range revokedTokens {
		_, err := ts.Validate(tok)
		if !errors.Is(err, ErrTokenRevoked) {
			t.Errorf("expected ErrTokenRevoked, got: %v", err)
		}
	}
}

// TestEmpirical_TokenStore_TTLExpiryRapidAccess tests rapid validate calls during TTL expiry transition.
func TestEmpirical_TokenStore_TTLExpiryRapidAccess(t *testing.T) {
	ttl := 60 * time.Millisecond
	ts := NewTokenStore(ttl)

	const tokenCount = 50
	tokens := make([]string, tokenCount)
	for i := 0; i < tokenCount; i++ {
		tok, err := ts.Issue(fmt.Sprintf("user-ttl-%d", i))
		if err != nil {
			t.Fatalf("issue failed: %v", err)
		}
		tokens[i] = tok
	}

	const numReaders = 20
	var wg sync.WaitGroup
	stopSignal := make(chan struct{})

	var validReads int64
	var expiredReads int64

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(rID int) {
			defer wg.Done()
			for {
				select {
				case <-stopSignal:
					return
				default:
					tok := tokens[rID%tokenCount]
					_, err := ts.Validate(tok)
					if err == nil {
						atomic.AddInt64(&validReads, 1)
					} else if errors.Is(err, ErrTokenExpired) || errors.Is(err, ErrTokenNotFound) {
						atomic.AddInt64(&expiredReads, 1)
					}
					time.Sleep(1 * time.Millisecond)
				}
			}
		}(i)
	}

	// Sleep past TTL
	time.Sleep(100 * time.Millisecond)
	close(stopSignal)
	wg.Wait()

	if validReads == 0 {
		t.Errorf("expected some valid reads before TTL, got 0")
	}
	if expiredReads == 0 {
		t.Errorf("expected some expired/not found reads after TTL, got 0")
	}

	// Final validation on all tokens must return ErrTokenExpired or ErrTokenNotFound
	for _, tok := range tokens {
		_, err := ts.Validate(tok)
		if !errors.Is(err, ErrTokenExpired) && !errors.Is(err, ErrTokenNotFound) {
			t.Errorf("expected expired or not found post-TTL, got: %v", err)
		}
	}
}
